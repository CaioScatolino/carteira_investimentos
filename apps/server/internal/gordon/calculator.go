package gordon

import (
	"carteira_investimentos/server/internal/domain"
)

// AnalisadorGordon implementa o Modelo de Desconto de Dividendos de Gordon (DDM)
// Avalia o Preço Justo de uma ação com base no fluxo futuro projetado de proventos:
// Preço Teto = D1 / (k - g)
// Onde:
// - D1: Próximo dividendo estimado (Dividendos12M * (1 + g))
// - k: Taxa de retorno exigida pelo investidor (custo de oportunidade, ex: 11% a.a.)
// - g: Taxa de crescimento perpétuo sustentável dos dividendos (ex: 3% a 4.5% a.a.)
type AnalisadorGordon struct {
	taxaDesconto float64 // Custo de capital próprio / taxa de desconto (ex: 0.11 = 11% a.a.)
}

func Novo(taxaDesconto float64) *AnalisadorGordon {
	if taxaDesconto <= 0 {
		taxaDesconto = 0.11 // Padrão conservador de 11% a.a. para o Brasil
	}
	return &AnalisadorGordon{
		taxaDesconto: taxaDesconto,
	}
}

func (g *AnalisadorGordon) Nome() string {
	return "Gordon Growth Model (DDM)"
}

func (g *AnalisadorGordon) Executar(ativo *domain.Ativo) error {
	// Aplicável a ações com proventos distribuídos e lucro positivo
	if ativo.Classe != domain.ClasseAcao || ativo.Dividendos12M <= 0 || ativo.LPA <= 0 {
		ativo.PrecoTetoGordon = 0
		return nil
	}

	// Se provento for atípico ou payout > 100%, usa a base sustentável (60% do LPA)
	baseDividendo := ativo.Dividendos12M
	if ativo.IsProventoAtipico || ativo.Payout > 100.0 {
		divSustentavel := ativo.LPA * 0.60
		if divSustentavel < baseDividendo {
			baseDividendo = divSustentavel
		}
	}

	// Taxa de crescimento sustentável g:
	// Estima g com base na rentabilidade sobre o patrimônio (ROE * taxa de retenção)
	crescimento := 0.035 // Padrão conservador de 3.5% a.a. (em linha com PIB/inflação longa)

	if ativo.VPA > 0 && ativo.LPA > 0 {
		roe := ativo.LPA / ativo.VPA
		// Considera retenção média típica de 40% (payout médio de 60%)
		crescCalculado := roe * 0.40
		if crescCalculado >= 0.02 && crescCalculado <= 0.045 {
			crescimento = crescCalculado
		} else if crescCalculado > 0.045 {
			crescimento = 0.045 // Trava conservadora em 4.5% a.a.
		}
	}

	// Evita denominador nulo ou negativo caso a taxa de desconto seja menor ou igual ao crescimento
	if g.taxaDesconto <= crescimento {
		return nil
	}

	d1 := baseDividendo * (1.0 + crescimento)
	ativo.PrecoTetoGordon = d1 / (g.taxaDesconto - crescimento)

	return nil
}
