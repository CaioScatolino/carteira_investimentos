package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"carteira_investimentos/server/internal/advisor"
	"carteira_investimentos/server/internal/auth"
)

// AnalisarAporteAdvisor processa a requisição síncrona do consultor de aportes
func (h *Handler) AnalisarAporteAdvisor(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok || claims == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "sessão de usuário inválida"})
		return
	}

	if h.advisorService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de consultoria inteligente não inicializado",
		})
		return
	}

	var req advisor.RequisicaoAporte
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "corpo JSON inválido"})
		return
	}

	resp, err := h.advisorService.AnalisarAporte(r.Context(), claims.UsuarioID, req)
	if err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": err.Error()})
		return
	}

	responderJSON(w, http.StatusOK, resp)
}

// StreamAdvisor transmite a recomendação via Server-Sent Events (SSE) com digitação em tempo real
func (h *Handler) StreamAdvisor(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok || claims == nil {
		http.Error(w, "sessão de usuário inválida", http.StatusUnauthorized)
		return
	}

	if h.advisorService == nil {
		http.Error(w, "serviço de consultoria inteligente não inicializado", http.StatusServiceUnavailable)
		return
	}

	q := r.URL.Query()
	valorStr := q.Get("valor")
	valorAporte, _ := strconv.ParseFloat(valorStr, 64)
	if valorAporte <= 0 {
		valorAporte = 1000.0
	}

	perfilStr := q.Get("perfil")
	if perfilStr == "" {
		perfilStr = "EQUILIBRADO"
	}

	permitirNovos := q.Get("permitir_novos") == "true"

	req := advisor.RequisicaoAporte{
		ValorAporte:         valorAporte,
		Perfil:              advisor.PerfilInvestidor(perfilStr),
		PermitirNovosAtivos: permitirNovos,
	}

	h.advisorService.StreamAdvisorSSE(w, r, claims.UsuarioID, req)
}

