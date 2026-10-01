package domain

// ObterUniversoLiquidoB3 retorna os ativos de referência mais negociados da B3
func ObterUniversoLiquidoB3() []string {
	return []string{
		// Ações Blue Chips e Dividendos
		"PETR4", "VALE3", "BBAS3", "ITUB4", "BBDC4",
		"WEGE3", "TAEE11", "CPLE3", "CSAN3", "RENT3",
		// Fundos Imobiliários mais negociados (IFIX)
		"MXRF11", "HGLG11", "XPML11", "KNRI11", "BTLG11",
		"VISC11", "XPLG11",
	}
}
