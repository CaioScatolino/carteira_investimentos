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

	// O teto ponderado com Graham descartado pelo filtro estatístico:
	// DCF (70) e Bazin 5A (65) com 41.18% cada + Gordon (62) com 17.65% = R$ 66.53
	if vale3.PrecoTetoConsolidado < 64.0 || vale3.PrecoTetoConsolidado > 68.0 {
		t.Errorf("Preço teto de VALE3 (%.2f) fora da faixa esperada (~66.53)", vale3.PrecoTetoConsolidado)
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
		PrecoAtual:       28.00,
		LPA:              3.80,
		VPA:              21.00,
		Dividendos12M:    2.40,
		PrecoTetoBazin5A: 34.00, // 50%
		PrecoTetoGordon:  36.00, // 30%
		PrecoJustoLynch:  45.00, // Descartado pelo filtro estatístico (> 1.5 DP da mediana)
		PiotroskiScore:   8,
		Pareceres:        make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(itub4)

	if itub4.PrecoTetoDCF != 0 {
		t.Errorf("DCF para banco deve ser nulo/0, obteve %.2f", itub4.PrecoTetoDCF)
	}

	// Teto com Lynch descartado: 50/80 * 34 + 30/80 * 36 = 34.75
	t.Logf("ITUB4 Teto Consensual: R$ %.2f | Modelos: %v", itub4.PrecoTetoConsolidado, itub4.ModelosTetoConsolidado)
	if itub4.PrecoTetoConsolidado < 34.0 || itub4.PrecoTetoConsolidado > 36.0 {
		t.Errorf("Teto de ITUB4 fora da faixa esperada (~34.75): %.2f", itub4.PrecoTetoConsolidado)
	}

	// Margem: (34.75 - 28.00) / 34.75 = 19.4% >= 15% -> COMPENSA_RISCO
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
