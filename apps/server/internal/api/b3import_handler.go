package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"carteira_investimentos/server/internal/auth"
	"carteira_investimentos/server/internal/b3import"
	"carteira_investimentos/server/internal/portfolio"
)

// PreviewImportacaoB3 recebe o upload do arquivo XLSX e retorna o preview sem gravar no banco
func (h *Handler) PreviewImportacaoB3(w http.ResponseWriter, r *http.Request) {
	if h.b3ImportService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de importação B3 indisponível",
		})
		return
	}

	_, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{
			"erro": "sessão não autenticada",
		})
		return
	}

	// Limite de 25 MB para upload de arquivo
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "falha ao processar formulário multipart: " + err.Error(),
		})
		return
	}

	file, header, err := r.FormFile("arquivo")
	if err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "o campo 'arquivo' é obrigatório no envio",
		})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "o arquivo deve ser uma planilha Excel (.xlsx) oficial da Área do Investidor B3",
		})
		return
	}

	relatorio, err := h.b3ImportService.AnalisarPlanilha(file)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"erro": "falha ao interpretar planilha da B3: " + err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, relatorio)
}

// ConfirmarImportacaoB3 persiste as posições e histórico aprovados pelo usuário
func (h *Handler) ConfirmarImportacaoB3(w http.ResponseWriter, r *http.Request) {
	if h.b3ImportService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de importação B3 indisponível",
		})
		return
	}

	claims, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{
			"erro": "sessão não autenticada",
		})
		return
	}

	var req b3import.ConfirmarImportacaoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "corpo JSON inválido: " + err.Error(),
		})
		return
	}

	if len(req.Posicoes) == 0 && len(req.Operacoes) == 0 {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "a requisição não contém posições ou operações para importar",
		})
		return
	}

	res, err := h.b3ImportService.SalvarImportacao(r.Context(), claims.UsuarioID, req)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{
			"erro": "falha ao persistir importação no banco: " + err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, res)
}

// ObterHistoricoProventos retorna o resumo e histórico de proventos recebidos na conta
func (h *Handler) ObterHistoricoProventos(w http.ResponseWriter, r *http.Request) {
	if h.b3ImportService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de proventos indisponível",
		})
		return
	}

	claims, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{
			"erro": "sessão não autenticada",
		})
		return
	}

	resumo, err := h.b3ImportService.ObterHistoricoProventos(r.Context(), claims.UsuarioID)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{
			"erro": "falha ao obter histórico de proventos: " + err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, resumo)
}

