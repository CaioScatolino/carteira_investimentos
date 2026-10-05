package score

import (
	"testing"

	"carteira_investimentos/server/internal/domain"
)

func TestConsolidarPrecoTetoAcao(t *testing.T) {
	// Simulação real de PETR4 com Bazin 5A normalizado vs Graham e Gordon
	petr4 := &domain.Ativo{
		Ticker:             "PETR4",
		Classe:             domain.ClasseAcao,
		PrecoAtual:         55.83,
		LPA:                10.35,
		VPA:                37.32,
		Dividendos12M:      3.66,
		PrecoTetoBazin:     61.04,
		PrecoTetoBazin5A:   104.67, // Normalizado com Winsorização (ao invés de 134)
		ValorGraham:        93.22,
		PrecoTetoGordon:    58.88,
		PrecoJustoLynch:    258.75, // Distorcido pelo CAGR bruto, será travado prudencialmente
		Pareceres:          make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(petr4)

	if petr4.PrecoTetoConsolidado <= 0 {
		t.Fatalf("Preço teto consolidado deveria ser maior que zero")
	}

	t.Logf("PETR4 Cotação: R$ %.2f", petr4.PrecoAtual)
	t.Logf("PETR4 Preço Teto Consensual: R$ %.2f", petr4.PrecoTetoConsolidado)
	t.Logf("PETR4 Prêmio / Upside: %+.1f%%", petr4.PremioDescontoPercentual)
	t.Logf("Modelos participantes: %v", petr4.ModelosTetoConsolidado)

	// O teto consensual de PETR4 não deve ser os 134 inflados pelo pico de 2022
	if petr4.PrecoTetoConsolidado > 95.0 {
		t.Errorf("Preço teto de PETR4 (%.2f) ficou muito alto (esperado entre R$ 75 e R$ 90)", petr4.PrecoTetoConsolidado)
	}
	if petr4.PrecoTetoConsolidado < 70.0 {
		t.Errorf("Preço teto de PETR4 (%.2f) ficou muito baixo (esperado entre R$ 75 e R$ 90)", petr4.PrecoTetoConsolidado)
	}

	// Verifica se gerou parecer
	if p, ok := petr4.Pareceres["Preço Teto"]; !ok || p.Status == "" {
		t.Errorf("Deveria gerar parecer de Preço Teto")
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
		PrecoTetoBazin5A: 102.50,
		PrecoTetoBazin:   101.80,
		Pareceres:        make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(knip11)

	t.Logf("KNIP11 Cotação: R$ %.2f", knip11.PrecoAtual)
	t.Logf("KNIP11 Preço Teto Consensual: R$ %.2f", knip11.PrecoTetoConsolidado)
	t.Logf("KNIP11 Prêmio / Upside: %+.1f%%", knip11.PremioDescontoPercentual)
	t.Logf("KNIP11 Modelos: %v", knip11.ModelosTetoConsolidado)

	if knip11.PrecoTetoConsolidado <= 0 {
		t.Fatalf("Preço teto consolidado do FII deveria ser maior que zero")
	}
	if knip11.PrecoTetoConsolidado > 105.0 || knip11.PrecoTetoConsolidado < 94.0 {
		t.Errorf("Teto de FII Papel fora da faixa prudente (%.2f)", knip11.PrecoTetoConsolidado)
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
		Dividendos12M:     13.20,
		MediaDividendos5A: 13.00,
		Pareceres:         make(map[string]domain.ParecerItem),
	}

	ConsolidarPrecoTeto(hglg11)

	t.Logf("HGLG11 Cotação: R$ %.2f", hglg11.PrecoAtual)
	t.Logf("HGLG11 Preço Teto Consensual: R$ %.2f", hglg11.PrecoTetoConsolidado)
	t.Logf("HGLG11 Prêmio / Upside: %+.1f%%", hglg11.PremioDescontoPercentual)
	t.Logf("HGLG11 Modelos: %v", hglg11.ModelosTetoConsolidado)

	if hglg11.PrecoTetoConsolidado <= 0 {
		t.Fatalf("Preço teto de HGLG11 deveria ser maior que zero")
	}
	// Em Logística, com NTN-B a 6.5% + spread de 1.5% = 8.0%, Teto de Renda ~ R$ 162.50
	// Teto Patrimonial = 155.00 * 1.02 = 158.10. Consenso ponderado ~ R$ 160.74
	if hglg11.PrecoTetoConsolidado > 175.0 || hglg11.PrecoTetoConsolidado < 155.0 {
		t.Errorf("Teto de Tijolo Logístico fora da faixa prudente (%.2f)", hglg11.PrecoTetoConsolidado)
	}
}

func TestVereditoRisco(t *testing.T) {
	analisador := Novo()

	// 1. Ação com alto rendimento remunerando o risco
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

	// 2. FII com rendimento inferior à NTN-B
	fiiCaro := &domain.Ativo{
		Ticker:        "CARO11",
		Classe:        domain.ClasseFII,
		Segmento:      "Lajes Corporativas",
		PrecoAtual:    100.0,
		DY:            5.2,
		Dividendos12M: 5.20,
		VPCota:        90.0,
		Pareceres:     make(map[string]domain.ParecerItem),
	}
	_ = analisador.Executar(fiiCaro)

	if fiiCaro.VereditoRisco != "RISCO_DESCOMPENSADO" {
		t.Errorf("FII rendendo 5.2%% com NTN-B a 6.5%% deveria ser RISCO_DESCOMPENSADO, obteve %s", fiiCaro.VereditoRisco)
	}
}
