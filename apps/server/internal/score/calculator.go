package score

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"carteira_investimentos/server/internal/dcf"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/macro"
	"carteira_investimentos/server/internal/piotroski"
)

// AnalisadorScore consolida as métricas calculadas pelos outros motores,
// gera pareceres pedagógicos por escola de investimento
// e calcula o Score Fundamentalista Composto (0 a 100) com blindagem contra Value Traps.
type AnalisadorScore struct{}

func Novo() *AnalisadorScore {
	return &AnalisadorScore{}
}

func (s *AnalisadorScore) Nome() string {
	return "Score Fundamentalista & Pareceres"
}

func (s *AnalisadorScore) Executar(ativo *domain.Ativo) error {
	if ativo.Pareceres == nil {
		ativo.Pareceres = make(map[string]domain.ParecerItem)
	}

	// 1. Derivação de Métricas Comuns
	if ativo.PrecoAtual > 0 && ativo.Dividendos12M > 0 {
		ativo.DY = (ativo.Dividendos12M / ativo.PrecoAtual) * 100.0
	}

	// ==========================================
	// CASO 1: AÇÕES
	// ==========================================
	if ativo.Classe == domain.ClasseAcao {
		// Métricas fundamentalistas complementares
		if ativo.LPA > 0 && ativo.PrecoAtual > 0 {
			ativo.PL = ativo.PrecoAtual / ativo.LPA
			ativo.EarningsYield = (ativo.LPA / ativo.PrecoAtual) * 100.0
		}
		if ativo.VPA > 0 && ativo.PrecoAtual > 0 {
			ativo.PVPReal = ativo.PrecoAtual / ativo.VPA
		}
		if ativo.VPA > 0 && ativo.LPA > 0 {
			ativo.ROE = (ativo.LPA / ativo.VPA) * 100.0
		}
		if ativo.PrecoTetoBazin > 0 && ativo.PrecoAtual > 0 {
			ativo.MargemBazin = ((ativo.PrecoTetoBazin - ativo.PrecoAtual) / ativo.PrecoTetoBazin) * 100.0
		}
		if ativo.ValorGraham > 0 && ativo.PrecoAtual > 0 {
			ativo.MargemGraham = ((ativo.ValorGraham - ativo.PrecoAtual) / ativo.ValorGraham) * 100.0
		}

		// Parecer 1: Décio Bazin
		if ativo.IsProventoAtipico && ativo.PrecoTetoBazinSustentavel > 0 {
			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("Teto Prudente: R$ %6.2f | Yield: %4.1f%%", ativo.PrecoTetoBazinSustentavel, ativo.DY),
				Detalhe: fmt.Sprintf("⚠️ Yield Atípico (Payout %4.0f%%): Teto ajustado para capacidade de lucro real", ativo.Payout),
			}
		} else if ativo.IsProventoAtipico && ativo.LPA <= 0 {
			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("Teto: Inaplicável | Yield: %4.1f%%", ativo.DY),
				Detalhe: "⚠️ Prejuízo operacional: dividendo insustentável sem suporte de lucro",
			}
		} else if ativo.PrecoTetoBazin > 0 || ativo.PrecoTetoBazin5A > 0 {
			tetoRef := ativo.PrecoTetoBazin5A
			if tetoRef <= 0 {
				tetoRef = ativo.PrecoTetoBazin
			}
			margemRef := ativo.MargemBazin5A
			if ativo.PrecoTetoBazin5A <= 0 {
				margemRef = ativo.MargemBazin
			}

			status := domain.ParecerReprovado
			if ativo.PrecoAtual <= tetoRef {
				status = domain.ParecerAprovado
			}

			metricaStr := fmt.Sprintf("Teto 5A: R$ %6.2f | Teto 12M: R$ %6.2f", ativo.PrecoTetoBazin5A, ativo.PrecoTetoBazin)
			detalheStr := fmt.Sprintf("Média 5A: R$ %.2f (Margem 5A: %+.1f%% | 12M: %+.1f%%)", ativo.MediaDividendos5A, margemRef, ativo.MargemBazin)
			if ativo.AderenciaStatusInvest != "" {
				detalheStr += fmt.Sprintf(" • %s", ativo.AderenciaStatusInvest)
			}

			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  status,
				Metrica: metricaStr,
				Detalhe: detalheStr,
			}
		} else {
			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  domain.ParecerNaoAplica,
				Metrica: "DY 12M: 0,0%",
				Detalhe: "Não pagou proventos no último ano (reinvestimento ou crescimento)",
			}
		}

		// Parecer 2: Benjamin Graham
		if ativo.ValorGraham > 0 {
			if ativo.PrecoAtual <= ativo.ValorGraham {
				ativo.Pareceres["Graham"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("VI: R$ %6.2f", ativo.ValorGraham),
					Detalhe: fmt.Sprintf("Margem Seg: %+.1f%%", ativo.MargemGraham),
				}
			} else if ativo.PrecoAtual <= ativo.ValorGraham*1.15 {
				ativo.Pareceres["Graham"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("VI: R$ %6.2f", ativo.ValorGraham),
					Detalhe: fmt.Sprintf("Próximo do VI (+%4.1f%%)", -ativo.MargemGraham),
				}
			} else {
				ativo.Pareceres["Graham"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("VI: R$ %6.2f", ativo.ValorGraham),
					Detalhe: fmt.Sprintf("Acima do VI (+%4.1f%%)", -ativo.MargemGraham),
				}
			}
		} else {
			ativo.Pareceres["Graham"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: "VI: Inaplicável",
				Detalhe: "Prejuízo acumulado ou patrimônio líquido negativo",
			}
		}

		// Parecer 3: Peter Lynch (PEG Ratio)
		if ativo.PEGRatio > 0 {
			if ativo.PEGRatio <= 1.0 {
				ativo.Pareceres["Lynch"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | PEG: %4.2f", ativo.PrecoJustoLynch, ativo.PEGRatio),
					Detalhe: "Crescimento a preço muito atrativo",
				}
			} else if ativo.PEGRatio <= 1.5 {
				ativo.Pareceres["Lynch"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | PEG: %4.2f", ativo.PrecoJustoLynch, ativo.PEGRatio),
					Detalhe: "Crescimento a preço justo",
				}
			} else {
				ativo.Pareceres["Lynch"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | PEG: %4.2f", ativo.PrecoJustoLynch, ativo.PEGRatio),
					Detalhe: "Preço elevado para a taxa de crescimento",
				}
			}
		}

		// Parecer 4: Gordon DDM
		if ativo.IsProventoAtipico || ativo.Payout > 100.0 {
			if ativo.PrecoTetoGordon > 0 && ativo.PrecoAtual <= ativo.PrecoTetoGordon {
				ativo.Pareceres["Gordon"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("Teto Prudente: R$ %6.2f", ativo.PrecoTetoGordon),
					Detalhe: "Crescimento perpétuo calculado sobre payout sustentável de 60%",
				}
			} else {
				ativo.Pareceres["Gordon"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: "Teto: Inaplicável",
					Detalhe: "Provento atípico inviabiliza projeção perpétua de dividendos",
				}
			}
		} else if ativo.PrecoTetoGordon > 0 {
			if ativo.PrecoAtual <= ativo.PrecoTetoGordon {
				ativo.Pareceres["Gordon"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f", ativo.PrecoTetoGordon),
					Detalhe: "Dividendos com crescimento sustentável abaixo do teto",
				}
			} else {
				ativo.Pareceres["Gordon"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f", ativo.PrecoTetoGordon),
					Detalhe: "Acima do Teto de Gordon",
				}
			}
		}

		// Porte e Robustez Institucional da Ação
		if ativo.ValorMercado >= 15_000_000_000.0 && ativo.LiquidezMediaDiaria >= 10_000_000.0 {
			ativo.Porte = "BLUE_CHIP"
		} else if ativo.ValorMercado >= 3_000_000_000.0 && ativo.LiquidezMediaDiaria >= 2_000_000.0 {
			ativo.Porte = "MID_CAP"
		} else if ativo.ValorMercado >= 500_000_000.0 && ativo.LiquidezMediaDiaria >= 500_000.0 {
			ativo.Porte = "SMALL_CAP"
		} else {
			ativo.Porte = "MICRO_CAP"
		}

		// Parecer 5: Porte & Robustez Institucional
		switch ativo.Porte {
		case "BLUE_CHIP":
			ativo.Pareceres["Porte & Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("MktCap: R$ %5.1fB | Liq: R$ %5.1fM/dia", ativo.ValorMercado/1e9, ativo.LiquidezMediaDiaria/1e6),
				Detalhe: "🛡️ Blue Chip / Fortaleza Institucional da B3",
			}
		case "MID_CAP":
			ativo.Pareceres["Porte & Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("MktCap: R$ %5.1fB | Liq: R$ %5.1fM/dia", ativo.ValorMercado/1e9, ativo.LiquidezMediaDiaria/1e6),
				Detalhe: "Mid Cap Estabelecida com Boa Liquidez",
			}
		case "SMALL_CAP":
			ativo.Pareceres["Porte & Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("MktCap: R$ %5.0fM | Liq: R$ %5.0fK/dia", ativo.ValorMercado/1e6, ativo.LiquidezMediaDiaria/1e3),
				Detalhe: "Small Cap: Maior Volatilidade e Liquidez Moderada",
			}
		default:
			ativo.Pareceres["Porte & Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("MktCap: R$ %5.0fM | Liq: R$ %5.0fK/dia", ativo.ValorMercado/1e6, ativo.LiquidezMediaDiaria/1e3),
				Detalhe: "⚠️ Microcap / Baixa Liquidez: Risco de Execução",
			}
		}

		// Avaliação Macroeconômica de Risco Relativo vs NTN-B / Selic
		cenario := macro.ObterCenario()
		ativo.YieldExigido = math.Round((cenario.TaxaNTNB+cenario.SpreadMinimoAcoes)*100) / 100
		ativo.SpreadNTNB = math.Round((ativo.DY-cenario.TaxaNTNB)*100) / 100
		cresc := math.Min(math.Max(ativo.CrescimentoLucro5A, 0.0), 5.0)
		ativo.TIRProjetada = math.Round((ativo.DY+cresc)*10) / 10

		if ativo.SpreadNTNB >= 2.0 {
			ativo.VereditoRisco = "COMPENSA_RISCO"
			ativo.JustificativaRisco = fmt.Sprintf("🟢 Retorno Projetado atraente (+%.1f%% vs NTN-B). Remunera amplamente a volatilidade de equity.", ativo.SpreadNTNB)
		} else if ativo.SpreadNTNB >= 0.0 {
			ativo.VereditoRisco = "NEUTRO"
			ativo.JustificativaRisco = fmt.Sprintf("🟡 Retorno moderado (+%.1f%% vs NTN-B). Preço próximo do equilíbrio justo com a renda fixa.", ativo.SpreadNTNB)
		} else {
			ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
			ativo.JustificativaRisco = fmt.Sprintf("🔴 Dividend Yield (%.1f%%) inferior à NTN-B (%.1f%% a.a.). Risco de equity descompensado.", ativo.DY, cenario.TaxaNTNB)
		}

		statusRisco := domain.ParecerReprovado
		if ativo.VereditoRisco == "COMPENSA_RISCO" {
			statusRisco = domain.ParecerAprovado
		} else if ativo.VereditoRisco == "NEUTRO" {
			statusRisco = domain.ParecerAtencao
		}
		ativo.Pareceres["VereditoRisco"] = domain.ParecerItem{
			Status:  statusRisco,
			Metrica: fmt.Sprintf("Spread: %+.1f%% vs NTN-B | Hurdle: %.1f%%", ativo.SpreadNTNB, ativo.YieldExigido),
			Detalhe: ativo.JustificativaRisco,
		}

		// ==========================================
		// CÁLCULO DO SCORE DE AÇÕES (0 a 100)
		// ==========================================
		score := 0.0

		// Ajuste de Risco Macroeconômico no Score
		if ativo.VereditoRisco == "COMPENSA_RISCO" {
			score += 5.0
		} else if ativo.VereditoRisco == "RISCO_DESCOMPENSADO" {
			score = math.Max(0.0, score-8.0)
		}

		// Bloco 1: Bazin / Dividend Yield (até 15 pontos)
		if ativo.IsProventoAtipico {
			if ativo.PrecoTetoBazinSustentavel > 0 && ativo.PrecoAtual <= ativo.PrecoTetoBazinSustentavel {
				score += 6.0
			} else {
				score += 2.0
			}
		} else {
			if ativo.DY >= 8.0 {
				score += 15.0
			} else if ativo.DY >= 6.0 {
				score += 11.0 + (ativo.DY-6.0)*2.0
			} else if ativo.DY >= 4.0 {
				score += 6.0
			} else if ativo.DY > 0.0 {
				score += 3.0
			}
		}

		// Bloco 2: Graham / Desconto de Valor Intrínseco (até 15 pontos)
		if ativo.MargemGraham >= 25.0 {
			score += 15.0
		} else if ativo.MargemGraham >= 0.0 {
			score += 10.0 + (ativo.MargemGraham/25.0)*5.0
		} else if ativo.MargemGraham >= -15.0 {
			score += 4.0
		}

		// Bloco 3: Greenblatt (Magic Formula) (até 18 pontos)
		if pGreen, ok := ativo.Pareceres["Greenblatt"]; ok {
			if pGreen.Status == domain.ParecerAprovado {
				score += 18.0
			} else if pGreen.Status == domain.ParecerAtencao {
				score += 11.0
			} else if pGreen.Status == domain.ParecerReprovado && ativo.PL > 0 {
				score += 3.0
			}
		}

		// Bloco 4: Piotroski (Solvência & Saúde Contábil) (até 18 pontos)
		if pPio, ok := ativo.Pareceres["Piotroski"]; ok {
			if pPio.Status == domain.ParecerAprovado {
				score += 18.0
			} else if pPio.Status == domain.ParecerAtencao {
				score += 11.0
			} else {
				score += 3.0
			}
		}

		// Bloco 5: Lynch PEG & Crescimento (até 9 pontos)
		if ativo.PEGRatio > 0 {
			if ativo.PEGRatio <= 0.8 {
				score += 9.0
			} else if ativo.PEGRatio <= 1.2 {
				score += 6.0
			} else if ativo.PEGRatio <= 1.6 {
				score += 3.0
			}
		}

		// Bloco 6: Gordon Dividend Growth (até 10 pontos)
		if ativo.PrecoTetoGordon > 0 && ativo.PrecoAtual <= ativo.PrecoTetoGordon {
			score += 10.0
		} else if ativo.PrecoTetoGordon > 0 && ativo.PrecoAtual <= ativo.PrecoTetoGordon*1.15 {
			score += 5.0
		}

		// Bloco 7: Robustez Institucional & Porte (até 15 pontos)
		switch ativo.Porte {
		case "BLUE_CHIP":
			score += 15.0
		case "MID_CAP":
			score += 10.0
		case "SMALL_CAP":
			score += 4.0
		default:
			score += 0.0
		}

		// --- BLINDAGEM CONTRA VALUE TRAPS & MICROCAPS ILÍQUIDAS ---
		// Se a empresa opera em prejuízo (LPA <= 0) ou tem Patrimônio Negativo (VPA <= 0):
		if ativo.LPA <= 0 || ativo.VPA <= 0 {
			score = math.Min(score, 25.0)
		}

		// Penalidade de Yield Trap (provento atípico / payout > 115% / DY > 18%):
		if ativo.IsProventoAtipico {
			score = math.Max(0.0, score-15.0)
			score = math.Min(score, 65.0) // Trava prudencial: Impede recomendação de COMPRAR_MAIS
		}

		// Microcaps ou ativos de baixa liquidez (< R$ 1 milhão/dia):
		// Não podem roubar o topo das gigantes sólidas no ranking geral
		if ativo.Porte == "MICRO_CAP" || ativo.LiquidezMediaDiaria < 1_000_000.0 {
			score = math.Min(score, 78.0)
		}

		// Se a liquidez for menor que R$ 50k/dia (risco severo de saída):
		if ativo.LiquidezMediaDiaria < 50_000.0 && ativo.LiquidezMediaDiaria > 0 {
			score = math.Max(0.0, score-20.0)
		}

		ativo.Score = math.Round(score)
	}

	// ==========================================
	// CASO 2: FUNDOS IMOBILIÁRIOS (FIIs)
	// ==========================================
	if ativo.Classe == domain.ClasseFII {
		if ativo.PrecoTetoBazin > 0 && ativo.PrecoAtual > 0 {
			ativo.MargemBazin = ((ativo.PrecoTetoBazin - ativo.PrecoAtual) / ativo.PrecoTetoBazin) * 100.0
		}

		// Porte e Pulverização de Cotistas do FII
		if ativo.NumeroCotistas >= 150_000.0 && ativo.LiquidezMediaDiaria >= 3_000_000.0 {
			ativo.Porte = "FII_GIGANTE"
		} else if ativo.NumeroCotistas >= 50_000.0 && ativo.LiquidezMediaDiaria >= 1_000_000.0 {
			ativo.Porte = "FII_CONSOLIDADO"
		} else if ativo.NumeroCotistas >= 20_000.0 && ativo.LiquidezMediaDiaria >= 300_000.0 {
			ativo.Porte = "FII_MEDIO"
		} else {
			ativo.Porte = "FII_CONCENTRADO"
		}

		// Parecer: Porte & Pulverização
		switch ativo.Porte {
		case "FII_GIGANTE":
			ativo.Pareceres["Porte & Cotistas"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("%4.0fK cotistas | R$ %4.1fM/dia", ativo.NumeroCotistas/1000.0, ativo.LiquidezMediaDiaria/1e6),
				Detalhe: "🏰 FII Baleia: Altíssima Liquidez e Pulverização",
			}
		case "FII_CONSOLIDADO":
			ativo.Pareceres["Porte & Cotistas"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("%4.0fK cotistas | R$ %4.1fM/dia", ativo.NumeroCotistas/1000.0, ativo.LiquidezMediaDiaria/1e6),
				Detalhe: "FII Consolidado no Mercado",
			}
		case "FII_MEDIO":
			ativo.Pareceres["Porte & Cotistas"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("%4.0fK cotistas | R$ %4.0fK/dia", ativo.NumeroCotistas/1000.0, ativo.LiquidezMediaDiaria/1000.0),
				Detalhe: "FII de Médio Porte: Liquidez Moderada",
			}
		default:
			ativo.Pareceres["Porte & Cotistas"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("%4.0f cotistas | R$ %4.0fK/dia", ativo.NumeroCotistas, ativo.LiquidezMediaDiaria/1000.0),
				Detalhe: "⚠️ FII Concentrado / Baixa Liquidez",
			}
		}

		// Parecer 1: P/VP e Desconto Patrimonial Oficial CVM
		if ativo.PVP > 0 {
			if ativo.PVP <= 0.95 {
				ativo.Pareceres["PVP"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("P/VP: %4.2f (VP: R$ %6.2f)", ativo.PVP, ativo.VPCota),
					Detalhe: fmt.Sprintf("Desconto de %4.1f%% sobre laudo", (1.0-ativo.PVP)*100.0),
				}
			} else if ativo.PVP <= 1.02 {
				ativo.Pareceres["PVP"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("P/VP: %4.2f (VP: R$ %6.2f)", ativo.PVP, ativo.VPCota),
					Detalhe: "Negociando próximo ao valor justo patrimonial",
				}
			} else {
				ativo.Pareceres["PVP"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("P/VP: %4.2f (VP: R$ %6.2f)", ativo.PVP, ativo.VPCota),
					Detalhe: fmt.Sprintf("Ágio de %4.1f%% acima do patrimônio", (ativo.PVP-1.0)*100.0),
				}
			}
		}

		// Avaliação Macroeconômica de Risco Relativo de FIIs
		cenario := macro.ObterCenario()
		segUpper := strings.ToUpper(ativo.Segmento + " " + ativo.Setor)
		isPapel := strings.Contains(segUpper, "PAP") || strings.Contains(segUpper, "CRI") || strings.Contains(segUpper, "RECEB")
		isTijolo := strings.Contains(segUpper, "LOGÍST") || strings.Contains(segUpper, "LOGIST") || strings.Contains(segUpper, "SHOPPING") || strings.Contains(segUpper, "LAJES") || strings.Contains(segUpper, "IMÓVEIS") || strings.Contains(segUpper, "IMOVEIS") || strings.Contains(segUpper, "HOTEL") || strings.Contains(segUpper, "HOSPITAL") || strings.Contains(segUpper, "RENDA URBANA")

		ativo.SpreadNTNB = math.Round((ativo.DY-cenario.TaxaNTNB)*100) / 100

		if isPapel {
			ativo.YieldExigido = math.Round(cenario.TaxaSelic*0.85*10) / 10 // ex: 10.0% a.a. (CDI líquido isento)
			ativo.TIRProjetada = ativo.DY
		} else if isTijolo {
			spread := cenario.ObterSpreadSegmentoFII(ativo.Segmento)
			ativo.YieldExigido = math.Round((cenario.TaxaNTNB+spread)*10) / 10
			ativo.TIRProjetada = math.Round((ativo.DY+3.5)*10) / 10 // Dividend Yield + reajuste inflacionário longo
		} else {
			ativo.YieldExigido = math.Round((cenario.TaxaNTNB+2.0)*10) / 10
			ativo.TIRProjetada = math.Round((ativo.DY+2.5)*10) / 10
		}

		if ativo.IsProventoAtipico || ativo.DY >= 18.0 {
			ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
			ativo.JustificativaRisco = "⚠️ Amortização extraordinária ou devolução de capital: yield não sustentável."
		} else if isPapel {
			if ativo.PVP > 1.02 {
				ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
				ativo.JustificativaRisco = fmt.Sprintf("🔴 Ágio prejudicial em FII de Papel (P/VP %.2fx). Comprar crédito acima do VP corrói o retorno.", ativo.PVP)
			} else if ativo.DY >= ativo.YieldExigido && ativo.PVP <= 1.01 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 P/VP seguro (%.2fx) e Yield de %.1f%% a.a. supera o CDI isento (hurdle %.1f%%).", ativo.PVP, ativo.DY, ativo.YieldExigido)
			} else {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = "🟡 Preço justo e rendimento em linha com a renda fixa de crédito privado."
			}
		} else { // Tijolo / Híbridos
			if ativo.DY >= ativo.YieldExigido && ativo.PVP <= 1.02 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 Yield Real de %.1f%% a.a. supera o hurdle de %s (%.1f%%). Compensa o risco.", ativo.DY, ativo.Segmento, ativo.YieldExigido)
			} else if ativo.DY >= cenario.TaxaNTNB {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = fmt.Sprintf("🟡 Yield Real de %.1f%% a.a. supera a NTN-B básica, mas tem prêmio de risco reduzido.", ativo.DY)
			} else {
				ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
				ativo.JustificativaRisco = fmt.Sprintf("🔴 Yield Real (%.1f%%) inferior à NTN-B (%.1f%%). Não compensa o risco imobiliário.", ativo.DY, cenario.TaxaNTNB)
			}
		}

		statusRiscoFII := domain.ParecerReprovado
		if ativo.VereditoRisco == "COMPENSA_RISCO" {
			statusRiscoFII = domain.ParecerAprovado
		} else if ativo.VereditoRisco == "NEUTRO" {
			statusRiscoFII = domain.ParecerAtencao
		}
		ativo.Pareceres["VereditoRisco"] = domain.ParecerItem{
			Status:  statusRiscoFII,
			Metrica: fmt.Sprintf("Spread: %+.1f%% vs NTN-B | Hurdle: %.1f%%", ativo.SpreadNTNB, ativo.YieldExigido),
			Detalhe: ativo.JustificativaRisco,
		}

		// Parecer 2: Spread de Renda vs NTN-B (mantido para compatibilidade)
		ativo.Pareceres["SpreadNTNB"] = ativo.Pareceres["VereditoRisco"]

		// Parecer 3: Teto Bazin / Cap Rate FII
		if ativo.IsProventoAtipico {
			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: "Teto: Inaplicável",
				Detalhe: "Fundo em amortização extraordinária ou liquidação",
			}
		} else if ativo.PrecoTetoBazin > 0 || ativo.PrecoTetoBazin5A > 0 {
			tetoRef := ativo.PrecoTetoBazin5A
			if tetoRef <= 0 {
				tetoRef = ativo.PrecoTetoBazin
			}
			status := domain.ParecerReprovado
			if ativo.PrecoAtual <= tetoRef {
				status = domain.ParecerAprovado
			}
			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  status,
				Metrica: fmt.Sprintf("Teto 5A: R$ %6.2f | Teto 12M: R$ %6.2f", ativo.PrecoTetoBazin5A, ativo.PrecoTetoBazin),
				Detalhe: fmt.Sprintf("Média 5A: R$ %.2f (Margem 5A: %+.1f%%)", ativo.MediaDividendos5A, ativo.MargemBazin5A),
			}
		}

		// ==========================================
		// CÁLCULO DO SCORE DE FIIs (0 a 100)
		// ==========================================
		score := 0.0

		// Ajuste Macroeconômico no Score de FIIs
		if ativo.VereditoRisco == "COMPENSA_RISCO" {
			score += 5.0
		} else if ativo.VereditoRisco == "RISCO_DESCOMPENSADO" {
			score = math.Max(0.0, score-10.0)
		}

		// Bloco 1: Desconto P/VP com trava de ágio (até 25 pontos)
		if ativo.PVP > 0 {
			if ativo.PVP >= 0.85 && ativo.PVP <= 0.95 {
				score += 25.0 // Desconto ideal
			} else if ativo.PVP > 0.95 && ativo.PVP <= 1.01 {
				score += 20.0 // Preço justo
			} else if ativo.PVP < 0.85 && ativo.PVP >= 0.35 {
				score += 15.0 // Desconto acentuado
			} else if ativo.PVP <= 1.05 {
				score += 6.0
			}
		}

		// Bloco 2: Spread NTN-B (até 20 pontos)
		if !ativo.IsProventoAtipico && ativo.DY < 18.0 {
			if ativo.SpreadNTNB >= 3.0 {
				score += 20.0
			} else if ativo.SpreadNTNB >= 1.5 {
				score += 15.0
			} else if ativo.SpreadNTNB >= 0.0 {
				score += 9.0
			}
		}

		// Bloco 3: Dividend Yield 12M (até 20 pontos)
		if !ativo.IsProventoAtipico && ativo.DY < 18.0 {
			if ativo.DY >= 11.5 {
				score += 20.0
			} else if ativo.DY >= 9.5 {
				score += 16.0
			} else if ativo.DY >= 7.5 {
				score += 10.0
			} else if ativo.DY > 0.0 {
				score += 4.0
			}
		}

		// Bloco 4: Estabilidade de Proventos (até 10 pontos)
		if pProv, ok := ativo.Pareceres["FII Proventos"]; ok {
			if pProv.Status == domain.ParecerAprovado {
				score += 10.0
			} else if pProv.Status == domain.ParecerAtencao {
				score += 6.0
			}
		}

		// Bloco 5: Robustez Institucional, Base de Cotistas & Liquidez (até 25 pontos)
		switch ativo.Porte {
		case "FII_GIGANTE":
			score += 25.0
		case "FII_CONSOLIDADO":
			score += 18.0
		case "FII_MEDIO":
			score += 10.0
		default:
			score += 3.0
		}

		// --- BLINDAGEM DE RISCO DE FIIs ---
		// Se o fundo tem provento atípico (amortização/liquidação), P/VP em colapso (< 0.35) ou DY >= 18%:
		if ativo.IsProventoAtipico || ativo.PVP < 0.35 || ativo.DY >= 18.0 {
			score = math.Min(score, 30.0) // Trava de segurança para ALERTA imediato
		}

		// FIIs Concentrados / baixa liquidez (< 20k cotistas ou < R$ 300k/dia):
		// Não podem liderar sobre fundos consolidados
		if ativo.Porte == "FII_CONCENTRADO" {
			score = math.Min(score, 78.0)
		}

		// Se o fundo não pagou proventos (DY == 0):
		if ativo.DY <= 0 {
			score = math.Min(score, 30.0)
		}

		ativo.Score = math.Round(score)
	}

	// ==========================================
	// CASO 3: ETFs (Fundos de Índice)
	// ==========================================
	if ativo.Classe == domain.ClasseETF {
		score := 0.0
		vol := ativo.VolumeTotal

		if vol >= 50_000_000.0 {
			score = 95.0
			ativo.Pareceres["Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("Volume: R$ %5.1fM/dia", vol/1_000_000.0),
				Detalhe: "Altíssima Liquidez Institucional",
			}
		} else if vol >= 10_000_000.0 {
			score = 85.0
			ativo.Pareceres["Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("Volume: R$ %5.1fM/dia", vol/1_000_000.0),
				Detalhe: "Excelente Liquidez",
			}
		} else if vol >= 5_000_000.0 {
			score = 75.0
			ativo.Pareceres["Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("Volume: R$ %5.1fM/dia", vol/1_000_000.0),
				Detalhe: "Boa Liquidez de Negociação",
			}
		} else if vol >= 1_000_000.0 {
			score = 60.0
			ativo.Pareceres["Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("Volume: R$ %5.1fM/dia", vol/1_000_000.0),
				Detalhe: "Liquidez Moderada",
			}
		} else if vol >= 500_000.0 {
			score = 50.0
			ativo.Pareceres["Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("Volume: R$ %5.1fK/dia", vol/1_000.0),
				Detalhe: "Liquidez Regular (Atenção ao Spread)",
			}
		} else {
			score = 30.0
			ativo.Pareceres["Liquidez"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("Volume: R$ %5.1fK/dia", vol/1_000.0),
				Detalhe: "Baixa Liquidez",
			}
		}

		ativo.Score = score
	}

	// 2. Alinha o Status do Semáforo ao Score Consolidado e Regras de Segurança
	ativo.AvaliarSemaforo()

	// 3. Consolidação Consensual Multi-Modelo do Preço Teto e % de Prêmio/Upside
	ConsolidarPrecoTeto(ativo)

	return nil
}

