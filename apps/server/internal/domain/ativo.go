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
	StatusComprarMais StatusRecomendacao = "COMPRAR_MAIS" // ✅ Margem favorável
	StatusManter      StatusRecomendacao = "MANTER"       // ⚠️ Neutro ou aprovado em apenas 1 método
	StatusAlerta      StatusRecomendacao = "ALERTA"       // 🚨 Acima do teto / Caro
)

// Ativo representa a entidade central da nossa carteira de investimentos
type Ativo struct {
	Ticker          string             `json:"ticker"`
	Classe          ClasseAtivo        `json:"classe"`
	PrecoAtual      float64            `json:"preco_atual"`
	Dividendos12M   float64            `json:"dividendos_12m"`
	LPA             float64            `json:"lpa"`
	VPA             float64            `json:"vpa"`
	PrecoTetoBazin  float64            `json:"preco_teto_bazin"`
	ValorGraham     float64            `json:"valor_graham"`
	MargemSeguranca float64            `json:"margem_seguranca"`
	Status          StatusRecomendacao `json:"status"`
}

// AvaliarSemaforo define a decisão de alocação com base nos motores analíticos
func (a *Ativo) AvaliarSemaforo() {
	if a.Classe == domainClasseFII() {
		// Para FIIs, a decisão é regida pelo fluxo de proventos de Bazin
		if a.PrecoAtual <= a.PrecoTetoBazin {
			a.Status = StatusComprarMais
		} else {
			a.Status = StatusAlerta
		}
		return
	}

	// Para Ações, combinamos Bazin + Graham
	bazinAprovado := a.PrecoAtual <= a.PrecoTetoBazin
	grahamAprovado := a.ValorGraham > 0 && a.PrecoAtual <= a.ValorGraham

	if bazinAprovado && grahamAprovado {
		a.Status = StatusComprarMais
	} else if bazinAprovado || grahamAprovado {
		a.Status = StatusManter
	} else {
		a.Status = StatusAlerta
	}
}

// Helper para comparação limpa
func domainClasseFII() ClasseAtivo {
	return ClasseFII
}
