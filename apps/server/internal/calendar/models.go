package calendar

const (
	TipoDataCom       = "DATA_COM"
	TipoDataPagamento = "DATA_PAGAMENTO"
)

// MesProvento detalha a frequência e probabilidade de pagamento em um mês específico (1 a 12)
type MesProvento struct {
	Mes             int      `json:"mes"`               // 1 a 12
	NomeMes         string   `json:"nome_mes"`          // "Janeiro", "Fevereiro", ...
	Probabilidade   float64  `json:"probabilidade"`     // 0.0 a 100.0%
	Ocorrencias     int      `json:"ocorrencias"`       // Anos em que houve provento no mês
	TotalAnos       int      `json:"total_anos"`        // Total de anos considerados (ex: 5)
	ValorMedio      float64  `json:"valor_medio"`       // Valor médio pago por ação no mês
	AnosComProvento []int    `json:"anos_com_provento"` // Lista de anos [2021, 2022, 2024]
	Tipos           []string `json:"tipos"`             // ["DIVIDENDO", "JRS CAP PROPRIO"]
}

// AtivoCalendario consolida os 12 meses de proventos previstos para uma ação ou FII
type AtivoCalendario struct {
	Ticker                string        `json:"ticker"`
	Nome                  string        `json:"nome"`
	Classe                string        `json:"classe"`                 // "ACAO", "FII"
	Setor                 string        `json:"setor"`
	Segmento              string        `json:"segmento"`
	PrecoAtual            float64       `json:"preco_atual"`
	DY                    float64       `json:"dy"`
	Score                 float64       `json:"score"`
	Porte                 string        `json:"porte"`
	Status                string        `json:"status"`                 // "COMPRAR_MAIS", "MANTER", "ALERTA"
	Meses                 []MesProvento `json:"meses"`                  // Sempre 12 posições (Jan a Dez)
	TotalMesesPagadores   int           `json:"total_meses_pagadores"`   // Meses com prob > 0
	MesMaisProvavel       string        `json:"mes_mais_provavel"`       // Nome do mês com maior prob
	MaiorProbabilidade    float64       `json:"maior_probabilidade"`     // Maior prob encontrada
	TotalEventosHistorico int           `json:"total_eventos_historico"` // Total de proventos na base
	PossuiNaCarteira      bool          `json:"possui_na_carteira"`      // Indicador se está na custódia
}

// ResumoCalendario encapsula a resposta consolidada do endpoint de calendário
type ResumoCalendario struct {
	TotalAtivos         int               `json:"total_ativos"`
	TipoData            string            `json:"tipo_data"`            // "DATA_COM" ou "DATA_PAGAMENTO"
	PeriodoAnos         int               `json:"periodo_anos"`         // Ex: 5
	ProbabilidadeMinima float64           `json:"probabilidade_minima"` // Filtro aplicado
	MesesMaisFrequentes map[string]int    `json:"meses_mais_frequentes"` // Mapeamento mês -> total ativos pagadores
	Ativos              []AtivoCalendario `json:"ativos"`
}

// FiltroCalendario define os critérios de busca para a API
type FiltroCalendario struct {
	TipoData            string   // "DATA_COM" ou "DATA_PAGAMENTO"
	ProbabilidadeMinima float64  // 0 a 100
	ApenasCarteira      bool     // Filtrar apenas posições do usuário
	TickersCarteira     []string // Tickers presentes na carteira do usuário
	Classe              string   // "ACAO", "FII" ou ""
	Setor               string   // Filtrar por setor específico
	Busca               string   // Termo de busca por ticker ou nome
}