// ModeloValuation armazena o teto e peso institucional de um modelo quantitativo
type ModeloValuation struct {
	Nome string
	Teto float64
	Peso float64
}

// FiltrarSanidadeEstatistica remove valores <= 0, nulos ou que apresentem desvio superior a 1,5 desvio-padrão da mediana
func FiltrarSanidadeEstatistica(modelos []ModeloValuation) []ModeloValuation {
	var validos []ModeloValuation
	for _, m := range modelos {
		if m.Teto > 0 && !math.IsNaN(m.Teto) && !math.IsInf(m.Teto, 0) {
			validos = append(validos, m)
		}
	}

	if len(validos) <= 2 {
		return validos
	}

	valores := make([]float64, len(validos))
	for i, m := range validos {
		valores[i] = m.Teto
	}
	sort.Float64s(valores)

	n := len(valores)
	var mediana float64
	if n%2 == 1 {
		mediana = valores[n/2]
	} else {
		mediana = (valores[n/2-1] + valores[n/2]) / 2.0
	}

	soma := 0.0
	for _, v := range valores {
		soma += v
	}
	media := soma / float64(n)

	somaDiffSq := 0.0
	for _, v := range valores {
		somaDiffSq += (v - media) * (v - media)
	}
	desvioPadrao := math.Sqrt(somaDiffSq / float64(n))

	if desvioPadrao < 1e-4 {
		return validos
	}

	limite := 1.5 * desvioPadrao
	var sobreviventes []ModeloValuation
	for _, m := range validos {
		if math.Abs(m.Teto-mediana) <= limite {
			sobreviventes = append(sobreviventes, m)
		}
	}

	if len(sobreviventes) == 0 {
		return validos
	}

	return sobreviventes
}

