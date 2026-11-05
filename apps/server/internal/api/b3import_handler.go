package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"carteira_investimentos/server/internal/auth"
	"carteira_investimentos/server/internal/b3import"
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

	radarFuturo := make([]b3import.ProventoFuturo, 0)
	agora := time.Now()

	for _, item := range resumoCarteira.Posicoes {
		if item.Quantidade <= 0 {
			continue
		}

		isFII := item.Classe == "FII"
		// Consulta eventos do ativo via StatusInvest / B3
		hist, err := h.b3Client.ObterProventosStatusInvest(item.Ticker, isFII)
		if err != nil || hist == nil {
			continue
		}

		for _, ev := range hist.AssetEarningsModels {
			dataCom, errCom := time.Parse("02/01/2006", ev.Ed)
			dataPag, errPag := time.Parse("02/01/2006", ev.Pd)
			if errPag != nil {
				continue
			}

			// Se a data de pagamento for futura (ou até 3 dias atrás para visualização imediata)
			if dataPag.After(agora.AddDate(0, 0, -3)) {
				status := "AGUARDANDO_PAGAMENTO"
				if errCom == nil && dataCom.After(agora) {
					status = "DATA_COM_IMINENTE"
				}

				totalPrevisto := item.Quantidade * ev.V
				radarFuturo = append(radarFuturo, b3import.ProventoFuturo{
					Ticker:          item.Ticker,
					Tipo:            ev.Et,
					DataCom:         ev.Ed,
					DataPagamento:   ev.Pd,
					ValorUnitario:   ev.V,
					QuantidadeCotas: item.Quantidade,
					ValorPrevisto:   totalPrevisto,
					Status:          status,
				})
			}
		}
	}

	responderJSON(w, http.StatusOK, map[string]interface{}{
		"total_previsto": len(radarFuturo),
		"proventos":      radarFuturo,
	})
}
