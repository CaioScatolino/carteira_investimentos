package portfolio

import (
	"math"
	"sort"
)

// SugestaoAporteItem detalha a alocação recomendada para um ativo específico.
type SugestaoAporteItem struct {
	Ticker             string      `json:"ticker"`
	Classe             ClasseAtivo `json:"classe"`
	QuantidadeSugerida float64     `json:"quantidade_sugerida"`
	PrecoEstimado      float64     `json:"preco_estimado"`
	ValorTotalSugerido float64     `json:"valor_total_sugerido"`
	Score              int         `json:"score"`
	MargemSegurancaPct float64     `json:"margem_seguranca_pct"`
	Justificativa      string      `json:"justificativa"`
}

// DiagnosticoMacroClasse resume o status de equilíbrio de cada caixinha patrimonial.
type DiagnosticoMacroClasse struct {
	Classe            ClasseAtivo `json:"classe"`
	PercentualAtual   float64     `json:"percentual_atual"`
	PercentualAlvo    float64     `json:"percentual_alvo"`
	DeltaPercentual   float64     `json:"delta_percentual"` // Atual - Alvo
	ValorAtual        float64     `json:"valor_atual"`
	ValorIdeal        float64     `json:"valor_ideal"`
	NecessidadeAporte float64     `json:"necessidade_aporte"`
	Status            string      `json:"status"` // 'DEFASADO_APORTAR', 'EQUILIBRADO', 'SOBREALOCADO'
}

// ResultadoRebalanceamento consolida a simulação do aporte para o investidor.
type ResultadoRebalanceamento struct {
	ValorAporteInformado float64                  `json:"valor_aporte_informado"`
	ValorAporteEfetivo   float64                  `json:"valor_aporte_efetivo"`
	SaldoRestanteCaixa   float64                  `json:"saldo_restante_caixa"`
	DiagnosticoMacro     []DiagnosticoMacroClasse `json:"diagnostico_macro"`
	Sugestoes            []SugestaoAporteItem     `json:"sugestoes"`
}