// parseDataFlexivel tenta converter diferentes formatos usuais de data retornados pela B3/StatusInvest
func parseDataFlexivel(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" || strings.EqualFold(s, "A definir") {
		return time.Time{}, false
	}
	for _, layout := range []string{"02/01/2006", "2006-01-02", "02/01/06", "02-01-2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// ObterRadarFuturoProventos cruza as cotas em custódia com proventos já anunciados
func (h *Handler) ObterRadarFuturoProventos(w http.ResponseWriter, r *http.Request) {
	if h.portfolioService == nil || h.b3Client == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de custódia ou provedor B3 indisponível",
		})
		return
	}

	claims, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{
			"erro": "sessão não autenticada",
		})
		return
	}

	resumoCarteira, err := h.portfolioService.ObterResumoCarteira(r.Context(), claims.UsuarioID)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{
			"erro": "falha ao consultar posições da carteira: " + err.Error(),
		})
		return
	}

	// 1. Agrupa posições da custódia por ticker (somando quantidades caso tenha em mais de uma corretora)
	// Filtra Renda Fixa e Caixa para não disparar consultas desnecessárias
	type itemCustodia struct {
		Ticker     string
		Quantidade float64
		IsFII      bool
	}
	mapaItens := make(map[string]*itemCustodia)

	for _, item := range resumoCarteira.Posicoes {
		if item.Quantidade <= 0 {
			continue
		}
		if item.Classe == portfolio.ClasseRendaFixa || item.Classe == portfolio.ClasseCaixa {
			continue
		}

		tk := strings.ToUpper(strings.TrimSpace(item.Ticker))
		if strings.HasPrefix(tk, "TESOURO") || strings.HasPrefix(tk, "CDB") || strings.HasPrefix(tk, "LCI") || strings.HasPrefix(tk, "LCA") || tk == "CAIXA" {
			continue
		}

		isFII := item.Classe == portfolio.ClasseFII || strings.HasSuffix(tk, "11")
		if pos, exists := mapaItens[tk]; exists {
			pos.Quantidade += item.Quantidade
		} else {
			mapaItens[tk] = &itemCustodia{
				Ticker:     tk,
				Quantidade: item.Quantidade,
				IsFII:      isFII,
			}
		}
	}

	// 2. Processamento concorrente ultra-rápido via Worker Pool
	itens := make([]*itemCustodia, 0, len(mapaItens))
	for _, it := range mapaItens {
		itens = append(itens, it)
	}

	radarFuturo := make([]b3import.ProventoFuturo, 0)
	if len(itens) == 0 {
		responderJSON(w, http.StatusOK, map[string]interface{}{
			"total_previsto": 0,
			"proventos":      radarFuturo,
		})
		return
	}

	workerCount := 8
	if len(itens) < workerCount {
		workerCount = len(itens)
	}

	jobs := make(chan *itemCustodia, len(itens))
	for _, it := range itens {
		jobs <- it
	}
	close(jobs)

	var mu sync.Mutex
	var wg sync.WaitGroup
	agora := time.Now()
	chavesVistas := make(map[string]bool)

	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				hist, err := h.b3Client.ObterProventosStatusInvest(item.Ticker, item.IsFII)
				if err != nil || hist == nil {
					continue
				}

				for _, ev := range hist.AssetEarningsModels {
					dataCom, okCom := parseDataFlexivel(ev.Ed)
					dataPag, okPag := parseDataFlexivel(ev.Pd)

					isFuturo := false
					status := "AGUARDANDO_PAGAMENTO"

					if okPag && dataPag.After(agora.AddDate(0, 0, -3)) {
						isFuturo = true
						if okCom && dataCom.After(agora) {
							status = "DATA_COM_IMINENTE"
						}
					} else if !okPag && okCom && dataCom.After(agora.AddDate(0, 0, -30)) {
						// Provento já deliberado com data com recente ou futura, mas data de pagamento a definir
						isFuturo = true
						status = "A_DEFINIR"
						if dataCom.After(agora) {
							status = "DATA_COM_IMINENTE"
						}
					}

					if isFuturo {
						dataPagDisplay := strings.TrimSpace(ev.Pd)
						if dataPagDisplay == "" || dataPagDisplay == "-" {
							dataPagDisplay = "A definir"
						}

						chaveUnica := fmt.Sprintf("%s|%s|%s|%s", item.Ticker, ev.Et, ev.Ed, dataPagDisplay)

						mu.Lock()
						if !chavesVistas[chaveUnica] {
							chavesVistas[chaveUnica] = true
							totalPrevisto := item.Quantidade * ev.V
							radarFuturo = append(radarFuturo, b3import.ProventoFuturo{
								Ticker:          item.Ticker,
								Tipo:            ev.Et,
								DataCom:         ev.Ed,
								DataPagamento:   dataPagDisplay,
								ValorUnitario:   ev.V,
								QuantidadeCotas: item.Quantidade,
								ValorPrevisto:   totalPrevisto,
								Status:          status,
							})
						}
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()

	// 3. Ordenação cronológica intuitiva:
	// Proventos com data de pagamento mais próxima aparecem no topo
	// Proventos "A definir" ficam ordenados por data-com
	sort.Slice(radarFuturo, func(i, j int) bool {
		t1, ok1 := parseDataFlexivel(radarFuturo[i].DataPagamento)
		t2, ok2 := parseDataFlexivel(radarFuturo[j].DataPagamento)
		if ok1 && ok2 {
			return t1.Before(t2)
		}
		if ok1 && !ok2 {
			return true
		}
		if !ok1 && ok2 {
			return false
		}
		// Ambos "A definir": ordena pela data-com
		c1, okC1 := parseDataFlexivel(radarFuturo[i].DataCom)
		c2, okC2 := parseDataFlexivel(radarFuturo[j].DataCom)
		if okC1 && okC2 {
			return c1.Before(c2)
		}
		return radarFuturo[i].Ticker < radarFuturo[j].Ticker
	})

	responderJSON(w, http.StatusOK, map[string]interface{}{
		"total_previsto": len(radarFuturo),
		"proventos":      radarFuturo,
	})
}
