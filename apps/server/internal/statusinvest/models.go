package statusinvest

// QueryResponse é o envelope padrão retornado pelos endpoints de busca avançada do StatusInvest
type QueryResponse[T any] struct {
	List         []T  `json:"list"`
	TotalResults int  `json:"totalResults"`
	HasForecast  bool `json:"hasForecast"`
}

// AcaoStatusInvest mapeia todas as 36 colunas de fundamentos e valuation de ações
type AcaoStatusInvest struct {
	CompanyID int    `json:"companyid"`
	CompanyName string `json:"companyname"`
	Ticker      string `json:"ticker"`
	Price       float64 `json:"price"`

	// Valuation e Múltiplos
	DY          float64 `json:"dy"`
	PL          float64 `json:"p_l"`
	PVP         float64 `json:"p_vp"`
	PEbit       float64 `json:"p_ebit"`
	PEGRatio    float64 `json:"peg_ratio"`
	EVEbit      float64 `json:"ev_ebit"`
	LPA         float64 `json:"lpa"`
	VPA         float64 `json:"vpa"`
	PSR         float64 `json:"p_sr"`
	PAtivo      float64 `json:"p_ativo"`
	PCapitalGiro float64 `json:"p_capitalgiro"`
	PAtivoCirculante float64 `json:"p_ativocirculante"`

	// Rentabilidade e Retorno
	ROE        float64 `json:"roe"`
	ROIC       float64 `json:"roic"`
	ROA        float64 `json:"roa"`
	GiroAtivos float64 `json:"giroativos"`

	// Eficiência e Margens
	MargemBruta   float64 `json:"margembruta"`
	MargemEbit    float64 `json:"margemebit"`
	MargemLiquida float64 `json:"margemliquida"`

	// Endividamento e Estrutura de Capital
	DividaLiquidaPatrimonioLiquido float64 `json:"dividaliquidapatrimonioliquido"`
	DividaLiquidaEbit              float64 `json:"dividaliquidaebit"`
	PLAtivo                        float64 `json:"pl_ativo"`
	PassivoAtivo                   float64 `json:"passivo_ativo"`
	LiquidezCorrente               float64 `json:"liquidezcorrente"`

	// Crescimento e Mercado
	LucrosCAGR5          float64 `json:"lucros_cagr5"`
	ReceitasCAGR5        float64 `json:"receitas_cagr5"`
	LiquidezMediaDiaria  float64 `json:"liquidezmediadiaria"`
	ValorMercado         float64 `json:"valormercado"`

	// Setores e Classificação B3
	SectorID      int    `json:"sectorid"`
	SectorName    string `json:"sectorname"`
	SubSectorID   int    `json:"subsectorid"`
	SubSectorName string `json:"subsectorname"`
	SegmentID     int    `json:"segmentid"`
	SegmentName   string `json:"segmentname"`
}

// FIIStatusInvest mapeia todas as 22 colunas de fundos imobiliários do StatusInvest
type FIIStatusInvest struct {
	CompanyID   int    `json:"companyid"`
	CompanyName string `json:"companyname"`
	Ticker      string `json:"ticker"`
	Price       float64 `json:"price"`

	// Múltiplos e Proventos
	DY                   float64 `json:"dy"`
	PVP                  float64 `json:"p_vp"`
	ValorPatrimonialCota float64 `json:"valorpatrimonialcota"`
	Patrimonio           float64 `json:"patrimonio"`
	LastDividend         float64 `json:"lastdividend"`

	// Operacional e Risco
	Gestao          int     `json:"gestao"`
	GestaoF         string  `json:"gestao_f"`
	Segment         string  `json:"segment"`
	SegmentID       int     `json:"segmentid"`
	SectorID        int     `json:"sectorid"`
	SectorName      string  `json:"sectorname"`
	SubSectorID     int     `json:"subsectorid"`
	SubSectorName   string  `json:"subsectorname"`
	PercentualCaixa float64 `json:"percentualcaixa"`
	DividendCAGR    float64 `json:"dividend_cagr"`
	CotaCAGR        float64 `json:"cota_cagr"`
	NumeroCotistas  float64 `json:"numerocotistas"`
	NumeroCotas     float64 `json:"numerocotas"`

	// Liquidez de Negociação
	LiquidezMediaDiaria float64 `json:"liquidezmediadiaria"`
}
