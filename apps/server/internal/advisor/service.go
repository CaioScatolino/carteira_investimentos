package advisor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/macro"
	"carteira_investimentos/server/internal/portfolio"
	"carteira_investimentos/server/internal/storage"
)

// PerfilInvestidor define a estratégia prioritária do investidor para o aporte
type PerfilInvestidor string

const (
	PerfilEquilibrado   PerfilInvestidor = "EQUILIBRADO"
	PerfilDividendos    PerfilInvestidor = "DIVIDENDOS"
	PerfilOportunidade  PerfilInvestidor = "OPORTUNIDADE"
	PerfilDefensivo     PerfilInvestidor = "DEFENSIVO"
)

// RequisicaoAporte carrega as opções informadas pelo investidor para o conselheiro
type RequisicaoAporte struct {
	ValorAporte         float64          `json:"valor_aporte"`
	Perfil              PerfilInvestidor `json:"perfil"`
	PermitirNovosAtivos bool             `json:"permitir_novos_ativos"`
}

// AssimetriaDetectada descreve uma distorção de preço/valor identificada pelo motor
type AssimetriaDetectada struct {
	Ticker             string                `json:"ticker"`
	Nome               string                `json:"nome"`
	Classe             portfolio.ClasseAtivo `json:"classe"`
	CotacaoAtual       float64               `json:"cotacao_atual"`
	PrecoTeto          float64               `json:"preco_teto"`
	MargemSegurancaPct float64               `json:"margem_seguranca_pct"`
	Score              int                   `json:"score"`
	DividendYield      float64               `json:"dividend_yield"`
	Motivo             string                `json:"motivo"`
}

// DiagnosticoAdvisor agrega o diagnóstico matemático e estratégico completo
type DiagnosticoAdvisor struct {
	ValorAporteInformado float64                             `json:"valor_aporte_informado"`
	ValorAporteEfetivo   float64                             `json:"valor_aporte_efetivo"`
	SaldoRestanteCaixa   float64                             `json:"saldo_restante_caixa"`
	PatrimonioAtual      float64                             `json:"patrimonio_atual"`
	PatrimonioProjetado  float64                             `json:"patrimonio_projetado"`
	ClasseMaisDefasada   portfolio.ClasseAtivo               `json:"classe_mais_defasada"`
	DeltaDefasagem       float64                             `json:"delta_defasagem"`
	CenarioMacro         macro.CenarioMacro                  `json:"cenario_macro"`
	Assimetrias          []AssimetriaDetectada               `json:"assimetrias"`
	Sugestoes            []portfolio.SugestaoAporteItem      `json:"sugestoes"`
	DiagnosticoMacro     []portfolio.DiagnosticoMacroClasse  `json:"diagnostico_macro"`
	ParecerSintese       string                              `json:"parecer_sintese,omitempty"`
}

// Service orquestra as decisões matemáticas e a geração do parecer cognitivo
type Service struct {
	portfolioService *portfolio.Service
	store            storage.SnapshotStore
	geminiAPIKey     string
	httpClient       *http.Client
}