// ConsolidarPrecoTeto calcula o Preço Teto Consensual e o Prêmio/Upside (%)
// cruzando as diferentes metodologias de valuation de forma robusta e ponderada
func ConsolidarPrecoTeto(ativo *domain.Ativo) {
	if ativo.PrecoAtual <= 0 {
		return
	}

	cenario := macro.ObterCenario()
	var modelosValidos []ModeloValuation

	// ==========================================
	// CASO 1: AÇÕES
	// ==========================================
	if ativo.Classe == domain.ClasseAcao {
		isFinanceiro := dcf.IsSetorFinanceiro(ativo.Setor, ativo.Subsetor, ativo.Segmento)

		// 1. Décio Bazin 5 Anos (Normalizado com Winsorização anti-outlier e calibrado com NTN-B)
		yieldMinimoAcoes := cenario.ObterYieldMinimoAcoes()
		media5A := ativo.MediaDividendos5ANormalizada
		if media5A <= 0 {
			media5A = ativo.MediaDividendos5A
		}
		tetoBazin5A := ativo.PrecoTetoBazin5A
		if media5A > 0 && yieldMinimoAcoes > 0 {
			tetoBazin5A = math.Round((media5A/yieldMinimoAcoes)*100) / 100
			ativo.PrecoTetoBazin5A = tetoBazin5A
			if ativo.PrecoAtual > 0 {
				ativo.MargemBazin5A = math.Round(((ativo.PrecoTetoBazin5A-ativo.PrecoAtual)/ativo.PrecoTetoBazin5A)*1000) / 10
			}
		}

		if isFinanceiro {
			// Setor Financeiro / Bancos / Seguradoras:
			// Ponderação: 50% Bazin 5A + 30% Gordon + 20% Lynch
			ativo.PrecoTetoDCF = 0 // DCF nulo para bancos

			if tetoBazin5A > 0 {
				modelosValidos = append(modelosValidos, ModeloValuation{
					Nome: fmt.Sprintf("Bazin 5A (%.1f%%)", yieldMinimoAcoes*100),
					Teto: tetoBazin5A,
					Peso: 0.50,
				})
			}

			if ativo.PrecoTetoGordon > 0 {
				modelosValidos = append(modelosValidos, ModeloValuation{
					Nome: "Gordon DDM",
					Teto: ativo.PrecoTetoGordon,
					Peso: 0.30,
				})
			}

			if ativo.PrecoJustoLynch > 0 && ativo.LPA > 0 {
				tetoLynchPrudente := math.Min(ativo.PrecoJustoLynch, 18.0*ativo.LPA)
				if tetoLynchPrudente > 0 {
					modelosValidos = append(modelosValidos, ModeloValuation{
						Nome: "Peter Lynch",
						Teto: tetoLynchPrudente,
						Peso: 0.20,
					})
				}
			}
		} else {
			// Setor Real / Empresas Operacionais:
			// Ponderação: 35% DCF Proxy + 35% Bazin 5A + 15% Graham + 15% Gordon

			// Se DCF ainda não tiver sido calculado, calcula agora
			if ativo.PrecoTetoDCF <= 0 {
				tetoDCF, _, _, _, err := dcf.CalcularDCF(ativo, cenario)
				if err == nil && tetoDCF > 0 {
					ativo.PrecoTetoDCF = tetoDCF
				}
			}

			if ativo.PrecoTetoDCF > 0 {
				modelosValidos = append(modelosValidos, ModeloValuation{
					Nome: "DCF Proxy",
					Teto: ativo.PrecoTetoDCF,
					Peso: 0.35,
				})
			}

			if tetoBazin5A > 0 {
				modelosValidos = append(modelosValidos, ModeloValuation{
					Nome: fmt.Sprintf("Bazin 5A (%.1f%%)", yieldMinimoAcoes*100),
					Teto: tetoBazin5A,
					Peso: 0.35,
				})
			}

			if ativo.ValorGraham > 0 && ativo.LPA > 0 && ativo.VPA > 0 {
				modelosValidos = append(modelosValidos, ModeloValuation{
					Nome: "Graham",
					Teto: ativo.ValorGraham,
					Peso: 0.15,
				})
			}

			if ativo.PrecoTetoGordon > 0 {
				modelosValidos = append(modelosValidos, ModeloValuation{
					Nome: "Gordon DDM",
					Teto: ativo.PrecoTetoGordon,
					Peso: 0.15,
				})
			}
		}

		// Filtro de Sanidade Estatística (elimina desvios > 1.5 DP da mediana)
		modelosSobreviventes := FiltrarSanidadeEstatistica(modelosValidos)

		if len(modelosSobreviventes) == 0 {
			return
		}

		somaPesos := 0.0
		for _, m := range modelosSobreviventes {
			somaPesos += m.Peso
		}

		tetoConsolidado := 0.0
		var nomesModelos []string
		for _, m := range modelosSobreviventes {
			pesoNorm := m.Peso / somaPesos
			tetoConsolidado += m.Teto * pesoNorm
			nomesModelos = append(nomesModelos, fmt.Sprintf("%s: R$ %.2f (%.0f%%)", m.Nome, m.Teto, pesoNorm*100))
		}
		tetoConsolidado = math.Round(tetoConsolidado*100) / 100

		premioUpside := ((tetoConsolidado - ativo.PrecoAtual) / ativo.PrecoAtual) * 100.0
		margemSeg := ((tetoConsolidado - ativo.PrecoAtual) / tetoConsolidado) * 100.0

		ativo.PrecoTetoConsolidado = tetoConsolidado
		ativo.PremioDescontoPercentual = math.Round(premioUpside*10) / 10
		ativo.MargemSegurancaConsolidada = math.Round(margemSeg*10) / 10
		ativo.ModelosTetoConsolidado = nomesModelos

		// Regra de Margem de Segurança & Veredito ancorada no Piotroski F-Score:
		fScore := ativo.PiotroskiScore
		if fScore == 0 {
			fScore = piotroski.CalcularFScore(ativo)
			ativo.PiotroskiScore = fScore
		}

		if fScore <= 4 {
			ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
			ativo.JustificativaRisco = fmt.Sprintf("🔴 Alerta de Solvência / Value Trap: Piotroski F-Score crítico (%d/9 pts). Risco contábil descompensado.", fScore)
		} else if fScore <= 6 {
			// Margem mínima recomendada = 25%
			if margemSeg >= 25.0 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 Margem de segurança de %.1f%% supera o piso de 25%% exigido para saúde financeira moderada (F-Score %d/9).", margemSeg, fScore)
			} else if margemSeg >= 0.0 {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = fmt.Sprintf("🟡 Margem de segurança de %.1f%% é positiva, mas inferior ao piso de 25%% recomendado (F-Score %d/9).", margemSeg, fScore)
			} else {
				ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
				ativo.JustificativaRisco = fmt.Sprintf("🔴 Preço atual acima do teto de consenso (Margem %.1f%%, F-Score %d/9). Risco descompensado.", margemSeg, fScore)
			}
		} else { // fScore >= 7
			// Margem mínima recomendada = 15%
			if margemSeg >= 15.0 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 Alta solidez financeira (F-Score %d/9) e margem de segurança de %.1f%% acima dos 15%% mínimos.", fScore, margemSeg)
			} else if margemSeg >= 0.0 {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = fmt.Sprintf("🟡 Empresa saudável (F-Score %d/9), mas negociando próxima ao preço justo (Margem %.1f%%).", fScore, margemSeg)
			} else {
				ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
				ativo.JustificativaRisco = fmt.Sprintf("🔴 Negociando com ágio sobre o preço teto consensual (Margem %.1f%%, F-Score %d/9).", margemSeg, fScore)
			}
		}

		statusTeto := domain.ParecerReprovado
		if ativo.VereditoRisco == "COMPENSA_RISCO" {
			statusTeto = domain.ParecerAprovado
		} else if ativo.VereditoRisco == "NEUTRO" {
			statusTeto = domain.ParecerAtencao
		}

		detalhe := fmt.Sprintf("Modelos: %s", strings.Join(nomesModelos, " | "))
		if ativo.TeveOutlier5A && ativo.ObservacaoOutlier != "" {
			detalhe += fmt.Sprintf(" • 🛡️ %s", ativo.ObservacaoOutlier)
		}

		ativo.Pareceres["Preço Teto"] = domain.ParecerItem{
			Status:  statusTeto,
			Metrica: fmt.Sprintf("Teto Consenso: R$ %.2f | Prêmio: %+.1f%%", tetoConsolidado, ativo.PremioDescontoPercentual),
			Detalhe: detalhe,
		}

		ativo.Pareceres["VereditoRisco"] = domain.ParecerItem{
			Status:  statusTeto,
			Metrica: fmt.Sprintf("F-Score: %d/9 pts | Margem: %+.1f%%", fScore, margemSeg),
			Detalhe: ativo.JustificativaRisco,
		}

	} else if ativo.Classe == domain.ClasseFII {
		// ==========================================
		// CASO 2: FUNDOS IMOBILIÁRIOS (FIIs)
		// ==========================================
		segUpper := strings.ToUpper(ativo.Segmento + " " + ativo.Setor)
		isPapel := strings.Contains(segUpper, "PAP") || strings.Contains(segUpper, "CRI") || strings.Contains(segUpper, "RECEB")
		isTijolo := strings.Contains(segUpper, "LOGÍST") || strings.Contains(segUpper, "LOGIST") || strings.Contains(segUpper, "SHOPPING") || strings.Contains(segUpper, "LAJES") || strings.Contains(segUpper, "IMÓVEIS") || strings.Contains(segUpper, "IMOVEIS") || strings.Contains(segUpper, "HOTEL") || strings.Contains(segUpper, "HOSPITAL") || strings.Contains(segUpper, "RENDA URBANA")

		tetoConsolidado := 0.0
		var nomesModelos []string

		if isPapel {
			// FIIs de Papel (CRI): Se P/VP > 1.02, emitir alerta de ágio de risco. Teto ancorado no VP Cota
			tetoPapel := ativo.VPCota
			if tetoPapel <= 0 {
				tetoPapel = ativo.PrecoAtual
			}
			tetoConsolidado = math.Round(tetoPapel*100) / 100
			nomesModelos = []string{fmt.Sprintf("VP Cota (Crédito): R$ %.2f", tetoConsolidado)}

			if ativo.PVP > 1.02 {
				ativo.AlertaRisco = "Ágio em FII de Papel"
				ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
				ativo.JustificativaRisco = fmt.Sprintf("🔴 Ágio prejudicial em FII de Papel (P/VP %.2fx). Comprar crédito acima do VP corrói o retorno.", ativo.PVP)
			} else if ativo.DY >= (cenario.TaxaSelic*0.85) && ativo.PVP <= 1.01 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 P/VP seguro (%.2fx) e Yield isento de %.1f%% a.a. supera o CDI líquido.", ativo.PVP, ativo.DY)
			} else {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = "🟡 P/VP em linha com o valor patrimonial contábil."
			}
		} else if isTijolo {
			// Tijolo: Preço Teto = dividendos_12m / (taxa_ntnb + spread_segmento_fii)
			spread := cenario.ObterSpreadSegmentoFII(ativo.Segmento)
			yieldExigido := (cenario.TaxaNTNB + spread) / 100.0

			divRef := ativo.Dividendos12M
			if divRef <= 0 && ativo.MediaDividendos5A > 0 {
				divRef = ativo.MediaDividendos5A
			}

			tetoTijolo := 0.0
			if yieldExigido > 0 && divRef > 0 {
				tetoTijolo = math.Round((divRef/yieldExigido)*100) / 100
			}

			tetoConsolidado = tetoTijolo
			nomesModelos = []string{fmt.Sprintf("Cap Rate (NTN-B+%.1f%%): R$ %.2f", spread, tetoTijolo)}

			if ativo.VPCota > 0 {
				nomesModelos = append(nomesModelos, fmt.Sprintf("VP CVM (Reposição): R$ %.2f", ativo.VPCota))
			}

			if ativo.DY >= (cenario.TaxaNTNB+spread) && ativo.PVP <= 1.02 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 Yield Real de %.1f%% a.a. supera o hurdle setorial de %.1f%%. Compensa o risco.", ativo.DY, cenario.TaxaNTNB+spread)
			} else if ativo.DY >= cenario.TaxaNTNB {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = fmt.Sprintf("🟡 Yield Real de %.1f%% a.a. supera a NTN-B básica, mas com prêmio reduzido.", ativo.DY)
			} else {
				ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
				ativo.JustificativaRisco = fmt.Sprintf("🔴 Yield Real (%.1f%%) inferior à NTN-B (%.1f%%). Risco imobiliário descompensado.", ativo.DY, cenario.TaxaNTNB)
			}
		} else {
			// FOFs e Híbridos
			tetoConsolidado = math.Round(ativo.VPCota*0.98*100) / 100
			if tetoConsolidado <= 0 {
				tetoConsolidado = ativo.PrecoAtual
			}
			nomesModelos = []string{fmt.Sprintf("VP FOF (Desconto Duplo): R$ %.2f", tetoConsolidado)}

			if ativo.PVP <= 0.95 {
				ativo.VereditoRisco = "COMPENSA_RISCO"
				ativo.JustificativaRisco = fmt.Sprintf("🟢 Desconto patrimonial duplo (P/VP %.2fx).", ativo.PVP)
			} else {
				ativo.VereditoRisco = "NEUTRO"
				ativo.JustificativaRisco = "🟡 Negociando próximo ao valor patrimonial."
			}
		}

		if tetoConsolidado > 0 {
			premioUpside := ((tetoConsolidado - ativo.PrecoAtual) / ativo.PrecoAtual) * 100.0
			margemSeg := ((tetoConsolidado - ativo.PrecoAtual) / tetoConsolidado) * 100.0

			ativo.PrecoTetoConsolidado = tetoConsolidado
			ativo.PremioDescontoPercentual = math.Round(premioUpside*10) / 10
			ativo.MargemSegurancaConsolidada = math.Round(margemSeg*10) / 10
			ativo.ModelosTetoConsolidado = nomesModelos

			statusTeto := domain.ParecerReprovado
			if ativo.VereditoRisco == "COMPENSA_RISCO" {
				statusTeto = domain.ParecerAprovado
			} else if ativo.VereditoRisco == "NEUTRO" {
				statusTeto = domain.ParecerAtencao
			}

			ativo.Pareceres["Preço Teto"] = domain.ParecerItem{
				Status:  statusTeto,
				Metrica: fmt.Sprintf("Teto Consenso: R$ %.2f | Prêmio: %+.1f%%", tetoConsolidado, ativo.PremioDescontoPercentual),
				Detalhe: fmt.Sprintf("Modelos: %s", strings.Join(nomesModelos, " | ")),
			}
			ativo.Pareceres["VereditoRisco"] = domain.ParecerItem{
				Status:  statusTeto,
				Metrica: fmt.Sprintf("Spread: %+.1f%% vs NTN-B | P/VP: %.2fx", ativo.SpreadNTNB, ativo.PVP),
				Detalhe: ativo.JustificativaRisco,
			}
		}
	}
}
