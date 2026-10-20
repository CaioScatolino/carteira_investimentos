package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"carteira_investimentos/server/internal/auth"
	"carteira_investimentos/server/internal/portfolio"
)

// ObterCarteira retorna o patrimônio consolidado do usuário logado
func (h *Handler) ObterCarteira(w http.ResponseWriter, r *http.Request) {
	if h.portfolioService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de carteira indisponível",
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

	resumo, err := h.portfolioService.ObterResumoCarteira(r.Context(), claims.UsuarioID)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{
			"erro": "falha ao consolidar carteira: " + err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, resumo)
}

// SalvarPosicaoCarteira cadastra ou atualiza um ativo na carteira do usuário
func (h *Handler) SalvarPosicaoCarteira(w http.ResponseWriter, r *http.Request) {
	if h.portfolioService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de carteira indisponível",
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

	var req portfolio.CriarPosicaoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "formato de corpo JSON inválido",
		})
		return
	}

	pos, err := h.portfolioService.AdicionarOuAtualizarPosicao(r.Context(), claims.UsuarioID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, portfolio.ErrDadosInvalidos) {
			status = http.StatusBadRequest
		}
		responderJSON(w, status, map[string]string{
			"erro": err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusCreated, pos)
}

// ExcluirPosicaoCarteira remove um ativo com proteção estrita Anti-IDOR
func (h *Handler) ExcluirPosicaoCarteira(w http.ResponseWriter, r *http.Request) {
	if h.portfolioService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de carteira indisponível",
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

	idStr := r.PathValue("id")
	posicaoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || posicaoID <= 0 {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "identificador de posição inválido",
		})
		return
	}

	err = h.portfolioService.ExcluirPosicao(r.Context(), claims.UsuarioID, posicaoID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, portfolio.ErrPosicaoNaoEncontrada) {
			status = http.StatusNotFound
		}
		responderJSON(w, status, map[string]string{
			"erro": err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{
		"mensagem": "posição excluída com sucesso",
	})
}

// ListarTransacoesCarteira retorna o extrato de movimentações contábeis
func (h *Handler) ListarTransacoesCarteira(w http.ResponseWriter, r *http.Request) {
	if h.portfolioService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de carteira indisponível",
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

	transacoes, err := h.portfolioService.ListarTransacoes(r.Context(), claims.UsuarioID)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{
			"erro": "falha ao consultar transações: " + err.Error(),
		})
		return
	}

	if transacoes == nil {
		transacoes = []portfolio.Transacao{}
	}

	responderJSON(w, http.StatusOK, transacoes)
}

type RequisicaoRebalancear struct {
	ValorAporte float64 `json:"valor_aporte"`
}

// RebalancearCarteira simula o rebalanceamento inteligente sem vendas
func (h *Handler) RebalancearCarteira(w http.ResponseWriter, r *http.Request) {
	if h.portfolioService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de carteira indisponível",
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

	var req RequisicaoRebalancear
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "formato de corpo JSON inválido",
		})
		return
	}

	resultado, err := h.portfolioService.RebalancearCarteira(r.Context(), claims.UsuarioID, req.ValorAporte)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{
			"erro": "falha ao calcular rebalanceamento: " + err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, resultado)
}
