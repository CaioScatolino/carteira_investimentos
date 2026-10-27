package portfolio

import "time"

// ClasseAtivo define as 5 macro-classes de ativos suportadas na carteira.
type ClasseAtivo string

const (
	ClasseRendaFixa     ClasseAtivo = "RENDA_FIXA"
	ClasseFII           ClasseAtivo = "FII"
	ClasseAcaoBR        ClasseAtivo = "ACAO_BR"
	ClasseInternacional ClasseAtivo = "INTERNACIONAL"
	ClasseCaixa         ClasseAtivo = "CAIXA"
)

// TipoTransacao define as operações contábeis aceitas.
type TipoTransacao string

const (
	TransacaoCompra  TipoTransacao = "COMPRA"
	TransacaoVenda   TipoTransacao = "VENDA"
	TransacaoAporte  TipoTransacao = "APORTE"
	TransacaoResgate TipoTransacao = "RESGATE"
)

// Posicao representa o registro consolidado na tabela `carteira_posicoes`.
type Posicao struct {
	ID             int64       `json:"id"`
	UsuarioID      int64       `json:"usuario_id"`
	Ticker         string      `json:"ticker"`
	Classe         ClasseAtivo `json:"classe"`
	CategoriaMacro string      `json:"categoria_macro"`
	Quantidade     float64     `json:"quantidade"`
	PrecoMedio     float64     `json:"preco_medio"`
	ValorAplicado  float64     `json:"valor_aplicado"`
	TaxaContratada *string     `json:"taxa_contratada,omitempty"`
	Indexador      *string     `json:"indexador,omitempty"`
	Vencimento     *time.Time  `json:"vencimento,omitempty"`
	Corretora      string      `json:"corretora"`
	Observacoes    *string     `json:"observacoes,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// Transacao representa o histórico contábil na tabela `carteira_transacoes`.
type Transacao struct {
	ID            int64         `json:"id"`
	UsuarioID     int64         `json:"usuario_id"`
	PosicaoID     *int64        `json:"posicao_id,omitempty"`
	Ticker        string        `json:"ticker"`
	Tipo          TipoTransacao `json:"tipo"`
	Quantidade    float64       `json:"quantidade"`
	PrecoUnitario float64       `json:"preco_unitario"`
	CustosTaxas   float64       `json:"custos_taxas"`
	ValorTotal    float64       `json:"valor_total"`
	DataOperacao  time.Time     `json:"data_operacao"`
	Corretora     string        `json:"corretora"`
	Notas         *string       `json:"notas,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// CriarPosicaoRequest é o payload recebido no endpoint POST /api/v1/carteira/posicoes.
type CriarPosicaoRequest struct {
	Ticker         string      `json:"ticker"`
	Classe         ClasseAtivo `json:"classe"`
	CategoriaMacro string      `json:"categoria_macro"`
	Quantidade     float64     `json:"quantidade"`
	PrecoMedio     float64     `json:"preco_medio"`
	TaxaContratada *string     `json:"taxa_contratada,omitempty"`
	Indexador      *string     `json:"indexador,omitempty"`
	Vencimento     *string     `json:"vencimento,omitempty"` // Formato YYYY-MM-DD
	Corretora      string      `json:"corretora"`
	Observacoes    *string     `json:"observacoes,omitempty"`
	TipoOperacao   string      `json:"tipo_operacao,omitempty"` // "APORTE", "SUBSTITUIR", "COMPRA"
}

// PosicaoConsolidada enriquece a posição física com a cotação e valuation atuais de mercado.
type PosicaoConsolidada struct {
	Posicao
	CotacaoAtual         float64 `json:"cotacao_atual"`
	ValorAtualMercado    float64 `json:"valor_atual_mercado"`
	LucroPrejuizoValor   float64 `json:"lucro_prejuizo_valor"`
	RentabilidadePct     float64 `json:"rentabilidade_pct"`
	PrecoTetoEstimado    float64 `json:"preco_teto_estimado"`
	MargemSegurancaPct   float64 `json:"margem_seguranca_pct"`
	ScoreFundamentalista int     `json:"score_fundamentalista"`
	SinalRebalanceamento string  `json:"sinal_rebalanceamento"` // 'COMPRAR_APORTE',s 'AGUARDAR', 'REAVALIAR'
}

// ResumoCarteira consolida o patrimônio total e as divisões por macro-classe.
type ResumoCarteira struct {
	PatrimonioTotal  float64                 `json:"patrimonio_total"`
	TotalInvestido   float64                 `json:"total_investido"`
	LucroTotalValor  float64                 `json:"lucro_total_valor"`
	RentabilidadePct float64                 `json:"rentabilidade_pct"`
	Posicoes         []PosicaoConsolidada    `json:"posicoes"`
	AlocacaoAtual    map[ClasseAtivo]float64 `json:"alocacao_atual_pct"` // Ex: RENDA_FIXA: 42.5%
	AlocacaoAlvo     map[ClasseAtivo]float64 `json:"alocacao_alvo_pct"`  // Ex: RENDA_FIXA: 40.0%
}
