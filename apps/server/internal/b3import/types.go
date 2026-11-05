package b3import

import (
	"time"

	"carteira_investimentos/server/internal/portfolio"
)

// LinhaMovimentacaoB3 representa uma linha bruta do arquivo XLSX da Área do Investidor B3
type LinhaMovimentacaoB3 struct {
	LinhaOrigem   int       `json:"linha_origem"`
	EntradaSaida  string    `json:"entrada_saida"` // "Credito" ou "Debito"
	Data          time.Time `json:"data"`
	Movimentacao  string    `json:"movimentacao"`
	Produto       string    `json:"produto"`
	Ticker        string    `json:"ticker"`
	Instituicao   string    `json:"instituicao"`
	Quantidade    float64   `json:"quantidade"`
	PrecoUnitario float64   `json:"preco_unitario"`
	ValorOperacao float64   `json:"valor_operacao"`
}

// PosicaoImportada resume a custódia líquida calculada ao final do replay cronológico
type PosicaoImportada struct {
	Ticker         string                `json:"ticker"`
	RazaoSocial    string                `json:"razao_social"`
	Classe         portfolio.ClasseAtivo `json:"classe"`
	Quantidade     float64               `json:"quantidade"`
	PrecoMedio     float64               `json:"preco_medio"`
	ValorAplicado  float64               `json:"valor_aplicado"`
	Corretora      string                `json:"corretora"`
	TotalCompras   int                   `json:"total_compras"`
	TotalVendas    int                   `json:"total_vendas"`
	TotalProventos float64               `json:"total_proventos_recebidos"`
}

// OperacaoImportada representa cada compra ou venda reconstruída para o livro-razão
type OperacaoImportada struct {
	Ticker        string                  `json:"ticker"`
	Classe        portfolio.ClasseAtivo   `json:"classe"`
	Tipo          portfolio.TipoTransacao `json:"tipo"` // "COMPRA", "VENDA"
	Data          time.Time               `json:"data"`
	Quantidade    float64                 `json:"quantidade"`
	PrecoUnitario float64                 `json:"preco_unitario"`
	ValorTotal    float64                 `json:"valor_total"`
	Corretora     string                  `json:"corretora"`
	Origem        string                  `json:"origem"`
	Notas         string                  `json:"notas"`
}

// ProventoImportado representa um provento pago recebido na conta do investidor
type ProventoImportado struct {
	Ticker        string    `json:"ticker"`
	Tipo          string    `json:"tipo"` // "RENDIMENTO", "DIVIDENDO", "JCP", "JUROS", "RESTITUICAO_CAPITAL"
	DataPagamento time.Time `json:"data_pagamento"`
	Quantidade    float64   `json:"quantidade"`
	ValorUnitario float64   `json:"valor_unitario"`
	ValorTotal    float64   `json:"valor_total"`
	Corretora     string    `json:"corretora"`
}

// RelatorioAnaliseB3 é o diagnóstico completo retornado no preview da importação
type RelatorioAnaliseB3 struct {
	TotalMovimentacoes int                 `json:"total_movimentacoes"`
	DataInicio         time.Time           `json:"data_inicio"`
	DataFim            time.Time           `json:"data_fim"`
	TotalAtivosAtivos  int                 `json:"total_ativos_ativos"`
	TotalInvestido     float64             `json:"total_investido_custo"`
	TotalProventos     float64             `json:"total_proventos_historicos"`
	Posicoes           []PosicaoImportada  `json:"posicoes"`
	Operacoes          []OperacaoImportada `json:"operacoes"`
	Proventos          []ProventoImportado `json:"proventos"`
	Avisos             []string            `json:"avisos,omitempty"`
}

// ConfirmarImportacaoRequest define as opções para salvar a importação no banco
type ConfirmarImportacaoRequest struct {
	LimparDadosAnteriores bool                `json:"limpar_dados_anteriores"` // Substitui custódia e transações antigas
	ImportarProventos     bool                `json:"importar_proventos"`      // Grava na tabela proventos_recebidos
	Posicoes              []PosicaoImportada  `json:"posicoes"`
	Operacoes             []OperacaoImportada `json:"operacoes"`
	Proventos             []ProventoImportado `json:"proventos"`
}