// SimularRebalanceamento executa o algoritmo hierárquico em 2 níveis para distribuir o aporte.
func SimularRebalanceamento(resumo *ResumoCarteira, valorAporte float64) *ResultadoRebalanceamento {
	if valorAporte <= 0 {
		valorAporte = 0
	}

	patrimonioFuturo := resumo.PatrimonioTotal + valorAporte

	// 1. Diagnóstico do Nível 1: Macro Alocação
	classes := []ClasseAtivo{ClasseRendaFixa, ClasseFII, ClasseAcaoBR, ClasseInternacional, ClasseCaixa}
	diagnosticos := make([]DiagnosticoMacroClasse, 0, len(classes))

	totalNecessidade := 0.0

	for _, classe := range classes {
		alvoPct := resumo.AlocacaoAlvo[classe]
		atualPct := resumo.AlocacaoAtual[classe]
		deltaPct := atualPct - alvoPct

		valorAtual := (atualPct / 100.0) * resumo.PatrimonioTotal
		valorIdeal := (alvoPct / 100.0) * patrimonioFuturo

		necessidade := math.Max(0, valorIdeal-valorAtual)
		totalNecessidade += necessidade

		status := "EQUILIBRADO"
		if deltaPct < -2.0 {
			status = "DEFASADO_APORTAR"
		} else if deltaPct > 3.0 {
			status = "SOBREALOCADO"
		}

		diagnosticos = append(diagnosticos, DiagnosticoMacroClasse{
			Classe:            classe,
			PercentualAtual:   atualPct,
			PercentualAlvo:    alvoPct,
			DeltaPercentual:   deltaPct,
			ValorAtual:        valorAtual,
			ValorIdeal:        valorIdeal,
			NecessidadeAporte: necessidade,
			Status:            status,
		})
	}

	// Ordena as classes da mais atrasada (maior necessidade relativa) para a mais sobrealocada
	sort.Slice(diagnosticos, func(i, j int) bool {
		return diagnosticos[i].DeltaPercentual < diagnosticos[j].DeltaPercentual
	})

	// 2. Nível 2: Micro Alocação Fundamentalista
	// Agrupa as posições existentes por classe e ordena pelo melhor Score + Margem de Segurança
	posicoesPorClasse := make(map[ClasseAtivo][]PosicaoConsolidada)
	for _, p := range resumo.Posicoes {
		posicoesPorClasse[p.Classe] = append(posicoesPorClasse[p.Classe], p)
	}

	for _, lista := range posicoesPorClasse {
		sort.Slice(lista, func(i, j int) bool {
			// Prioriza primeiro quem tem maior Margem de Segurança sobre o Preço Teto
			if lista[i].MargemSegurancaPct != lista[j].MargemSegurancaPct {
				return lista[i].MargemSegurancaPct > lista[j].MargemSegurancaPct
			}
			return lista[i].ScoreFundamentalista > lista[j].ScoreFundamentalista
		})
	}

	saldoDisponivel := valorAporte
	var sugestoes []SugestaoAporteItem
	valorEfetivo := 0.0

	// Distribui o dinheiro proporcionalmente para as classes que estão defasadas
	for _, diag := range diagnosticos {
		if diag.NecessidadeAporte <= 0 || saldoDisponivel <= 0 {
			continue
		}

		fatiaClasse := math.Min(saldoDisponivel, diag.NecessidadeAporte)
		candidatos := posicoesPorClasse[diag.Classe]

		if len(candidatos) > 0 {
			melhorAtivo := candidatos[0]
			precoCotacao := melhorAtivo.CotacaoAtual
			if precoCotacao <= 0 {
				precoCotacao = melhorAtivo.PrecoMedio
			}

			// Renda Fixa ou Caixa: permite aportes fracionados (ex: 0.1 de Tesouro Direto ou aportes livres em CDB)
			if diag.Classe == ClasseRendaFixa || diag.Classe == ClasseCaixa {
				if fatiaClasse > 0 {
					qtd := fatiaClasse / precoCotacao
					justificativa := "Aporte direcionado para restabelecer a meta de 40% em Renda Fixa/Segurança."

					sugestoes = append(sugestoes, SugestaoAporteItem{
						Ticker:             melhorAtivo.Ticker,
						Classe:             diag.Classe,
						QuantidadeSugerida: math.Round(qtd*100) / 100, // arredonda para 2 casas decimais
						PrecoEstimado:      precoCotacao,
						ValorTotalSugerido: fatiaClasse,
						Score:              melhorAtivo.ScoreFundamentalista,
						MargemSegurancaPct: melhorAtivo.MargemSegurancaPct,
						Justificativa:      justificativa,
					})

					saldoDisponivel -= fatiaClasse
					valorEfetivo += fatiaClasse
				}
			} else {
				// Renda Variável (Ações / FIIs): compra em cotas inteiras
				if precoCotacao > 0 && fatiaClasse >= precoCotacao {
					qtd := math.Floor(fatiaClasse / precoCotacao)
					custoTotal := qtd * precoCotacao

					if qtd > 0 {
						justificativa := "Classe defasada em relação à meta macro."
						if melhorAtivo.MargemSegurancaPct > 0 {
							justificativa += " Ativo com margem de segurança positiva e excelente score fundamentalista."
						}

						sugestoes = append(sugestoes, SugestaoAporteItem{
							Ticker:             melhorAtivo.Ticker,
							Classe:             diag.Classe,
							QuantidadeSugerida: qtd,
							PrecoEstimado:      precoCotacao,
							ValorTotalSugerido: custoTotal,
							Score:              melhorAtivo.ScoreFundamentalista,
							MargemSegurancaPct: melhorAtivo.MargemSegurancaPct,
							Justificativa:      justificativa,
						})

						saldoDisponivel -= custoTotal
						valorEfetivo += custoTotal
					}
				}
			}
		} else {
			// Classe sem ativos em custódia (ex: Internacional 0%): sugere iniciar alocação
			tickerSugerido := "NOVA_POSICAO_" + string(diag.Classe)
			if diag.Classe == ClasseInternacional {
				tickerSugerido = "IVVB11 / WRLD11"
			}
			sugestoes = append(sugestoes, SugestaoAporteItem{
				Ticker:             tickerSugerido,
				Classe:             diag.Classe,
				QuantidadeSugerida: 1,
				PrecoEstimado:      fatiaClasse,
				ValorTotalSugerido: fatiaClasse,
				Score:              100,
				MargemSegurancaPct: 0,
				Justificativa:      "Classe 100% defasada. Recomendado cadastrar ativo de referência global para proteção cambial.",
			})

			saldoDisponivel -= fatiaClasse
			valorEfetivo += fatiaClasse
		}
	}

	return &ResultadoRebalanceamento{
		ValorAporteInformado: valorAporte,
		ValorAporteEfetivo:   valorEfetivo,
		SaldoRestanteCaixa:   saldoDisponivel,
		DiagnosticoMacro:     diagnosticos,
		Sugestoes:            sugestoes,
	}
}
