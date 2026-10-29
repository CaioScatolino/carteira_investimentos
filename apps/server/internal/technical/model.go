package technical

// Candle representa uma vela financeira com Open, High, Low, Close, Volume
type Candle struct {
	Time   int64   `json:"time"` // UNIX timestamp em segundos
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// IndicatorPoint representa um ponto em uma série temporal de indicador (MMA, RSI)
type IndicatorPoint struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

// RespostaGrafico consolida dados para o frontend Lightweight Charts
type RespostaGrafico struct {
	Ticker          string           `json:"ticker"`
	Nome            string           `json:"nome"`
	Intervalo       string           `json:"intervalo"`
	Periodo         string           `json:"periodo"`
	PrecoAtual           float64          `json:"preco_atual"`
	PrecoTetoConsolidado float64          `json:"preco_teto_consolidado"`
	ModelosTeto          []string         `json:"modelos_teto,omitempty"`
	PrecoTetoBazin       float64          `json:"preco_teto_bazin"`
	ValorGraham          float64          `json:"valor_graham"`
	MargemSeguranca      float64          `json:"margem_seguranca"`
	SinalTecnico         string           `json:"sinal_tecnico"`
	DescricaoSinal       string           `json:"descricao_sinal"`
	Candles              []Candle         `json:"candles"`
	MMA8                 []IndicatorPoint `json:"mma8"`
	MMA21                []IndicatorPoint `json:"mma21"`
	MMA20                []IndicatorPoint `json:"mma20,omitempty"`
	MMA52                []IndicatorPoint `json:"mma52,omitempty"`
	MMA200               []IndicatorPoint `json:"mma200"`
	BollingerSuperior    []IndicatorPoint `json:"bollinger_superior,omitempty"`
	BollingerInferior    []IndicatorPoint `json:"bollinger_inferior,omitempty"`
	RSI14                []IndicatorPoint `json:"rsi14"`
}
