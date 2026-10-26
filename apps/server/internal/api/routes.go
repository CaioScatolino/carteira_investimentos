package api

import (
	"log"
	"net/http"
	"time"
)

// ConfigurarRotas inicializa o roteador ServeMux do Go 1.22+ e aplica os middlewares globais
func (h *Handler) ConfigurarRotas() http.Handler {
	mux := http.NewServeMux()

	// 1. Registro dos Endpoints com Verbos HTTP explícitos
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("GET /api/v1/macro", h.ObterCenarioMacro)
	mux.HandleFunc("GET /api/v1/rankings", h.ObterRankings)
	mux.HandleFunc("GET /api/v1/ativos/{ticker}", h.ObterAtivoPorTicker)
	mux.HandleFunc("GET /api/v1/ativos/{ticker}/dividendos", h.ObterAnaliseDividendos)
	mux.HandleFunc("GET /api/v1/proventos/calendario", h.ObterCalendarioProventos)
	// Endpoints Públicos de Autenticação
	mux.HandleFunc("POST /api/v1/auth/register", h.RegistrarUsuario)
	mux.HandleFunc("POST /api/v1/auth/login", h.LoginUsuario)

	// Endpoint Protegido (Requer JWT válido no cabeçalho)
	mux.Handle("GET /api/v1/auth/me", h.ExigirAutenticacao(http.HandlerFunc(h.ObterPerfilLogado)))

	// Endpoints Protegidos de Gestão de Carteira Multi-Classe (Requer JWT)
	mux.Handle("GET /api/v1/carteira", h.ExigirAutenticacao(http.HandlerFunc(h.ObterCarteira)))
	mux.Handle("POST /api/v1/carteira/posicoes", h.ExigirAutenticacao(http.HandlerFunc(h.SalvarPosicaoCarteira)))
	mux.Handle("DELETE /api/v1/carteira/posicoes/{id}", h.ExigirAutenticacao(http.HandlerFunc(h.ExcluirPosicaoCarteira)))
	mux.Handle("POST /api/v1/carteira/rebalancear", h.ExigirAutenticacao(http.HandlerFunc(h.RebalancearCarteira)))
	mux.Handle("GET /api/v1/carteira/transacoes", h.ExigirAutenticacao(http.HandlerFunc(h.ListarTransacoesCarteira)))

	// 2. Encadeamento de Middlewares: Logging -> CORS -> Mux
	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)

	return handler
}

// corsMiddleware permite que aplicações web (Next.js/React) consumam nossa API sem bloqueios
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Libera acesso para qualquer origem (ou restrinja ao localhost:3000 se preferir)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Trata a requisição de pré-voo (preflight OPTIONS) do navegador
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware registra cada chamada HTTP recebida e o tempo de execução
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("[HTTP] %-6s %-25s | Duração: %v", r.Method, r.URL.Path, time.Since(inicio))
	})
}
