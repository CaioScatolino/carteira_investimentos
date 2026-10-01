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

// Ativo representa a entidade central da nossa carteira de investimentos
type Ativo struct {
	Ticker             string             `json:"ticker"`
	Classe             ClasseAtivo        `json:"classe"`
	PrecoAtual         float64            `json:"preco_atual"`
	Dividendos12M      float64            `json:"dividendos_12m"`
	LPA                float64            `json:"lpa"`
	VPA                float64            `json:"vpa"`
	CrescimentoLucro5A float64            `json:"crescimento_lucro_5a"` // % anual (para Peter Lynch)
	VPCota             float64            `json:"vp_cota"`              // Valor Patrimonial p/ Cota (para FIIs)
	PrecoTetoBazin     float64            `json:"preco_teto_bazin"`
	ValorGraham        float64            `json:"valor_graham"`
	PrecoJustoLynch    float64            `json:"preco_justo_lynch"`
	PEGRatio           float64            `json:"peg_ratio"`
	PVP                float64            `json:"pvp"`
	SpreadNTNB         float64            `json:"spread_ntnb"`
	Status             StatusRecomendacao `json:"status"`
}

// AvaliarSemaforo define a recomendação consolidada
func (a *Ativo) AvaliarSemaforo() {
	if a.Classe == ClasseFII {
		// Para FIIs: abaixo do teto Bazin e negociando abaixo ou próximo do VP (P/VP <= 1.02)
		if a.PrecoAtual <= a.PrecoTetoBazin && (a.PVP > 0 && a.PVP <= 1.02) {
			a.Status = StatusComprarMais
		} else if a.PrecoAtual <= a.PrecoTetoBazin {
			a.Status = StatusManter
		} else {
			a.Status = StatusAlerta
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
