package greenblatt

import (
	"fmt"
	"strings"

	"carteira_investimentos/server/internal/domain"
)

// AnalisadorGreenblatt implementa The Magic Formula (Joel Greenblatt)
// adaptada para o mercado brasileiro com tratamento de financeiras.
type AnalisadorGreenblatt struct{}

func Novo() *AnalisadorGreenblatt {
	return &AnalisadorGreenblatt{}
}

func (g *AnalisadorGreenblatt) Nome() string {
	return "Joel Greenblatt (Magic Formula)"
}

func (g *AnalisadorGreenblatt) Executar(ativo *domain.Ativo) error {
	if ativo.Classe != domain.ClasseAcao {
		return nil
	}

	if ativo.Pareceres == nil {
		ativo.Pareceres = make(map[string]domain.ParecerItem)
	}

	isFinanceiro := strings.Contains(strings.ToUpper(ativo.Setor), "FINANC") ||
		strings.Contains(strings.ToUpper(ativo.Setor), "SEGURO") ||
		strings.Contains(strings.ToUpper(ativo.Segmento), "BANCO")

	// 1. Caso Especial: Instituições Financeiras (Bancos/Seguradoras)
	// Como bancos não possuem EBIT nem Dívida Líquida, a adaptação padrão de Wall Street
	// utiliza P/L (Preço sobre Lucro) como Earnings Yield e ROE como Retorno sobre o Capital.
	if isFinanceiro {
		if ativo.PL <= 0 || ativo.ROE <= 0 {
			ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("P/L: %4.1fx | ROE: %4.1f%%", ativo.PL, ativo.ROE),
				Detalhe: "Instituição financeira com lucro/ROE negativo",
			}
			return nil
		}

		// Critério adaptado: P/L <= 9.0x e ROE >= 14%
		if ativo.PL <= 9.0 && ativo.ROE >= 14.0 {
			ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("P/L: %4.1fx | ROE: %4.1f%%", ativo.PL, ativo.ROE),
				Detalhe: "Magic Formula Financeira: Barato e Alto Retorno",
			}
		} else if ativo.PL <= 13.0 && ativo.ROE >= 10.0 {
			ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("P/L: %4.1fx | ROE: %4.1f%%", ativo.PL, ativo.ROE),
				Detalhe: "Magic Formula Financeira: Múltiplos regulares",
			}
		} else {
			ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("P/L: %4.1fx | ROE: %4.1f%%", ativo.PL, ativo.ROE),
				Detalhe: "Magic Formula Financeira: P/L elevado ou ROE baixo",
			}
		}
		return nil
	}

	// 2. Empresas Não-Financeiras (Fórmula Mágica Pura: EV/EBIT + ROIC)
	evEbit := ativo.DividaLiquidaEbit // StatusInvest tem ev_ebit direto no Ativo
	// Se ev_ebit não estiver disponível, usamos P/L e ROIC
	roic := ativo.ROIC

	if roic <= 0 {
		ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
			Status:  domain.ParecerReprovado,
			Metrica: fmt.Sprintf("ROIC: %4.1f%%", roic),
			Detalhe: "Retorno sobre o Capital Investido nulo ou negativo",
		}
		return nil
	}

	// Se temos P/L positivo e ROIC forte
	if ativo.PL > 0 && ativo.PL <= 10.0 && roic >= 15.0 {
		ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
			Status:  domain.ParecerAprovado,
			Metrica: fmt.Sprintf("P/L: %4.1fx | ROIC: %4.1f%%", ativo.PL, roic),
			Detalhe: "Fórmula Mágica: Empresa excelente a preço de pechincha",
		}
	} else if (ativo.PL > 0 && ativo.PL <= 15.0) && roic >= 10.0 {
		ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
			Status:  domain.ParecerAtencao,
			Metrica: fmt.Sprintf("P/L: %4.1fx | ROIC: %4.1f%%", ativo.PL, roic),
			Detalhe: "Fórmula Mágica: Rentabilidade e preço moderados",
		}
	} else {
		ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
			Status:  domain.ParecerReprovado,
			Metrica: fmt.Sprintf("P/L: %4.1fx | ROIC: %4.1f%%", ativo.PL, roic),
			Detalhe: "Fórmula Mágica: Fora dos parâmetros de Greenblatt",
		}
	}
	_ = evEbit
	return nil
}