// NovoService instancia o serviço do Assistente Cognitivo de Aportes
func NovoService(ps *portfolio.Service, store storage.SnapshotStore, geminiKey string) *Service {
	return &Service{
		portfolioService: ps,
		store:            store,
		geminiAPIKey:     strings.TrimSpace(geminiKey),
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// AnalisarAporte realiza o diagnóstico matemático completo sem acionar a LLM
func (s *Service) AnalisarAporte(ctx context.Context, usuarioID int64, req RequisicaoAporte) (*DiagnosticoAdvisor, error) {
	if req.ValorAporte <= 0 {
		return nil, fmt.Errorf("o valor do aporte deve ser maior que zero")
	}
	if req.Perfil == "" {
		req.Perfil = PerfilEquilibrado
	}

	var resumo *portfolio.ResumoCarteira
	if s.portfolioService != nil {
		var err error
		resumo, err = s.portfolioService.ObterResumoCarteira(ctx, usuarioID)
		if err != nil {
			return nil, fmt.Errorf("falha ao carregar carteira do usuário: %w", err)
		}
	}
	if resumo == nil {
		resumo = &portfolio.ResumoCarteira{
			AlocacaoAlvo: map[portfolio.ClasseAtivo]float64{
				portfolio.ClasseRendaFixa:     40.0,
				portfolio.ClasseFII:           22.5,
				portfolio.ClasseAcaoBR:        21.0,
				portfolio.ClasseInternacional: 16.0,
				portfolio.ClasseCaixa:         0.5,
			},
			AlocacaoAtual: make(map[portfolio.ClasseAtivo]float64),
		}
	}

	cenarioMacro := macro.ObterCenario()
	todosAtivos := s.store.ListarTodos()
	mapaMercado := make(map[string]*domain.Ativo, len(todosAtivos))
	for _, a := range todosAtivos {
		mapaMercado[a.Ticker] = a
	}

	// 1. Simulação Macro Inicial
	simulacao := portfolio.SimularRebalanceamento(resumo, req.ValorAporte)

	// 2. Identificação da Classe Mais Defasada
	classeMaisDefasada := portfolio.ClasseRendaFixa
	maiorDeficit := 0.0
	for _, d := range simulacao.DiagnosticoMacro {
		if d.DeltaPercentual < maiorDeficit {
			maiorDeficit = d.DeltaPercentual
			classeMaisDefasada = d.Classe
		}
	}

	// 3. Radar de Assimetrias Brutais de Mercado
	assimetrias := s.detectarAssimetrias(resumo, mapaMercado, req.PermitirNovosAtivos)

	// 4. Refinamento de Sugestões Conforme Perfil Escolhido
	sugestoesFinais, valorGasto := s.refinarSugestoes(req, simulacao, assimetrias, mapaMercado, resumo)

	diagnostico := &DiagnosticoAdvisor{
		ValorAporteInformado: req.ValorAporte,
		ValorAporteEfetivo:   valorGasto,
		SaldoRestanteCaixa:   math.Max(0, req.ValorAporte-valorGasto),
		PatrimonioAtual:      resumo.PatrimonioTotal,
		PatrimonioProjetado:  resumo.PatrimonioTotal + valorGasto,
		ClasseMaisDefasada:   classeMaisDefasada,
		DeltaDefasagem:       maiorDeficit,
		CenarioMacro:         cenarioMacro,
		Assimetrias:          assimetrias,
		Sugestoes:            sugestoesFinais,
		DiagnosticoMacro:     simulacao.DiagnosticoMacro,
	}

	diagnostico.ParecerSintese = s.gerarParecerDeterministo(diagnostico, req.Perfil)

	return diagnostico, nil
}

// detectarAssimetrias garimpa oportunidades com descontos severos e fundamentos sólidos
func (s *Service) detectarAssimetrias(resumo *portfolio.ResumoCarteira, mapaMercado map[string]*domain.Ativo, permitirNovos bool) []AssimetriaDetectada {
	var assimetrias []AssimetriaDetectada
	jaVistos := make(map[string]bool)

	// 1. Analisa as posições que o investidor já possui
	for _, p := range resumo.Posicoes {
		jaVistos[p.Ticker] = true
		ativoMercado := mapaMercado[p.Ticker]
		if ativoMercado == nil {
			continue
		}

		if ativoMercado.Classe == domain.ClasseAcao {
			if ativoMercado.MargemSegurancaConsolidada >= 20.0 && ativoMercado.Score >= 80 && ativoMercado.PrecoAtual < ativoMercado.PrecoTetoConsolidado {
				assimetrias = append(assimetrias, AssimetriaDetectada{
					Ticker:             p.Ticker,
					Nome:               ativoMercado.Nome,
					Classe:             p.Classe,
					CotacaoAtual:       ativoMercado.PrecoAtual,
					PrecoTeto:          ativoMercado.PrecoTetoConsolidado,
					MargemSegurancaPct: ativoMercado.MargemSegurancaConsolidada,
					Score:              int(math.Round(ativoMercado.Score)),
					DividendYield:      ativoMercado.DY,
					Motivo:             fmt.Sprintf("Ativo em custódia com Margem de Segurança de %.1f%% e Score %.0f", ativoMercado.MargemSegurancaConsolidada, ativoMercado.Score),
				})
			}
		} else if ativoMercado.Classe == domain.ClasseFII {
			if ativoMercado.PVPReal > 0 && ativoMercado.PVPReal <= 0.92 && ativoMercado.DY >= 8.5 && ativoMercado.Score >= 75 {
				assimetrias = append(assimetrias, AssimetriaDetectada{
					Ticker:             p.Ticker,
					Nome:               ativoMercado.Nome,
					Classe:             p.Classe,
					CotacaoAtual:       ativoMercado.PrecoAtual,
					PrecoTeto:          ativoMercado.PrecoTetoConsolidado,
					MargemSegurancaPct: ativoMercado.MargemSegurancaConsolidada,
					Score:              int(math.Round(ativoMercado.Score)),
					DividendYield:      ativoMercado.DY,
					Motivo:             fmt.Sprintf("FII descontado negociando a P/VP %.2fx com DY %.1f%%", ativoMercado.PVPReal, ativoMercado.DY),
				})
			}
		}
	}

	// 2. Se permitido sugerir novos ativos, garimpa os líderes do ranking geral B3
	if permitirNovos {
		for _, a := range mapaMercado {
			if jaVistos[a.Ticker] {
				continue
			}
			if a.VereditoRisco == "RISCO_DESCOMPENSADO" {
				continue
			}

			if a.Classe == domain.ClasseAcao && a.Score >= 85 && a.MargemSegurancaConsolidada >= 25.0 && a.Porte == "BLUE_CHIP" {
				assimetrias = append(assimetrias, AssimetriaDetectada{
					Ticker:             a.Ticker,
					Nome:               a.Nome,
					Classe:             portfolio.ClasseAcaoBR,
					CotacaoAtual:       a.PrecoAtual,
					PrecoTeto:          a.PrecoTetoConsolidado,
					MargemSegurancaPct: a.MargemSegurancaConsolidada,
					Score:              int(math.Round(a.Score)),
					DividendYield:      a.DY,
					Motivo:             fmt.Sprintf("Blue Chip líder da B3 com Margem de Segurança de %.1f%% e Score %.0f", a.MargemSegurancaConsolidada, a.Score),
				})
				jaVistos[a.Ticker] = true
			} else if a.Classe == domain.ClasseFII && a.Score >= 85 && a.PVPReal > 0 && a.PVPReal <= 0.88 && a.DY >= 9.0 {
				assimetrias = append(assimetrias, AssimetriaDetectada{
					Ticker:             a.Ticker,
					Nome:               a.Nome,
					Classe:             portfolio.ClasseFII,
					CotacaoAtual:       a.PrecoAtual,
					PrecoTeto:          a.PrecoTetoConsolidado,
					MargemSegurancaPct: a.MargemSegurancaConsolidada,
					Score:              int(math.Round(a.Score)),
					DividendYield:      a.DY,
					Motivo:             fmt.Sprintf("FII Top 10 com P/VP %.2fx e Dividend Yield anual de %.1f%%", a.PVPReal, a.DY),
				})
				jaVistos[a.Ticker] = true
			}

			if len(assimetrias) >= 8 {
				break
			}
		}
	}

	// Ordena pela maior margem de segurança
	sort.Slice(assimetrias, func(i, j int) bool {
		return assimetrias[i].MargemSegurancaPct > assimetrias[j].MargemSegurancaPct
	})

	if len(assimetrias) > 5 {
		assimetrias = assimetrias[:5]
	}

	return assimetrias
}

// refinarSugestoes balanceia as recomendações matemáticas considerando o perfil do investidor
func (s *Service) refinarSugestoes(
	req RequisicaoAporte,
	simulacao *portfolio.ResultadoRebalanceamento,
	assimetrias []AssimetriaDetectada,
	mapaMercado map[string]*domain.Ativo,
	resumo *portfolio.ResumoCarteira,
) ([]portfolio.SugestaoAporteItem, float64) {
	// Se o perfil for estritamente equilibrado e já gerou sugestões sólidas, mantém
	if req.Perfil == PerfilEquilibrado && len(simulacao.Sugestoes) > 0 {
		return simulacao.Sugestoes, simulacao.ValorAporteEfetivo
	}

	saldo := req.ValorAporte
	var resultado []portfolio.SugestaoAporteItem
	valorGasto := 0.0

	// Caso OPORTUNIDADE ou DIVIDENDOS: se houver assimetrias de alto calibre, destina ao menos uma cota
	if (req.Perfil == PerfilOportunidade || req.Perfil == PerfilDividendos) && len(assimetrias) > 0 {
		topAssimetria := assimetrias[0]
		preco := topAssimetria.CotacaoAtual
		if preco > 0 && saldo >= preco {
			// Reserva até 60% do saldo para a assimetria
			fatia := saldo * 0.60
			qtd := math.Floor(fatia / preco)
			if qtd < 1 {
				qtd = 1
			}
			custo := qtd * preco

			resultado = append(resultado, portfolio.SugestaoAporteItem{
				Ticker:             topAssimetria.Ticker,
				Classe:             topAssimetria.Classe,
				QuantidadeSugerida: qtd,
				PrecoEstimado:      preco,
				ValorTotalSugerido: custo,
				Score:              topAssimetria.Score,
				MargemSegurancaPct: topAssimetria.MargemSegurancaPct,
				Justificativa:      fmt.Sprintf("🎯 Assimetria de Alta Relevância: %s", topAssimetria.Motivo),
			})

			saldo -= custo
			valorGasto += custo
		}
	}

	// O restante do saldo é direcionado pelo motor de rebalanceamento da carteira
	if saldo > 0 && len(simulacao.Sugestoes) > 0 {
		for _, sug := range simulacao.Sugestoes {
			if saldo <= 0 {
				break
			}
			// Não duplica ticker já sugerido
			jaAdicionado := false
			for _, r := range resultado {
				if r.Ticker == sug.Ticker {
					jaAdicionado = true
					break
				}
			}
			if jaAdicionado {
				continue
			}

			if sug.Classe == portfolio.ClasseRendaFixa || sug.Classe == portfolio.ClasseCaixa {
				fatia := math.Min(saldo, sug.ValorTotalSugerido)
				if fatia > 0 {
					qtd := fatia / sug.PrecoEstimado
					resultado = append(resultado, portfolio.SugestaoAporteItem{
						Ticker:             sug.Ticker,
						Classe:             sug.Classe,
						QuantidadeSugerida: math.Round(qtd*100) / 100,
						PrecoEstimado:      sug.PrecoEstimado,
						ValorTotalSugerido: fatia,
						Score:              sug.Score,
						MargemSegurancaPct: sug.MargemSegurancaPct,
						Justificativa:      sug.Justificativa,
					})
					saldo -= fatia
					valorGasto += fatia
				}
			} else {
				if sug.PrecoEstimado > 0 && saldo >= sug.PrecoEstimado {
					fatia := math.Min(saldo, sug.ValorTotalSugerido)
					qtd := math.Floor(fatia / sug.PrecoEstimado)
					if qtd >= 1 {
						custo := qtd * sug.PrecoEstimado
						resultado = append(resultado, portfolio.SugestaoAporteItem{
							Ticker:             sug.Ticker,
							Classe:             sug.Classe,
							QuantidadeSugerida: qtd,
							PrecoEstimado:      sug.PrecoEstimado,
							ValorTotalSugerido: custo,
							Score:              sug.Score,
							MargemSegurancaPct: sug.MargemSegurancaPct,
							Justificativa:      sug.Justificativa,
						})
						saldo -= custo
						valorGasto += custo
					}
				}
			}
		}
	}

	if len(resultado) == 0 {
		return simulacao.Sugestoes, simulacao.ValorAporteEfetivo
	}

	return resultado, valorGasto
}

// gerarParecerDeterministo cria o parecer institucional e pedagógico Buy & Hold
func (s *Service) gerarParecerDeterministo(d *DiagnosticoAdvisor, perfil PerfilInvestidor) string {
	var sb strings.Builder

	sb.WriteString("### 🏛️ Estratégia Institucional de Aporte Buy & Hold\n\n")

	sb.WriteString(fmt.Sprintf("• **Cenário Macroeconômico:** Selic Meta a **%.2f%% a.a.** e NTN-B (IPCA+) a **%.2f%% a.a.** real. O Hurdle mínimo exigido para ações de dividendo é de **%.2f%% a.a.**.\n",
		d.CenarioMacro.TaxaSelic, d.CenarioMacro.TaxaNTNB, d.CenarioMacro.TaxaNTNB+d.CenarioMacro.SpreadMinimoAcoes))

	sb.WriteString(fmt.Sprintf("• **Diagnóstico de Custódia:** Patrimônio atual em **R$ %.2f**. A caixinha patrimonial com maior defasagem matemática no momento é **%s** (desvio de **%.1f%%** da meta institucional).\n\n",
		d.PatrimonioAtual, d.ClasseMaisDefasada, d.DeltaDefasagem))

	sb.WriteString("#### 🎯 Recomendações Prioritárias para o Aporte de Hoje:\n\n")

	for i, sug := range d.Sugestoes {
		sb.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, sug.Ticker, sug.Classe))
		if sug.Classe == portfolio.ClasseRendaFixa || sug.Classe == portfolio.ClasseCaixa {
			sb.WriteString(fmt.Sprintf("   - **Alocação:** R$ %.2f (Aporte em Título/CDB)\n", sug.ValorTotalSugerido))
		} else {
			sb.WriteString(fmt.Sprintf("   - **Quantidade Sugerida:** %.0f cota(s) a R$ %.2f (Total: R$ %.2f)\n",
				sug.QuantidadeSugerida, sug.PrecoEstimado, sug.ValorTotalSugerido))
		}
		if sug.MargemSegurancaPct > 0 {
			sb.WriteString(fmt.Sprintf("   - **Margem de Segurança:** +%.1f%% sobre o Preço Teto Consensual\n", sug.MargemSegurancaPct))
		}
		sb.WriteString(fmt.Sprintf("   - **Racional:** %s\n\n", sug.Justificativa))
	}

	if len(d.Assimetrias) > 0 {
		sb.WriteString("#### ⚡ Radar de Assimetrias & Distorções de Preço:\n")
		for _, a := range d.Assimetrias {
			sb.WriteString(fmt.Sprintf("- **%s**: Negociando a R$ %.2f com Preço Teto de R$ %.2f (Margem: **+%.1f%%** | Score: **%d**). %s.\n",
				a.Ticker, a.CotacaoAtual, a.PrecoTeto, a.MargemSegurancaPct, a.Score, a.Motivo))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("💵 **Resumo Financeiro:** Aporte total recomendado de **R$ %.2f** (Saldo restante para caixa/próximo aporte: **R$ %.2f**).\n\n",
		d.ValorAporteEfetivo, d.SaldoRestanteCaixa))

	sb.WriteString("> *\"No Buy & Hold, você não tenta adivinhar o fundo do mercado. Você compra empresas boas, abaixo do preço teto, e reinveste os proventos com regularidade matemática.\"*")

	return sb.String()
}

// StreamAdvisorSSE transmite o diagnóstico e a síntese da IA em streaming Server-Sent Events
func (s *Service) StreamAdvisorSSE(w http.ResponseWriter, r *http.Request, usuarioID int64, req RequisicaoAporte) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming SSE não suportado pelo servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 1. Executa a auditoria matemática
	diagnostico, err := s.AnalisarAporte(r.Context(), usuarioID, req)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// 2. Envia evento com o diagnóstico estruturado (para montagem dos cards de aporte no front)
	diagJSON, _ := json.Marshal(diagnostico)
	fmt.Fprintf(w, "event: diagnostico\ndata: %s\n\n", string(diagJSON))
	flusher.Flush()

	// 3. Se houver chave Gemini, transmite via LLM streaming; senão transmite o parecer analítico nativo
	if s.geminiAPIKey != "" {
		s.transmitirGeminiStreaming(w, flusher, r.Context(), diagnostico, req)
	} else {
		s.transmitirTextoEmStreaming(w, flusher, diagnostico.ParecerSintese)
	}

	// 4. Sinaliza fim da transmissão
	fmt.Fprintf(w, "event: done\ndata: [DONE]\n\n")
	flusher.Flush()
}

