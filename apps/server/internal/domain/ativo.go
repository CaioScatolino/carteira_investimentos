package domain

// ClasseAtivo representa o tipo de ativo negociado na B3
type ClasseAtivo string

const (
	ClasseAcao ClasseAtivo = "ACAO"
	ClasseFII  ClasseAtivo = "FII"
	ClasseETF  ClasseAtivo = "ETF"
)

// StatusRecomendacao define os estados possíveis do semáforo da carteira
type StatusRecomendacao string

const (
	StatusComprarMais StatusRecomendacao = "COMPRAR_MAIS"
	StatusManter      StatusRecomendacao = "MANTER"
	StatusAlerta      StatusRecomendacao = "ALERTA"
)

// StatusParecer define o julgamento individual de cada escola de valuation
type StatusParecer string

const (
	ParecerAprovado  StatusParecer = "APROVADO"
	ParecerAtencao   StatusParecer = "ATENCAO"
	ParecerReprovado StatusParecer = "REPROVADO"
	ParecerNaoAplica StatusParecer = "NAO_APLICA"
)

// ParecerItem sintetiza o parecer de um motor específico para o investidor
type ParecerItem struct {
	Status  StatusParecer `json:"status"`  // APROVADO, ATENCAO, REPROVADO
	Metrica string        `json:"metrica"` // Ex: "Teto: R$ 53.80 | Yield: 6.49%"
	Detalhe string        `json:"detalhe"` // Ex: "Margem: +8.1%"
}

