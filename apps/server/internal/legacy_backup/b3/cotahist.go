package b3

import "time"

// CotacaoDiaria armazena as informações oficiais de negociação de um ativo no pregão da B3
type CotacaoDiaria struct {
	Data            time.Time `json:"data"`
	Ticker          string    `json:"ticker"`
	NomeResumido    string    `json:"nome_resumido"`
	PrecoAbertura   float64   `json:"preco_abertura"`
	PrecoMaximo     float64   `json:"preco_maximo"`
	PrecoMinimo     float64   `json:"preco_minimo"`
	PrecoMedio      float64   `json:"preco_medio"`
	PrecoFechamento float64   `json:"preco_fechamento"` // Campo PREULT do COTAHIST
	VolumeTotal     float64   `json:"volume_total"`     // Campo VOLTOT (liquidez financeira)
	QuantidadeNeg   int       `json:"quantidade_negocios"`
	CodigoBDI       string    `json:"codigo_bdi"` // "02" = Lote Padrão, "12" = FII
}

// ResumoCargaB3 estatísticas consolidadas do pregão processado
type ResumoCargaB3 struct {
	DataPregao      string
	TotalAtivos     int
	VolumeTotal     float64
	TempoExecucaoMs int64
}
