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
	SpreadNTNB         float64                `json:"spread_ntnb"`
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

	Pareceres          map[string]ParecerItem `json:"pareceres,omitempty"`  // Pareceres individuais por motor
	Status             StatusRecomendacao     `json:"status"`
}

// AvaliarSemaforo define a recomendação consolidada
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

	if a.Classe == ClasseFII {
		// Se temos o Preço Teto Bazin calculado:
		if a.PrecoTetoBazin > 0 {
			// Exige P/VP justo E prêmio de risco positivo sobre a NTN-B
			if a.PrecoAtual <= a.PrecoTetoBazin && (a.PVP > 0 && a.PVP <= 1.02) && a.SpreadNTNB >= 0 {
				a.Status = StatusComprarMais
			} else if a.PrecoAtual <= a.PrecoTetoBazin || (a.PVP > 0 && a.PVP <= 1.00) {
				a.Status = StatusManter
			} else {
				a.Status = StatusAlerta
			}
		} else {
			// Se o histórico de proventos ainda não foi carregado, avalia pelo desconto patrimonial oficial (CVM):
			if a.PVP > 0 && a.PVP <= 0.95 {
				a.Status = StatusComprarMais // Mais de 5% de desconto sobre o valor patrimonial
			} else if a.PVP > 0 && a.PVP <= 1.02 {
				a.Status = StatusManter // Negociando no valor justo patrimonial
			} else {
				a.Status = StatusAlerta // Ágio excessivo (P/VP esticado)
			}
		}
		return
	}

	// Para Ações: cruzamento de múltiplos motores
	pontos := 0
	if a.PrecoTetoBazin > 0 && a.PrecoAtual <= a.PrecoTetoBazin {
		pontos++
	}
	if a.ValorGraham > 0 && a.PrecoAtual <= a.ValorGraham {
		pontos++
	}
	if a.PEGRatio > 0 && a.PEGRatio <= 1.0 { // PEG de Lynch subavaliado
		pontos++
	}

	if pontos >= 2 {
		a.Status = StatusComprarMais
	} else if pontos == 1 {
		a.Status = StatusManter
	} else {
		a.Status = StatusAlerta
	}
}
