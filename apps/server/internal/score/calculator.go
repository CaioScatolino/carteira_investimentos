package score

import (
	"fmt"
	"math"

	"carteira_investimentos/server/internal/domain"
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
		} else if ativo.PrecoTetoBazin > 0 {
			if ativo.PrecoAtual <= ativo.PrecoTetoBazin {
				ativo.Pareceres["Bazin"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | Yield: %4.1f%%", ativo.PrecoTetoBazin, ativo.DY),
					Detalhe: fmt.Sprintf("Margem de Segurança: %+.1f%%", ativo.MargemBazin),
				}
			} else {
				ativo.Pareceres["Bazin"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | Yield: %4.1f%%", ativo.PrecoTetoBazin, ativo.DY),
					Detalhe: fmt.Sprintf("Acima do Teto: %+.1f%%", -ativo.MargemBazin),
				}
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

		// ==========================================
		// CÁLCULO DO SCORE DE AÇÕES (0 a 100)
		// ==========================================
		score := 0.0

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

		// Parecer 2: Spread de Renda vs NTN-B (IPCA+ 6.5% a.a.)
		if ativo.IsProventoAtipico || ativo.DY >= 18.0 {
			ativo.Pareceres["SpreadNTNB"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("Spread: Inaplicável (DY %4.1f%%)", ativo.DY),
				Detalhe: "Provento extraordinário de amortização não reflete renda perpétua",
			}
		} else if ativo.SpreadNTNB >= 2.0 {
			ativo.Pareceres["SpreadNTNB"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("Spread: %+.2f%% vs NTN-B", ativo.SpreadNTNB),
				Detalhe: fmt.Sprintf("Yield Efetivo: %4.1f%% a.a.", ativo.DY),
			}
		} else if ativo.SpreadNTNB >= 0.0 {
			ativo.Pareceres["SpreadNTNB"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("Spread: %+.2f%% vs NTN-B", ativo.SpreadNTNB),
				Detalhe: "Prêmio de risco moderado",
			}
		} else {
			ativo.Pareceres["SpreadNTNB"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("Spread: %+.2f%% vs NTN-B", ativo.SpreadNTNB),
				Detalhe: "Rende menos que o Tesouro IPCA+",
			}
		}

		// Parecer 3: Teto Bazin FII
		if ativo.IsProventoAtipico {
			ativo.Pareceres["Bazin"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: "Teto: Inaplicável",
				Detalhe: "Fundo em amortização extraordinária ou liquidação",
			}
		} else if ativo.PrecoTetoBazin > 0 {
			if ativo.PrecoAtual <= ativo.PrecoTetoBazin {
				ativo.Pareceres["Bazin"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f", ativo.PrecoTetoBazin),
					Detalhe: fmt.Sprintf("Yield: %4.1f%% a.a.", ativo.DY),
				}
			} else {
				ativo.Pareceres["Bazin"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f", ativo.PrecoTetoBazin),
					Detalhe: "Cotação acima do teto de 6%",
				}
			}
		}

		// ==========================================
		// CÁLCULO DO SCORE DE FIIs (0 a 100)
		// ==========================================
		score := 0.0

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

	return nil
}
