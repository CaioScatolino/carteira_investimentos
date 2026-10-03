package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/storage"
)

// Handler gerencia as requisições HTTP da plataforma
type Handler struct {
	store storage.SnapshotStore
}

// NovoHandler cria uma instância do Handler com as dependências injetadas
func NovoHandler(store storage.SnapshotStore) *Handler {
	return &Handler{
		store: store,
	}
}

// responderJSON é um helper auxiliar para padronizar respostas com Content-Type application/json
func responderJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// HealthCheck responde o status operacional da API
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	totalAtivos := len(h.store.ListarTodos())

	resposta := map[string]any{
		"status":       "ok",
		"timestamp":    time.Now().Format(time.RFC3339),
		"total_ativos": totalAtivos,
	}

	responderJSON(w, http.StatusOK, resposta)
}

// RespostaRankings define o formato do payload entregue ao front-end
type RespostaRankings struct {
	Total int             `json:"total"`
	Acoes []*domain.Ativo `json:"acoes"`
	FIIs  []*domain.Ativo `json:"fiis"`
	ETFs  []*domain.Ativo `json:"etfs"`
}

// ObterRankings entrega os 3 rankings completos separados por classe e ordenados por Score
func (h *Handler) ObterRankings(w http.ResponseWriter, r *http.Request) {
	todos := h.store.ListarTodos()

	acoes := make([]*domain.Ativo, 0)
	fiis := make([]*domain.Ativo, 0)
	etfs := make([]*domain.Ativo, 0)
	for _, a := range todos {
		switch a.Classe {
		case domain.ClasseFII:
			fiis = append(fiis, a)
		case domain.ClasseETF:
			etfs = append(etfs, a)
		default:
			acoes = append(acoes, a)
		}
	}

	// Função local para ordenar por Score decrescente (desempate por Volume)
	ordenarPorScore := func(slice []*domain.Ativo) {
		sort.Slice(slice, func(i, j int) bool {
			if slice[i].Score == slice[j].Score {
				return slice[i].VolumeTotal > slice[j].VolumeTotal
			}
			return slice[i].Score > slice[j].Score
		})
	}

	ordenarPorScore(acoes)
	ordenarPorScore(fiis)
	ordenarPorScore(etfs)

	payload := RespostaRankings{
		Total: len(todos),
		Acoes: acoes,
		FIIs:  fiis,
		ETFs:  etfs,
	}

	responderJSON(w, http.StatusOK, payload)
}

// ObterAtivoPorTicker busca os dados fundamentalistas e pareceres de um ativo específico
func (h *Handler) ObterAtivoPorTicker(w http.ResponseWriter, r *http.Request) {
	// Extrai a variável {ticker} da URL (recurso nativo do Go 1.22+)
	ticker := strings.ToUpper(strings.TrimSpace(r.PathValue("ticker")))
	if ticker == "" {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "Ticker não informado"})
		return
	}

	ativo, err := h.store.Obter(ticker)
	if err != nil || ativo == nil {
		responderJSON(w, http.StatusNotFound, map[string]string{
			"erro": fmt.Sprintf("Ativo '%s' não encontrado no cache de mercado", ticker),
		})
		return
	}

	responderJSON(w, http.StatusOK, ativo)
}
