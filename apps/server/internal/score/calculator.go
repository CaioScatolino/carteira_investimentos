package score

import (
	"fmt"
	"math"

	"carteira_investimentos/server/internal/domain"
)

// AnalisadorScore consolida as métricas calculadas pelos outros motores,
// gera pareceres individuais pedagógicos por escola de investimento
// e calcula o Score Fundamentalista Composto (0 a 100).
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
		if ativo.PrecoTetoBazin > 0 {
			if ativo.PrecoAtual <= ativo.PrecoTetoBazin {
				ativo.Pareceres["Bazin"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | Yield: %4.1f%%", ativo.PrecoTetoBazin, ativo.DY),
					Detalhe: fmt.Sprintf("Margem: %+.1f%%", ativo.MargemBazin),
				}
			} else {
				ativo.Pareceres["Bazin"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | Yield: %4.1f%%", ativo.PrecoTetoBazin, ativo.DY),
					Detalhe: fmt.Sprintf("Acima do Teto: %+.1f%%", -ativo.MargemBazin),
				}
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
		}

		// Parecer 3: Peter Lynch (PEG Ratio)
		if ativo.PEGRatio > 0 {
			if ativo.PEGRatio <= 1.0 {
				ativo.Pareceres["Lynch"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | PEG: %4.2f", ativo.PrecoJustoLynch, ativo.PEGRatio),
					Detalhe: "Subavaliado pelo Crescimento",
				}
			} else if ativo.PEGRatio <= 1.5 {
				ativo.Pareceres["Lynch"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | PEG: %4.2f", ativo.PrecoJustoLynch, ativo.PEGRatio),
					Detalhe: "Preço Justo para Crescimento",
				}
			} else {
				ativo.Pareceres["Lynch"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Teto: R$ %6.2f | PEG: %4.2f", ativo.PrecoJustoLynch, ativo.PEGRatio),
					Detalhe: "Esticado para o Crescimento",
				}
			}
		}

		// Parecer 4: Gordon Growth Model (DDM)
		if ativo.PrecoTetoGordon > 0 {
			if ativo.PrecoAtual <= ativo.PrecoTetoGordon {
				ativo.Pareceres["Gordon"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Teto Gordon: R$ %6.2f", ativo.PrecoTetoGordon),
					Detalhe: "Fluxo Futuro Sustentável",
				}
			} else {
				ativo.Pareceres["Gordon"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("Teto Gordon: R$ %6.2f", ativo.PrecoTetoGordon),
					Detalhe: "Acima do Teto de Gordon",
				}
			}
		}

		// Parecer 5: Joel Greenblatt (Earnings Yield)
		if ativo.EarningsYield > 0 {
			if ativo.EarningsYield >= 10.0 {
				ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
					Status:  domain.ParecerAprovado,
					Metrica: fmt.Sprintf("Earnings Yield: %4.1f%%", ativo.EarningsYield),
					Detalhe: "Lucro Supera Renda Fixa",
				}
			} else if ativo.EarningsYield >= 6.0 {
				ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
					Status:  domain.ParecerAtencao,
					Metrica: fmt.Sprintf("Earnings Yield: %4.1f%%", ativo.EarningsYield),
					Detalhe: "Retorno Operacional Moderado",
				}
			} else {
				ativo.Pareceres["Greenblatt"] = domain.ParecerItem{
					Status:  domain.ParecerReprovado,
					Metrica: fmt.Sprintf("Earnings Yield: %4.1f%%", ativo.EarningsYield),
					Detalhe: "Earnings Yield Baixo",
				}
			}
		}

		// CÁLCULO DO SCORE GERAL DE AÇÕES (0 a 100)
		score := 0.0

		// Bloco 1: Bazin (até 30 pontos)
		if ativo.DY >= 8.0 {
			score += 30.0
		} else if ativo.DY >= 6.0 {
			score += 22.0 + (ativo.DY-6.0)*4.0
		} else if ativo.DY >= 4.0 {
			score += 12.0
		} else if ativo.DY > 0.0 {
			score += 5.0
		}

		// Bloco 2: Graham (até 30 pontos)
		if ativo.MargemGraham >= 20.0 {
			score += 30.0
		} else if ativo.MargemGraham >= 0.0 {
			score += 22.0 + (ativo.MargemGraham/20.0)*8.0
		} else if ativo.MargemGraham >= -15.0 {
			score += 10.0
		}

		// Bloco 3: Lynch PEG (até 15 pontos)
		if ativo.PEGRatio > 0 {
			if ativo.PEGRatio <= 0.8 {
				score += 15.0
			} else if ativo.PEGRatio <= 1.0 {
				score += 12.0
			} else if ativo.PEGRatio <= 1.5 {
				score += 6.0
			}
		}

		// Bloco 4: Gordon (até 15 pontos)
		if ativo.PrecoTetoGordon > 0 && ativo.PrecoAtual <= ativo.PrecoTetoGordon {
			score += 15.0
		} else if ativo.PrecoTetoGordon > 0 && ativo.PrecoAtual <= ativo.PrecoTetoGordon*1.15 {
			score += 7.0
		}

		// Bloco 5: ROE / Rentabilidade (até 10 pontos)
		if ativo.ROE >= 15.0 {
			score += 10.0
		} else if ativo.ROE >= 10.0 {
			score += 6.0
		} else if ativo.ROE > 0.0 {
			score += 3.0
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
		if ativo.SpreadNTNB >= 2.0 {
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
		if ativo.PrecoTetoBazin > 0 {
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

		// CÁLCULO DO SCORE DE FIIs (0 a 100)
		score := 0.0

		// Bloco 1: Desconto P/VP (até 40 pontos)
		if ativo.PVP > 0 {
			if ativo.PVP >= 0.85 && ativo.PVP <= 0.95 {
				score += 40.0 // Desconto ideal com margem
			} else if ativo.PVP > 0.95 && ativo.PVP <= 1.02 {
				score += 32.0 // Preço justo
			} else if ativo.PVP < 0.85 {
				score += 25.0 // Desconto profundo (cautela)
			} else if ativo.PVP <= 1.06 {
				score += 10.0
			}
		}

		// Bloco 2: Spread NTN-B (até 40 pontos)
		if ativo.SpreadNTNB >= 3.0 {
			score += 40.0
		} else if ativo.SpreadNTNB >= 1.0 {
			score += 30.0
		} else if ativo.SpreadNTNB >= 0.0 {
			score += 20.0
		}

		// Bloco 3: Yield Absoluto (até 20 pontos)
		if ativo.DY >= 11.0 {
			score += 20.0
		} else if ativo.DY >= 9.0 {
			score += 15.0
		} else if ativo.DY >= 7.0 {
			score += 10.0
		} else if ativo.DY > 0.0 {
			score += 5.0
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

	// 2. Alinha o Status do Semáforo ao Score Consolidado
	if ativo.Score >= 70.0 {
		ativo.Status = domain.StatusComprarMais
	} else if ativo.Score >= 50.0 {
		ativo.Status = domain.StatusManter
	} else {
		ativo.Status = domain.StatusAlerta
	}

	return nil
}
