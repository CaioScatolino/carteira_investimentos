package domain

// Analisador define o contrato universal que todo motor de valuation deve cumprir
type Analisador interface {
	Nome() string
	Executar(ativo *Ativo) error
}
