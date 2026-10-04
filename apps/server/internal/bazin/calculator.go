package bazin

import (
	"errors"
	"math"

	"carteira_investimentos/server/internal/domain"
)

// Calcular aplica a fórmula de Décio Bazin: Preço Teto = Proventos / Yield Mínimo
func Calcular(dividendos12M float64, yieldMinimo float64) (float64, error) {
	if yieldMinimo <= 0 {
		return 0, errors.New("o yield mínimo deve ser maior que zero (ex: 0.06 para 6%)")
	}
	precoTeto := dividendos12M / yieldMinimo
	return precoTeto, nil
}

// CalcularMargem calcula a margem percentual em relação à cotação atual
func CalcularMargem(precoAtual float64, precoTeto float64) float64 {
	if precoAtual <= 0 {
		return 0
	}
	return ((precoTeto - precoAtual) / precoAtual) * 100.0
}

// AnalisadorBazin implementa domain.Analisador
type AnalisadorBazin struct {
	YieldMinimo float64
}

func Novo(yieldMinimo float64) *AnalisadorBazin {
	return &AnalisadorBazin{YieldMinimo: yieldMinimo}
}

func (b *AnalisadorBazin) Nome() string {
	return "Décio Bazin"
}

func (b *AnalisadorBazin) Executar(ativo *domain.Ativo) error {
	if b.YieldMinimo <= 0 {
		b.YieldMinimo = 0.06
	}

	// 1. Caso FII com amortização/liquidação atípica: Bazin é inaplicável
	if ativo.Classe == domain.ClasseFII {
		if ativo.IsProventoAtipico {
			ativo.PrecoTetoBazin = 0
			ativo.PrecoTetoBazinSustentavel = 0
			return nil
		}
		teto, err := Calcular(ativo.Dividendos12M, b.YieldMinimo)
		if err != nil {
			return err
		}
		ativo.PrecoTetoBazin = teto
		ativo.PrecoTetoBazinSustentavel = teto
		return nil
	}

	// 2. Caso AÇÕES:
	// Se a empresa está em prejuízo contábil (LPA <= 0), proventos não têm base operacional
	if ativo.LPA <= 0 {
		ativo.PrecoTetoBazin = 0
		ativo.PrecoTetoBazinSustentavel = 0
		return nil
	}

	// Se o provento é atípico (venda de ativos, redução de capital, payout > 115% ou DY > 18%):
	// Aplica Bazin Prudencial: limita o dividendo à capacidade de lucro (payout sustentável de 60% do LPA)
	if ativo.IsProventoAtipico {
		divSustentavel := math.Min(ativo.Dividendos12M, ativo.LPA*0.60)
		tetoSustentavel, _ := Calcular(divSustentavel, b.YieldMinimo)
		ativo.PrecoTetoBazinSustentavel = tetoSustentavel
		ativo.PrecoTetoBazin = tetoSustentavel // Previne teto inflado no dashboard
		return nil
	}

	// Caso regular de dividendos sustentáveis
	teto, err := Calcular(ativo.Dividendos12M, b.YieldMinimo)
	if err != nil {
		return err
	}
	ativo.PrecoTetoBazin = teto
	ativo.PrecoTetoBazinSustentavel = teto
	return nil
}