// transmitirTextoEmStreaming simula a digitação token a token para parecer nativo
func (s *Service) transmitirTextoEmStreaming(w http.ResponseWriter, flusher http.Flusher, texto string) {
	palavras := strings.Split(texto, " ")
	for _, p := range palavras {
		fmt.Fprintf(w, "event: token\ndata: %s \n\n", p)
		flusher.Flush()
		time.Sleep(20 * time.Millisecond)
	}
}

// transmitirGeminiStreaming conecta à API do Google Gemini e transmite chunks SSE em tempo real
func (s *Service) transmitirGeminiStreaming(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, d *DiagnosticoAdvisor, req RequisicaoAporte) {
	prompt := fmt.Sprintf(`Você é o Estrategista-Chefe Institucional da plataforma B3 Core, especialista em Valuation Fundamentalista (Métodos Décio Bazin, Benjamin Graham e Peter Lynch) e Filosofia Buy & Hold.
Analise a carteira do investidor e o aporte planejado com base nos dados REAIS e AUDITADOS a seguir:

DADOS MACROECONÔMICOS:
- Selic Meta: %.2f%% a.a.
- NTN-B IPCA+: %.2f%% a.a. real
- Hurdle Mínimo Ações: %.2f%% a.a.

CARTEIRA DO INVESTIDOR:
- Patrimônio Atual: R$ %.2f
- Valor do Aporte Disponível: R$ %.2f
- Perfil Selecionado: %s
- Classe Mais Defasada: %s (Déficit de %.1f%%)

SUGESTÕES MATEMÁTICAS PRÉ-CALCULADAS:
%s

ASSIMETRIAS DE MERCADO:
%s

INSTRUÇÕES DE RESPOSTA:
1. Escreva um parecer conciso, elegante e altamente embasado.
2. Destaque por que a disciplina de recompor a classe defasada e aproveitar assimetrias comprovadas traz retornos assimétricos no longo prazo.
3. Não invente números. Use exatamente os valores, cotas e preços sugeridos acima.
4. Conclua com uma recomendação prática de execução Buy & Hold.`,
		d.CenarioMacro.TaxaSelic, d.CenarioMacro.TaxaNTNB, d.CenarioMacro.TaxaNTNB+d.CenarioMacro.SpreadMinimoAcoes,
		d.PatrimonioAtual, d.ValorAporteInformado, req.Perfil, d.ClasseMaisDefasada, d.DeltaDefasagem,
		formatarSugestoes(d.Sugestoes), formatarAssimetrias(d.Assimetrias),
	)

	payload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":     0.2,
			"topP":            0.8,
			"maxOutputTokens": 1024,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		s.transmitirTextoEmStreaming(w, flusher, d.ParecerSintese)
		return
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:streamGenerateContent?key=%s", s.geminiAPIKey)
	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		s.transmitirTextoEmStreaming(w, flusher, d.ParecerSintese)
		return
	}
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(reqHTTP)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Fallback gracioso para o parecer nativo
		s.transmitirTextoEmStreaming(w, flusher, d.ParecerSintese)
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			lineStr := strings.TrimSpace(string(line))
			if strings.HasPrefix(lineStr, "\"text\":") {
				// Extrai o texto gerado
				partes := strings.SplitN(lineStr, "\"text\":", 2)
				if len(partes) == 2 {
					var textoToken string
					textoLimpo := strings.Trim(partes[1], " \",")
					json.Unmarshal([]byte("\""+textoLimpo+"\""), &textoToken)
					if textoToken != "" {
						fmt.Fprintf(w, "event: token\ndata: %s\n\n", textoToken)
						flusher.Flush()
					}
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				// Qualquer erro encerra
			}
			break
		}
	}
}

func formatarSugestoes(sugestoes []portfolio.SugestaoAporteItem) string {
	var sb strings.Builder
	for _, s := range sugestoes {
		sb.WriteString(fmt.Sprintf("- %s (%s): %.0f cotas a R$ %.2f (Total: R$ %.2f | Margem: +%.1f%% | Score: %d)\n",
			s.Ticker, s.Classe, s.QuantidadeSugerida, s.PrecoEstimado, s.ValorTotalSugerido, s.MargemSegurancaPct, s.Score))
	}
	return sb.String()
}

func formatarAssimetrias(assimetrias []AssimetriaDetectada) string {
	var sb strings.Builder
	for _, a := range assimetrias {
		sb.WriteString(fmt.Sprintf("- %s: Preço R$ %.2f vs Teto R$ %.2f (Margem: +%.1f%% | Score: %d). %s\n",
			a.Ticker, a.CotacaoAtual, a.PrecoTeto, a.MargemSegurancaPct, a.Score, a.Motivo))
	}
	return sb.String()
}
