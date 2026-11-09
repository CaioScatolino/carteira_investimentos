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
	Vencimento     *string     `json:"vencimento,omitempty"`    // Formato YYYY-MM-DD
	Corretora      string      `json:"corretora"`
	Observacoes    *string     `json:"observacoes,omitempty"`
	TipoOperacao   string      `json:"tipo_operacao,omitempty"` // "APORTE", "SUBSTITUIR", "COMPRA"
	DataOperacao   *string     `json:"data_operacao,omitempty"` // Data de início/aquisição (YYYY-MM-DD)
}

// PosicaoConsolidada enriquece a posição física com a cotação, valuation e proventos acumulados.
type PosicaoConsolidada struct {
	Posicao
	NomeAtivo               string  `json:"nome_ativo,omitempty"`
	Setor                   string  `json:"setor,omitempty"`
	Subsetor                string  `json:"subsetor,omitempty"`
	Segmento                string  `json:"segmento,omitempty"`
	CotacaoAtual            float64 `json:"cotacao_atual"`
	ValorAtualMercado       float64 `json:"valor_atual_mercado"`
	LucroPrejuizoValor      float64 `json:"lucro_prejuizo_valor"`
	RentabilidadePct        float64 `json:"rentabilidade_pct"`
	PrecoTetoEstimado       float64 `json:"preco_teto_estimado"`
	MargemSegurancaPct      float64 `json:"margem_seguranca_pct"`
	ScoreFundamentalista    int     `json:"score_fundamentalista"`
	SinalRebalanceamento    string  `json:"sinal_rebalanceamento"`   // 'COMPRAR_APORTE', 'AGUARDAR', 'REAVALIAR'
	PercentualNaClasse      float64 `json:"percentual_na_classe"`     // Peso relativo do ativo dentro da sua classe (ex: 50.0% das Ações)
	PercentualNaCarteira    float64 `json:"percentual_na_carteira"`   // Peso absoluto do ativo na carteira inteira (ex: 10.0% do Total)
	TotalProventos          float64 `json:"total_proventos"`          // Total em R$ de proventos recebidos deste ativo
	PrecoMedioDescontado    float64 `json:"preco_medio_descontado"`   // Preço Médio subtraindo os proventos recebidos por cota
	LucroComProventos       float64 `json:"lucro_com_proventos"`      // Lucro Total = Ganho de Capital + Proventos
	RentabilidadeTotalPct   float64 `json:"rentabilidade_total_pct"`  // Rentabilidade % considerando Valorização + Proventos
	ProventosYieldOnCostPct float64 `json:"proventos_yoc_pct"`       // Yield on Cost Histórico (% de proventos sobre o capital investido)
}

// SetorConsolidado agrupa os ativos por setor econômico / segmento oficial da B3.
type SetorConsolidado struct {
	Nome                 string      `json:"nome"`
	ClassePrincipal      ClasseAtivo `json:"classe_principal"`
	ValorAtualMercado    float64     `json:"valor_atual_mercado"`
	PercentualNaCarteira float64     `json:"percentual_na_carteira"` // % do patrimônio total (ex: 28.5%)
	QuantidadeAtivos     int         `json:"quantidade_ativos"`
	Tickers              []string    `json:"tickers"`                // Lista de tickers participantes
	StatusConcentracao   string      `json:"status_concentracao"`    // "DIVERSIFICADO", "MODERADO", "SOBRECONCENTRADO"
}

// AnaliseConcentracaoHHI consolida as métricas do Índice Herfindahl-Hirschman e alertas prudenciais.
type AnaliseConcentracaoHHI struct {
	HHIAtivos            float64            `json:"hhi_ativos"`             // 0 a 10.000 (Soma dos pesos dos ativos ao quadrado)
	ClassificacaoAtivos  string             `json:"classificacao_ativos"`  // "BEM_DIVERSIFICADO" (< 1500), "MODERADO" (1500-2500), "CONCENTRADO" (>= 2500)
	DescricaoHHIAtivos   string             `json:"descricao_hhi_ativos"`  // Texto explicativo institucional
	HHISetores           float64            `json:"hhi_setores"`            // 0 a 10.000 (Soma dos pesos dos setores ao quadrado)
	ClassificacaoSetores string             `json:"classificacao_setores"` // "BEM_DIVERSIFICADO", "MODERADO", "CONCENTRADO"
	DescricaoHHISetores  string             `json:"descricao_hhi_setores"` // Texto explicativo institucional
	MaiorAtivoTicker     string             `json:"maior_ativo_ticker"`     // Ticker de maior peso
	MaiorAtivoPercentual float64            `json:"maior_ativo_pct"`        // % da carteira do maior ativo
	MaiorSetorNome       string             `json:"maior_setor_nome"`       // Nome do maior setor
	MaiorSetorPercentual float64            `json:"maior_setor_pct"`        // % da carteira do maior setor
	Setores              []SetorConsolidado `json:"setores"`                // Lista ordenada por maior exposição financeira
	Alertas              []string           `json:"alertas"`                // Alertas pedagógicos de risco
}

// ClasseConsolidada agrupa os dados de uma macro-classe para a visão hierárquica por níveis e gráficos de composição.
type ClasseConsolidada struct {
	Classe               ClasseAtivo          `json:"classe"`
	Nome                 string               `json:"nome"`
	ValorAtualMercado    float64              `json:"valor_atual_mercado"`
	TotalInvestido       float64              `json:"total_investido"`
	LucroPrejuizoValor   float64              `json:"lucro_prejuizo_valor"`
	RentabilidadePct     float64              `json:"rentabilidade_pct"`
	PercentualNaCarteira float64              `json:"percentual_na_carteira"` // % do patrimônio total
	PercentualAlvo       float64              `json:"percentual_alvo"`         // Meta institucional
	QuantidadeAtivos     int                  `json:"quantidade_ativos"`
	Posicoes             []PosicaoConsolidada `json:"posicoes"`
	TotalProventos       float64              `json:"total_proventos"`        // Total em R$ de proventos desta classe
	LucroComProventos    float64              `json:"lucro_com_proventos"`    // Lucro somando proventos da classe
}

// ResumoCarteira consolida o patrimônio total e as divisões por macro-classe.
type ResumoCarteira struct {
	PatrimonioTotal         float64                 `json:"patrimonio_total"`
	TotalInvestido          float64                 `json:"total_investido"`
	LucroTotalValor         float64                 `json:"lucro_total_valor"`
	RentabilidadePct        float64                 `json:"rentabilidade_pct"`
	Posicoes                []PosicaoConsolidada    `json:"posicoes"`
	Classes                 []ClasseConsolidada     `json:"classes"`
	AlocacaoAtual           map[ClasseAtivo]float64 `json:"alocacao_atual_pct"` // Ex: RENDA_FIXA: 42.5%
	AlocacaoAlvo            map[ClasseAtivo]float64 `json:"alocacao_alvo_pct"`  // Ex: RENDA_FIXA: 40.0%
	TotalProventosRecebidos float64                 `json:"total_proventos_recebidos"`
	LucroTotalComProventos  float64                 `json:"lucro_total_com_proventos"`
	RentabilidadeTotalPct   float64                 `json:"rentabilidade_total_pct"`
	AnaliseConcentracao     *AnaliseConcentracaoHHI `json:"analise_concentracao,omitempty"`
}
