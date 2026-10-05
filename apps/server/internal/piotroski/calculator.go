package piotroski

import (
	"fmt"
	"strings"

	"carteira_investimentos/server/internal/domain"
)

// AnalisadorPiotroski calcula o F-Score de Joseph Piotroski (0 a 9)
// avaliando Rentabilidade, Estrutura de Capital e Eficiência Operacional.
type AnalisadorPiotroski struct{}

func Novo() *AnalisadorPiotroski {
	return &AnalisadorPiotroski{}
}

func (p *AnalisadorPiotroski) Nome() string {
	return "Joseph Piotroski (F-Score)"
}

func (p *AnalisadorPiotroski) Executar(ativo *domain.Ativo) error {
	if ativo.Classe != domain.ClasseAcao {
		return nil
	}

	if ativo.Pareceres == nil {
		ativo.Pareceres = make(map[string]domain.ParecerItem)
	}

	score := 0
	isFinanceiro := strings.Contains(strings.ToUpper(ativo.Setor), "FINANC") ||
		strings.Contains(strings.ToUpper(ativo.Setor), "SEGURO") ||
		strings.Contains(strings.ToUpper(ativo.Segmento), "BANCO")

	// --- 1. RENTABILIDADE (ROA, Lucro e Margens) ---
	if ativo.ROA > 0 {
		score++
	}
	if ativo.MargemLiquida > 0 {
		score++
	}
	if ativo.MargemLiquida >= 8.0 {
		score++
	}
	if ativo.ROE >= 12.0 {
		score++
	}

	// --- 2. ESTRUTURA DE CAPITAL & LIQUIDEZ ---
	if isFinanceiro {
		// Para bancos e seguradoras: solvência por Liquidez e ROE superior
		if ativo.LiquidezCorrente >= 1.0 || ativo.ROE >= 14.0 {
			score++
		}
		if ativo.VPA > 0 {
			score++
		}
	} else {
		// Dívida líquida sob controle (< 2.5x EBIT ou < 1.2x PL)
		if (ativo.DividaLiquidaEbit >= 0 && ativo.DividaLiquidaEbit <= 2.5) || (ativo.DividaLiquidaPL >= 0 && ativo.DividaLiquidaPL <= 1.2) {
			score++
		}
		// Liquidez corrente confortável (> 1.1x)
		if ativo.LiquidezCorrente >= 1.1 {
			score++
		}
	}

	// --- 3. EFICIÊNCIA OPERACIONAL & CRESCIMENTO ---
	if ativo.MargemBruta >= 20.0 {
		score++
	}
	if ativo.GiroAtivos >= 0.4 || isFinanceiro {
		score++
	}
	if ativo.CrescimentoLucro5A > 0 {
		score++
	}

	ativo.PiotroskiScore = score

	// Gera o Parecer
	status := domain.ParecerReprovado
	detalhe := "Risco de Solvência / Alerta de Value Trap"

	if score >= 7 {
		status = domain.ParecerAprovado
		detalhe = "Solvência Blindada: Empresa Financeiramente Saudável"
	} else if score >= 5 {
		status = domain.ParecerAtencao
		detalhe = "Saúde Financeira Moderada: Monitorar Endividamento"
	}

	ativo.Pareceres["Piotroski"] = domain.ParecerItem{
		Status:  status,
		Metrica: fmt.Sprintf("F-Score: %d/9 pts", score),
		Detalhe: detalhe,
	}

	return nil
}

// CalcularFScore executa os 9 critérios de Piotroski e retorna a pontuação de 0 a 9
func CalcularFScore(ativo *domain.Ativo) int {
	if ativo.Classe != domain.ClasseAcao {
		return 0
	}
	score := 0
	isFinanceiro := strings.Contains(strings.ToUpper(ativo.Setor), "FINANC") ||
		strings.Contains(strings.ToUpper(ativo.Setor), "SEGURO") ||
		strings.Contains(strings.ToUpper(ativo.Segmento), "BANCO")

	if ativo.ROA > 0 {
		score++
	}
	if ativo.MargemLiquida > 0 {
		score++
	}
	if ativo.MargemLiquida >= 8.0 {
		score++
	}
	if ativo.ROE >= 12.0 {
		score++
	}

	if isFinanceiro {
		if ativo.LiquidezCorrente >= 1.0 || ativo.ROE >= 14.0 {
			score++
		}
		if ativo.VPA > 0 {
			score++
		}
	} else {
		if (ativo.DividaLiquidaEbit >= 0 && ativo.DividaLiquidaEbit <= 2.5) || (ativo.DividaLiquidaPL >= 0 && ativo.DividaLiquidaPL <= 1.2) {
			score++
		}
		if ativo.LiquidezCorrente >= 1.1 {
			score++
		}
	}

	if ativo.MargemBruta >= 20.0 {
		score++
	}
	if ativo.GiroAtivos >= 0.4 || isFinanceiro {
		score++
	}
	if ativo.CrescimentoLucro5A > 0 {
		score++
	}
	return score
}
