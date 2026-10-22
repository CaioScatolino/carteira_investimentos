package score

import (
	"testing"

	"carteira_investimentos/server/internal/domain"
)

func TestConsolidarPrecoTetoAcaoOperacional(t *testing.T) {
	// 1. Simulação de empresa do Setor Real (Operacional) com DCF Proxy, Bazin 5A, Graham e Gordon
	vale3 := &domain.Ativo{
		Ticker:             "VALE3",
		Classe:             domain.ClasseAcao,
		Setor:              "Materiais Básicos",
		PrecoAtual:         50.00,  // Preço que garante margem > 15% sobre o teto de ~66.53
		LPA:                9.50,
		VPA:                42.00,
		Dividendos12M:      5.20,
		PrecoTetoDCF:       70.00,  // DCF Proxy (peso 35%)
		PrecoTetoBazin5A:   65.00,  // Bazin 5A (peso 35%)
		ValorGraham:        94.70,  // Graham outlier descartado pelo filtro de 1.5 DP da mediana
		PrecoTetoGordon:    62.00,  // Gordon DDM (peso 15%)
		PiotroskiScore:     7,      // Alta solidez financeira
		Pareceres:          make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(vale3)

	if vale3.PrecoTetoConsolidado <= 0 {
		t.Fatalf("Preço teto consolidado deveria ser maior que zero")
	}

	t.Logf("VALE3 Cotação: R$ %.2f", vale3.PrecoAtual)
	t.Logf("VALE3 Preço Teto Consensual: R$ %.2f", vale3.PrecoTetoConsolidado)
	t.Logf("VALE3 Prêmio / Upside: %+.1f%%", vale3.PremioDescontoPercentual)
	t.Logf("VALE3 Modelos participantes: %v", vale3.ModelosTetoConsolidado)

	// O teto ponderado com Graham de Ciclo integrado:
	// DCF Proxy (70 - 35%) + Bazin 5A (65 - 35%) + Graham de Ciclo (75.60 - 30%) = R$ 69.93
	if vale3.PrecoTetoConsolidado < 64.0 || vale3.PrecoTetoConsolidado > 72.0 {
		t.Errorf("Preço teto de VALE3 (%.2f) fora da faixa esperada (~69.93)", vale3.PrecoTetoConsolidado)
	}

	// Com cotação a R$ 50.00, a margem de segurança é de 24.8% (supera o piso de 15% para Piotroski >= 7)
	if vale3.VereditoRisco != "COMPENSA_RISCO" {
		t.Errorf("Esperado COMPENSA_RISCO para VALE3 com Piotroski 7 e margem > 15%%, obteve %s", vale3.VereditoRisco)
	}
}

func TestConsolidarPrecoTetoBanco(t *testing.T) {
	// 2. Simulação de Banco / Setor Financeiro (Sem DCF, ponderação 50% Bazin 5A + 30% Gordon + 20% Lynch)
	itub4 := &domain.Ativo{
		Ticker:           "ITUB4",
		Classe:           domain.ClasseAcao,
		Setor:            "Financeiro",
		Subsetor:         "Intermediários Financeiros",
		Segmento:         "Bancos",
		PrecoAtual:       26.00, // Cotação que garante margem institucional de segurança (> 15%)
		LPA:              3.80,
		VPA:              21.00,
		Dividendos12M:    2.40,
		PrecoTetoBazin5A: 34.00,
		PrecoTetoGordon:  36.00,
		PrecoJustoLynch:  45.00,
		PiotroskiScore:   8,
		Pareceres:        make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(itub4)

	if itub4.PrecoTetoDCF != 0 {
		t.Errorf("DCF para banco deve ser nulo/0, obteve %.2f", itub4.PrecoTetoDCF)
	}

	// Teto Consensual Bancário: Bazin 5A (40%) + Gordon DDM (30%) + Justified P/B (30%) = ~R$ 31.99
	t.Logf("ITUB4 Teto Consensual: R$ %.2f | Modelos: %v", itub4.PrecoTetoConsolidado, itub4.ModelosTetoConsolidado)
	if itub4.PrecoTetoConsolidado < 30.0 || itub4.PrecoTetoConsolidado > 35.0 {
		t.Errorf("Teto de ITUB4 fora da faixa esperada (~32.00): %.2f", itub4.PrecoTetoConsolidado)
	}

	// Margem: (31.99 - 26.00) / 31.99 = 18.7% >= 15% -> COMPENSA_RISCO
	if itub4.VereditoRisco != "COMPENSA_RISCO" {
		t.Errorf("Esperava COMPENSA_RISCO para ITUB4, obteve %s", itub4.VereditoRisco)
	}
}

func TestPiotroskiThresholds(t *testing.T) {
	// Ação com Piotroski <= 4 deve ter RISCO_DESCOMPENSADO independente de margem
	empresaProblematica := &domain.Ativo{
		Ticker:           "RISK3",
		Classe:           domain.ClasseAcao,
		Setor:            "Varejo",
		PrecoAtual:       10.00,
		PrecoTetoDCF:     20.00,
		PrecoTetoBazin5A: 18.00,
		PiotroskiScore:   3, // F-Score crítico
		Pareceres:        make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(empresaProblematica)

	if empresaProblematica.VereditoRisco != "RISCO_DESCOMPENSADO" {
		t.Errorf("Empresa com Piotroski <= 4 deve ser RISCO_DESCOMPENSADO, obteve %s", empresaProblematica.VereditoRisco)
	}

	// Ação com Piotroski entre 5 e 6: Margem mínima é de 25%
	empresaModerada := &domain.Ativo{
		Ticker:           "MOD3",
		Classe:           domain.ClasseAcao,
		Setor:            "Indústria",
		PrecoAtual:       80.00,
		PrecoTetoDCF:     100.00,
		PrecoTetoBazin5A: 100.00,
		PiotroskiScore:   5, // Margem de (100 - 80) / 100 = 20% (< 25%)
		Pareceres:        make(map[string]domain.ParecerItem),
	}
	ConsolidarPrecoTeto(empresaModerada)

	if empresaModerada.VereditoRisco != "NEUTRO" {
		t.Errorf("Empresa com F-Score 5 e margem de 20%% (< 25%%) deveria ser NEUTRO, obteve %s", empresaModerada.VereditoRisco)
	}
}

func TestConsolidarPrecoTetoFII(t *testing.T) {
	// Simulação real de FII de Papel (KNIP11)
	knip11 := &domain.Ativo{
		Ticker:           "KNIP11",
		Classe:           domain.ClasseFII,
		Segmento:         "Papel / CRI",
		PrecoAtual:       93.50,
		VPCota:           94.20,
		PVP:              0.99,
		DY:               12.5,
		PrecoTetoBazin5A: 102.50,
		Pareceres:        make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(knip11)

	t.Logf("KNIP11 Cotação: R$ %.2f", knip11.PrecoAtual)
	t.Logf("KNIP11 Preço Teto Consensual: R$ %.2f", knip11.PrecoTetoConsolidado)
	t.Logf("KNIP11 Prêmio / Upside: %+.1f%%", knip11.PremioDescontoPercentual)

	if knip11.PrecoTetoConsolidado != 94.20 {
		t.Errorf("Teto de FII Papel deve ser exatamente o VP Cota (94.20), obteve %.2f", knip11.PrecoTetoConsolidado)
	}

	// Teste de ágio em FII de papel (> 1.02)
	papelAgio := &domain.Ativo{
		Ticker:     "AGIO11",
		Classe:     domain.ClasseFII,
		Segmento:   "Papel / CRI",
		PrecoAtual: 105.00,
		VPCota:     100.00,
		PVP:        1.05,
		DY:         11.0,
		Pareceres:  make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(papelAgio)
	if papelAgio.AlertaRisco != "Ágio em FII de Papel" {
		t.Errorf("Esperava alerta 'Ágio em FII de Papel', obteve '%s'", papelAgio.AlertaRisco)
	}
	if papelAgio.VereditoRisco != "RISCO_DESCOMPENSADO" {
		t.Errorf("Esperava RISCO_DESCOMPENSADO para FII com ágio > 1.02, obteve %s", papelAgio.VereditoRisco)
	}
}

func TestConsolidarPrecoTetoFIITijolo(t *testing.T) {
	// Simulação real de FII de Logística (HGLG11)
	hglg11 := &domain.Ativo{
		Ticker:            "HGLG11",
		Classe:            domain.ClasseFII,
		Segmento:          "Logística",
		PrecoAtual:        162.50,
		VPCota:            155.00,
		PVP:               1.01,
		DY:                8.2,
		Dividendos12M:     13.20,
		MediaDividendos5A: 13.00,
		Pareceres:         make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(hglg11)

	t.Logf("HGLG11 Cotação: R$ %.2f", hglg11.PrecoAtual)
	t.Logf("HGLG11 Preço Teto Consensual: R$ %.2f", hglg11.PrecoTetoConsolidado)
	t.Logf("HGLG11 Modelos: %v", hglg11.ModelosTetoConsolidado)

	// Em Logística, NTN-B (6.5%) + spread (1.5%) = 8.0%.
	// Teto = 13.20 / 0.08 = R$ 165.00
	if hglg11.PrecoTetoConsolidado != 165.00 {
		t.Errorf("Teto de Tijolo Logístico esperado 165.00, obteve %.2f", hglg11.PrecoTetoConsolidado)
	}
}

func TestVereditoRisco(t *testing.T) {
	analisador := Novo()

	bbas3 := &domain.Ativo{
		Ticker:        "BBAS3",
		Classe:        domain.ClasseAcao,
		PrecoAtual:    26.50,
		DY:            9.8,
		Dividendos12M: 2.60,
		LPA:           5.20,
		VPA:           32.10,
		Pareceres:     make(map[string]domain.ParecerItem),
	}
	_ = analisador.Executar(bbas3)

	if bbas3.VereditoRisco != "COMPENSA_RISCO" {
		t.Errorf("BBAS3 com DY de 9.8%% deveria ter veredito COMPENSA_RISCO, obteve %s", bbas3.VereditoRisco)
	}
	if bbas3.SpreadNTNB < 3.0 {
		t.Errorf("BBAS3 SpreadNTNB deveria ser >= 3.0, obteve %.2f", bbas3.SpreadNTNB)
	}
}

func TestValuationBBAS3Institucional(t *testing.T) {
	analisador := Novo()

	// 1. Cenário Real com Série Auditada B3 (R$ 2.69/ano e ROE de 18.5%)
	bbas3 := &domain.Ativo{
		Ticker:                       "BBAS3",
		Nome:                         "BANCO DO BRASIL S.A.",
		Classe:                       domain.ClasseAcao,
		Setor:                        "Financeiro e Outros",
		Subsetor:                     "Intermediários Financeiros",
		Segmento:                     "Bancos",
		PrecoAtual:                   23.83,
		LPA:                          5.60,
		VPA:                          32.47,
		ROE:                          18.5,
		DY:                           10.8,
		Dividendos12M:                2.69,
		MediaDividendos5A:            2.69,
		MediaDividendos5ANormalizada: 2.69,
		PrecoTetoGordon:              28.50,
		ValorGraham:                  63.90,
		Pareceres:                    make(map[string]domain.ParecerItem),
	}

	_ = analisador.Executar(bbas3)
	ConsolidarPrecoTeto(bbas3)

	t.Logf("BBAS3 Normalizado Teto Consensual: R$ %.2f | Modelos: %v", bbas3.PrecoTetoConsolidado, bbas3.ModelosTetoConsolidado)

	if bbas3.PrecoTetoConsolidado < 24.0 || bbas3.PrecoTetoConsolidado > 36.0 {
		t.Errorf("Teto de BBAS3 (%.2f) fora da faixa institucional (R$ 24 a R$ 36)", bbas3.PrecoTetoConsolidado)
	}

	// 2. Cenário de Estresse (Scraper com dados truncados para R$ 0.69)
	bbas3Truncado := &domain.Ativo{
		Ticker:        "BBAS3",
		Nome:          "BANCO DO BRASIL S.A.",
		Classe:        domain.ClasseAcao,
		Setor:         "Financeiro e Outros",
		Subsetor:      "Intermediários Financeiros",
		Segmento:      "Bancos",
		PrecoAtual:    23.83,
		LPA:           3.25, // Lucro recorrente
		VPA:           32.47,
		ROE:           10.5,
		Dividendos12M: 0.69,
		Pareceres:     make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(bbas3Truncado)
	t.Logf("BBAS3 Truncado Teto com Piso de Sanidade: R$ %.2f | Modelos: %v", bbas3Truncado.PrecoTetoConsolidado, bbas3Truncado.ModelosTetoConsolidado)

	if bbas3Truncado.PrecoTetoConsolidado < 18.0 {
		t.Errorf("Piso de sanidade falhou! Teto (%.2f) não deveria cair abaixo de R$ 18.00", bbas3Truncado.PrecoTetoConsolidado)
	}
}

func TestScoreMultiFatorSLCE3(t *testing.T) {
	analisador := Novo()

	// Simulação real de SLCE3 (SLC Agrícola):
	// DY moderado (5.3%), mas ROIC robusto, P/VP descontado, alavancagem saudável
	slce3 := &domain.Ativo{
		Ticker:              "SLCE3",
		Nome:                "SLC AGRICOLA S.A.",
		Classe:              domain.ClasseAcao,
		Setor:               "Consumo não Cíclico",
		Subsetor:            "Agropecuária",
		Segmento:            "Agricultura",
		PrecoAtual:          16.99,
		LPA:                 2.05,  // P/L = 8.3x -> Earnings Yield = 12.1%
		VPA:                 19.50, // P/VP = 0.87x (desconto patrimonial em terras)
		Dividendos12M:       0.90,  // DY = 5.3%
		MediaDividendos5A:   1.15,  // Média 5A = 6.8%
		ROIC:                14.5,  // ROIC de 14.5%
		ROE:                 10.5,
		MargemLiquida:       12.0,
		CrescimentoLucro5A:  6.0,
		DividaLiquidaEbit:   1.8,   // Alavancagem saudável < 2.5x
		PiotroskiScore:      6,     // Solidez financeira
		Porte:               "MID_CAP",
		LiquidezMediaDiaria: 35_000_000.0, // Alta liquidez
		Pareceres:           make(map[string]domain.ParecerItem),
	}

	_ = analisador.Executar(slce3)

	t.Logf("SLCE3 Score Calculado: %.0f/100", slce3.Score)
	t.Logf("SLCE3 Veredito de Risco: %s", slce3.VereditoRisco)
	t.Logf("SLCE3 Earnings Yield: %.1f%%", slce3.EarningsYield)
	t.Logf("SLCE3 PrecoTetoConsolidado: R$ %.2f", slce3.PrecoTetoConsolidado)
	t.Logf("SLCE3 MargemSegurancaConsolidada: %.1f%%", slce3.MargemSegurancaConsolidada)
	t.Logf("SLCE3 Modelos: %v", slce3.ModelosTetoConsolidado)

	// Com o novo motor institucional de 4 pilares:
	// SLCE3 não deve ficar com score 30 (antigo). Deve ficar com score >= 70!
	if slce3.Score < 70 {
		t.Errorf("SLCE3 deveria ter Score >= 70 no novo modelo de 4 pilares, mas obteve %.0f", slce3.Score)
	}

	// Com Earnings Yield de 12.1% vs NTN-B 6.5%, o Veredito de Risco não deve ser RISCO_DESCOMPENSADO!
	if slce3.VereditoRisco == "RISCO_DESCOMPENSADO" {
		t.Errorf("SLCE3 com Earnings Yield de 12.1%% não deve ter RISCO_DESCOMPENSADO, obteve %s", slce3.VereditoRisco)
	}
}

func TestSizeFactorEBlindagemMicrocapALLD3(t *testing.T) {
	analisador := Novo()

	// 1. ALLD3 (Allied Tecnologia): Microcap com P/L artificial de 1.86x por ganhos tributários,
	// DY atípico de 14.8% e baixa liquidez diária (~R$ 380k/dia)
	alld3 := &domain.Ativo{
		Ticker:              "ALLD3",
		Nome:                "ALLIED TECNOLOGIA S.A.",
		Classe:              domain.ClasseAcao,
		Setor:               "Consumo Cíclico",
		Subsetor:            "Comércio",
		Segmento:            "Eletrodomésticos",
		PrecoAtual:          5.80,
		LPA:                 3.12,  // P/L = 1.86x (Trap quantitativo)
		VPA:                 13.53, // P/VP = 0.43x
		Dividendos12M:       0.86,  // DY = 14.8%
		MediaDividendos5A:   0.45,
		ROIC:                16.0,
		ROE:                 23.0,
		ValorMercado:        670_000_000.0, // R$ 670M (Microcap)
		LiquidezMediaDiaria: 380_000.0,     // R$ 380k/dia (Baixa Liquidez)
		PiotroskiScore:      7,
		Pareceres:           make(map[string]domain.ParecerItem),
	}

	// 2. PETR4 (Petrobras): Blue Chip de R$ 450 Bilhões com liquidez de R$ 1.5 Bi/dia
	petr4 := &domain.Ativo{
		Ticker:              "PETR4",
		Nome:                "PETROLEO BRASILEIRO S.A. PETROBRAS",
		Classe:              domain.ClasseAcao,
		Setor:               "Petróleo, Gás e Biocombustíveis",
		PrecoAtual:          36.50,
		LPA:                 9.20,  // P/L = 3.97x (Faixa saudável institucional)
		VPA:                 33.00, // P/VP = 1.10x
		Dividendos12M:       5.50,  // DY = 15.0%
		MediaDividendos5A:   5.10,
		ROIC:                22.0,
		ROE:                 27.8,
		MargemLiquida:       24.0,
		CrescimentoLucro5A:  8.0,
		DividaLiquidaEbit:   0.8,
		ValorMercado:        450_000_000_000.0,
		LiquidezMediaDiaria: 1_200_000_000.0,
		PiotroskiScore:      8,
		Pareceres:           make(map[string]domain.ParecerItem),
	}

	_ = analisador.Executar(alld3)
	_ = analisador.Executar(petr4)

	t.Logf("ALLD3 Porte: %s | Score: %.0f/100 | Teto: R$ %.2f", alld3.Porte, alld3.Score, alld3.PrecoTetoConsolidado)
	t.Logf("PETR4 Porte: %s | Score: %.0f/100 | Teto: R$ %.2f", petr4.Porte, petr4.Score, petr4.PrecoTetoConsolidado)

	// Validação 1: ALLD3 deve ser classificada como MICRO_CAP
	if alld3.Porte != "MICRO_CAP" {
		t.Errorf("ALLD3 deveria ser MICRO_CAP, obteve %s", alld3.Porte)
	}

	// Validação 2: ALLD3 deve estar blindada pelo teto de liquidez (Score <= 65 ou <= 74)
	if alld3.Score > 74 {
		t.Errorf("ALLD3 não pode ultrapassar o teto de 74 pts para microcaps! Obteve %.0f", alld3.Score)
	}

	// Validação 3: PETR4 deve ser classificada como BLUE_CHIP e pontuar fortemente no Topo
	if petr4.Porte != "BLUE_CHIP" {
		t.Errorf("PETR4 deveria ser BLUE_CHIP, obteve %s", petr4.Porte)
	}
	if petr4.Score < 85 {
		t.Errorf("PETR4 deveria ter Score >= 85, obteve %.0f", petr4.Score)
	}

	// Validação 4: PETR4 (Blue Chip consolidada) DEVE superar ALLD3 no ranking
	if petr4.Score <= alld3.Score {
		t.Errorf("PETR4 (%.0f) deveria ter pontuação superior a ALLD3 (%.0f) após o ajuste de porte", petr4.Score, alld3.Score)
	}
}
