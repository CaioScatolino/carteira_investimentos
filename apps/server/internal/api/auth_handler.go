package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"carteira_investimentos/server/internal/auth"
)

type contextoChave string

const ChaveUsuarioContexto contextoChave = "usuario_autenticado"

// RegistrarUsuario recebe as credenciais, valida a senha forte e cadastra
func (h *Handler) RegistrarUsuario(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de autenticação indisponível (banco de dados offline)",
		})
		return
	}

	var req auth.RequisicaoRegistro
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "formato de corpo JSON inválido",
		})
		return
	}

	resposta, err := h.authService.Registrar(r.Context(), &req)
	if err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusCreated, resposta)
}

// LoginUsuario autentica o usuário e emite o token JWT
func (h *Handler) LoginUsuario(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{
			"erro": "serviço de autenticação indisponível (banco de dados offline)",
		})
		return
	}

	var req auth.RequisicaoLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{
			"erro": "formato de corpo JSON inválido",
		})
		return
	}

	resposta, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{
			"erro": err.Error(),
		})
		return
	}

	responderJSON(w, http.StatusOK, resposta)
}

// ObterPerfilLogado retorna as informações do usuário autenticado no token
func (h *Handler) ObterPerfilLogado(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(ChaveUsuarioContexto).(*auth.ClaimsJWT)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{
			"erro": "sessão não autenticada",
		})
		return
	}

	responderJSON(w, http.StatusOK, map[string]any{
		"id":    claims.UsuarioID,
		"uuid":  claims.UsuarioUUID,
		"email": claims.Email,
		"role":  claims.Role,
	})
}

// ExigirAutenticacao é o Middleware que valida o Bearer JWT e injeta a identidade no contexto
func (h *Handler) ExigirAutenticacao(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.tokenManager == nil {
			responderJSON(w, http.StatusServiceUnavailable, map[string]string{
				"erro": "gerenciador de segurança não inicializado",
			})
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			responderJSON(w, http.StatusUnauthorized, map[string]string{
				"erro": "cabeçalho Authorization ausente. Forneça: Bearer <token>",
			})
			return
		}

		partes := strings.SplitN(authHeader, " ", 2)
		if len(partes) != 2 || strings.ToLower(partes[0]) != "bearer" {
			responderJSON(w, http.StatusUnauthorized, map[string]string{
				"erro": "formato de token inválido. Formato esperado: Bearer <token>",
			})
			return
		}

		tokenStr := strings.TrimSpace(partes[1])
		claims, err := h.tokenManager.ValidarToken(tokenStr)
		if err != nil {
			responderJSON(w, http.StatusUnauthorized, map[string]string{
				"erro": "token inválido ou expirado",
			})
			return
		}

		// Injeta os dados do usuário autenticado dentro do contexto da requisição
		ctx := context.WithValue(r.Context(), ChaveUsuarioContexto, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
