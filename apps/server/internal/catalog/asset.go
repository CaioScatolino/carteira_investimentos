package catalog

import "time"

// Asset representa um ativo registrado no MySQL
type Asset struct {
	ID          int       `json:"id"`
	Ticker      string    `json:"ticker"`
	CNPJ        string    `json:"cnpj"`
	Classe      string    `json:"classe"` // "ACAO", "FII", "ETF"
	Tipo        string    `json:"tipo"`   // "ON", "PN", "UNT", "CI"
	RazaoSocial string    `json:"razao_social"`
	CodigoCVM   string    `json:"codigo_cvm"`
	Setor       string    `json:"setor"`
	Ativo       bool      `json:"ativo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
