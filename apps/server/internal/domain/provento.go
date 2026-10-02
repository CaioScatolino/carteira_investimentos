package domain

import "time"

// TipoProvento representa as modalidades oficiais de proventos em dinheiro na B3
type TipoProvento string

const (
	ProventoDividendo  TipoProvento = "DIVIDENDO"      // Dividendo de ações (Isento de IR para PF)
	ProventoJCP        TipoProvento = "JCP"            // Juros sobre Capital Próprio (Tributado em 15% na fonte)
	ProventoRendimento TipoProvento = "RENDIMENTO_FII" // Rendimento mensal de Fundo Imobiliário (Isento de IR para PF)
)

// EventoProvento representa um pagamento individual de provento deliberado por uma empresa ou fundo
type EventoProvento struct {
	Identificador string       `json:"identificador"` // Ticker (ex: "PETR4") ou CNPJ oficial
	Tipo          TipoProvento `json:"tipo"`
	DataPagamento time.Time    `json:"data_pagamento"`
	ValorUnitario float64      `json:"valor_unitario"` // Valor bruto pago por ação ou cota
}

// ValorLiquido retorna o valor efetivo recebido pelo investidor pessoa física.
// Regra Décio Bazin: JCP sofre retenção compulsória de 15% de IR na fonte.
// O Preço Teto deve ser calculado exclusivamente sobre dinheiro limpo no bolso!
func (e EventoProvento) ValorLiquido() float64 {
	if e.Tipo == ProventoJCP {
		return e.ValorUnitario * 0.85 // Deduz 15% de imposto retido
	}
	return e.ValorUnitario
}

// CalcularProventos12M aplica a janela deslizante de 12 meses (Trailing 12 Months - TTM).
// Filtra apenas os eventos pagos nos últimos 365 dias em relação à data de referência
// e retorna a somatória líquida por ação/cota.
func CalcularProventos12M(eventos []EventoProvento, dataReferencia time.Time) float64 {
	// Data de corte: exatamente 1 ano (365 dias) antes da referência
	inicioJanela := dataReferencia.AddDate(-1, 0, 0)
	totalLiquido := 0.0

	for _, ev := range eventos {
		// Ignora eventos fora da janela de 12 meses
		if ev.DataPagamento.Before(inicioJanela) {
			continue
		}
		totalLiquido += ev.ValorLiquido()
	}

	return totalLiquido
}

// AgruparProventosPorAtivo organiza uma lista plana de eventos em um mapa indexado por Ticker/CNPJ.
// Essencial para processar arquivos em lote em alta velocidade (O(N)).
func AgruparProventosPorAtivo(eventos []EventoProvento) map[string][]EventoProvento {
	mapa := make(map[string][]EventoProvento)
	for _, ev := range eventos {
		mapa[ev.Identificador] = append(mapa[ev.Identificador], ev)
	}
	return mapa
}
