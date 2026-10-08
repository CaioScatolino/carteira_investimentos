package history

import (
	"encoding/json"
	"time"

	"carteira_investimentos/server/internal/domain"
)

// RegistroHistorico representa a linha congelada de um ativo no pregão diário
type RegistroHistorico struct {
	ID                       int64     `json:"id"`
	DataPregao               time.Time `json:"data_pregao"`
	Ticker                   string    `json:"ticker"`
	Classe                   string    `json:"classe"`
	PosicaoRanking           int       `json:"posicao_ranking"`
	Score                    float64   `json:"score"`
	PrecoAtual               float64   `json:"preco_atual"`
	PrecoTetoConsolidado     float64   `json:"preco_teto_consolidado"`
	PremioDescontoPercentual float64   `json:"premio_desconto_percentual"`
	DY                       float64   `json:"dy"`
	PL                       float64   `json:"pl"`
	PVP                      float64   `json:"pvp"`
	Status                   string    `json:"status"`
	IsProventoAtipico        bool      `json:"is_provento_atipico"`
	ModelosJSON              string    `json:"modelos_json"`
	CreatedAt                time.Time `json:"created_at"`
}

// ConverterAtivoParaRegistro transforma o ativo auditado em memória no registro para persistência
func ConverterAtivoParaRegistro(dataPregao time.Time, posicao int, a *domain.Ativo) *RegistroHistorico {
	// P/VP específico dependendo se é Ação ou FII
	pvp := a.PVPReal
	if a.Classe == domain.ClasseFII {
		pvp = a.PVP
	}

	// Serializa o snapshot dos motores de valuation em formato JSON
	modelos := map[string]interface{}{
		"preco_teto_bazin":    a.PrecoTetoBazin,
		"preco_teto_bazin_5a": a.PrecoTetoBazin5A,
		"valor_graham":        a.ValorGraham,
		"preco_justo_lynch":   a.PrecoJustoLynch,
		"preco_teto_gordon":   a.PrecoTetoGordon,
		"preco_teto_dcf":      a.PrecoTetoDCF,
		"modelos_convergiram": a.ModelosTetoConsolidado,
		"piotroski_score":     a.PiotroskiScore,
	}
	modelosBytes, _ := json.Marshal(modelos)

	return &RegistroHistorico{
		DataPregao:               dataPregao,
		Ticker:                   a.Ticker,
		Classe:                   string(a.Classe),
		PosicaoRanking:           posicao,
		Score:                    a.Score,
		PrecoAtual:               a.PrecoAtual,
		PrecoTetoConsolidado:     a.PrecoTetoConsolidado,
		PremioDescontoPercentual: a.PremioDescontoPercentual,
		DY:                       a.DY,
		PL:                       a.PL,
		PVP:                      pvp,
		Status:                   string(a.Status),
		IsProventoAtipico:        a.IsProventoAtipico,
		ModelosJSON:              string(modelosBytes),
	}
}