// Ativo representa a entidade central da nossa carteira de investimentos
type Ativo struct {
	Ticker             string                 `json:"ticker"`
	Nome               string                 `json:"nome,omitempty"` // Nome resumido da B3 (NOMRES)
	Classe             ClasseAtivo            `json:"classe"`
	CNPJ               string                 `json:"cnpj,omitempty"`         // CNPJ oficial cadastrado no MySQL/CVM
	VolumeTotal        float64                `json:"volume_total,omitempty"` // Volume financeiro real negociado no pregão da B3
	PrecoAtual         float64                `json:"preco_atual"`
	Dividendos12M      float64                `json:"dividendos_12m"`
	LPA                float64                `json:"lpa"`
	VPA                float64                `json:"vpa"`
	CrescimentoLucro5A float64                `json:"crescimento_lucro_5a"` // % anual (para Peter Lynch)
	VPCota             float64                `json:"vp_cota"`              // Valor Patrimonial p/ Cota (para FIIs)
	PrecoTetoBazin     float64                `json:"preco_teto_bazin"`
	ValorGraham        float64                `json:"valor_graham"`
	PrecoJustoLynch    float64                `json:"preco_justo_lynch"`
	PEGRatio           float64                `json:"peg_ratio"`
	PVP                float64                `json:"pvp"`
	SpreadNTNB         float64                `json:"spread_ntnb"`          // Spread de retorno sobre o Tesouro IPCA+ (% a.a.)
	YieldExigido       float64                `json:"yield_exigido"`        // Yield mínimo exigido para compensar a taxa de juros (% a.a.)
	VereditoRisco      string                 `json:"veredito_risco"`       // "COMPENSA_RISCO", "NEUTRO", "RISCO_DESCOMPENSADO"
	JustificativaRisco string                 `json:"justificativa_risco"`  // Explicação detalhada da relação risco vs retorno
	TIRProjetada       float64                `json:"tir_projetada"`        // Taxa interna de retorno estimada (% a.a.)
	Score              float64                `json:"score"`                // Score Fundamentalista Geral (0 a 100)
	DY                 float64                `json:"dy"`                   // % Dividend Yield nos últimos 12M
	PL                 float64                `json:"pl"`                   // P/L (Preço sobre Lucro)
	PVPReal            float64                `json:"pvp_real"`             // P/VP para ações (Preco / VPA)
	ROE                float64                `json:"roe"`                  // % ROE (LPA / VPA)
	EarningsYield      float64                `json:"earnings_yield"`       // % LPA / PrecoAtual (Inverso do P/L)
	PrecoTetoGordon    float64                `json:"preco_teto_gordon"`    // Teto Gordon com crescimento
	MargemBazin        float64                `json:"margem_bazin"`         // % Margem de Segurança Bazin
	MargemGraham       float64                `json:"margem_graham"`        // % Margem de Segurança Graham

	// Métricas Avançadas StatusInvest
	ROIC                float64 `json:"roic,omitempty"`
	ROA                 float64 `json:"roa,omitempty"`
	GiroAtivos          float64 `json:"giro_ativos,omitempty"`
	MargemBruta         float64 `json:"margem_bruta,omitempty"`
	MargemEbit          float64 `json:"margem_ebit,omitempty"`
	MargemLiquida       float64 `json:"margem_liquida,omitempty"`
	DividaLiquidaPL     float64 `json:"divida_liquida_pl,omitempty"`
	DividaLiquidaEbit   float64 `json:"divida_liquida_ebit,omitempty"`
	LiquidezCorrente    float64 `json:"liquidez_corrente,omitempty"`
	LiquidezMediaDiaria float64 `json:"liquidez_media_diaria,omitempty"`
	ValorMercado        float64 `json:"valor_mercado,omitempty"`
	Setor               string  `json:"setor,omitempty"`
	Subsetor            string  `json:"subsetor,omitempty"`
	Segmento            string  `json:"segmento,omitempty"`
	Gestao              string  `json:"gestao,omitempty"`
	PercentualCaixa     float64 `json:"percentual_caixa,omitempty"`
	NumeroCotistas      float64 `json:"numero_cotistas,omitempty"`
	LastDividend        float64 `json:"last_dividend,omitempty"`

	// Qualidade e Sustentabilidade de Proventos (Blindagem Anti-Yield Trap)
	Payout                    float64 `json:"payout,omitempty"`                      // % do LPA pago em proventos (Dividendos12M / LPA * 100)
	IsProventoAtipico         bool    `json:"is_provento_atipico,omitempty"`          // Flag para proventos não recorrentes, extraordinários ou amortizações
	AlertaRisco               string  `json:"alerta_risco,omitempty"`                // Alerta pedagógico (ex: "Yield Atípico (32.9%)", "Amortização de Capital")
	PrecoTetoBazinSustentavel float64 `json:"preco_teto_bazin_sustentavel,omitempty"` // Preço teto ajustado para capacidade de lucro real

	// Décio Bazin Histórico 5 Anos (B3 Oficial vs StatusInvest)
	MediaDividendos5A            float64            `json:"media_dividendos_5a,omitempty"`            // Média anual de proventos líquidos dos últimos 5 anos
	MediaDividendos5ANormalizada float64            `json:"media_dividendos_5a_normalizada,omitempty"` // Média anual saneada com Winsorização anti-distorção
	PrecoTetoBazin5A             float64            `json:"preco_teto_bazin_5a,omitempty"`            // Preço Teto Bazin calculado sobre a média de 5A (MediaDiv5A / 0.06)
	MargemBazin5A                float64            `json:"margem_bazin_5a,omitempty"`                // Margem de segurança sobre o Teto de 5A
	Dividendos12MB3              float64            `json:"dividendos_12m_b3,omitempty"`              // Proventos 12M auditados direto na B3
	HistoricoDividendosAnual     map[string]float64 `json:"historico_dividendos_anual,omitempty"`    // Proventos ano a ano (2021, 2022, 2023, 2024, 2025)
	Diferenca12MB3StatusInvest   float64            `json:"diferenca_12m_b3_statusinvest,omitempty"`  // Diferença monetária B3 vs StatusInvest
	AderenciaStatusInvest        string             `json:"aderencia_statusinvest,omitempty"`         // Auditoria de conferência ("100% Aderente", etc.)
	TeveOutlier5A                bool               `json:"teve_outlier_5a,omitempty"`                // Indica se a série histórica continha um ano de distribuição extraordinária
	ObservacaoOutlier            string             `json:"observacao_outlier,omitempty"`             // Explicação transparente da normalização de proventos

	// Preço Teto Consolidado e Prêmio de Valorização (Consenso Multi-Modelo)
	PrecoTetoConsolidado       float64  `json:"preco_teto_consolidado"`                   // Preço teto consensual dos motores de valuation
	PremioDescontoPercentual   float64  `json:"premio_desconto_percentual"`               // % de Prêmio/Upside: ((PrecoTetoConsolidado - PrecoAtual) / PrecoAtual) * 100
	MargemSegurancaConsolidada float64  `json:"margem_seguranca_consolidada"`             // % Margem de Segurança: ((PrecoTetoConsolidado - PrecoAtual) / PrecoTetoConsolidado) * 100
	ModelosTetoConsolidado     []string `json:"modelos_teto_consolidado,omitempty"`       // Modelos que convergiram (ex: ["Bazin 5A", "Graham", "Gordon"])

	// Porte e Robustez Institucional (Blue Chips vs Small/Micro Caps)
	Porte string `json:"porte,omitempty"` // BLUE_CHIP, MID_CAP, SMALL_CAP, MICRO_CAP, FII_GIGANTE, FII_CONSOLIDADO, FII_MEDIO, FII_CONCENTRADO

	Pareceres          map[string]ParecerItem `json:"pareceres,omitempty"`  // Pareceres individuais por motor
	Status             StatusRecomendacao     `json:"status"`
}

// AvaliarSemaforo define a recomendação consolidada respeitando o Score Composto e blindagens de risco
func (a *Ativo) AvaliarSemaforo() {
	if a.Classe == ClasseETF {
		if a.VolumeTotal >= 5_000_000.0 {
			a.Status = StatusComprarMais // Alta liquidez institucional
		} else if a.VolumeTotal >= 500_000.0 {
			a.Status = StatusManter // Liquidez regular de mercado
		} else {
			a.Status = StatusAlerta // Baixa liquidez
		}
		return
	}

	// BLINDAGEM 1: FIIs em liquidação ou amortização extraordinária (PATL11, BBFI11, etc)
	if a.Classe == ClasseFII && (a.IsProventoAtipico || a.PVP < 0.35 || a.DY > 20.0) {
		a.Status = StatusAlerta
		return
	}

	// BLINDAGEM 2: Ações com prejuízo (LPA <= 0) ou patrimônio líquido negativo (VPA <= 0)
	if a.Classe == ClasseAcao && (a.LPA <= 0 || a.VPA <= 0) {
		a.Status = StatusAlerta
		return
	}

	// Alinhamento Consolidado com o Score Fundamentalista Multi-Fatorial
	if a.Score >= 70.0 {
		a.Status = StatusComprarMais
	} else if a.Score >= 50.0 {
		a.Status = StatusManter
	} else {
		a.Status = StatusAlerta
	}
}
