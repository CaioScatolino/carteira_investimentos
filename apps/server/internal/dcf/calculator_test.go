package dcf

import (
	"testing"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/macro"
)

func TestCalcularDCF(t *testing.T) {
	cenario := macro.CenarioMacro{
		TaxaSelic:         13.75,
		TaxaNTNB:          6.50,
		EquityRiskPremium: 5.50,
	}

	t.Run("Empresa Operacional (Real Sector)", func(t *testing.T) {
		ativo := &domain.Ativo{
			Ticker:              "TEST3",
			Classe:              domain.ClasseAcao,
			Setor:               "Consumo e Varejo",
			PrecoAtual:          30.0,
			ValorMercado:        30_000_000_000.0, // 1 bilhão de ações
			PVPReal:             2.0,              // PL = 15B
			DividaLiquidaPL:     0.5,              // DL = 7.5B, EV = 37.5B
			PEbit:               10.0,             // EBIT = 3B
			ROIC:                15.0,             // ROIC = 15%
			CrescimentoLucro5A:  8.0,              // Crescimento = 8%
			Payout:              40.0,
		}

		teto, wacc, fcff, g, err := CalcularDCF(ativo, cenario)
		if err != nil {
			t.Fatalf("Erro inesperado no DCF: %v", err)
		}
		if teto <= 0 {
			t.Errorf("Preço teto DCF esperado > 0, obteve %.2f", teto)
		}
		if wacc <= 0 || wacc > 0.25 {
			t.Errorf("WACC fora do intervalo esperado: %.4f", wacc)
		}
		if fcff <= 0 {
			t.Errorf("FCFF esperado > 0, obteve %.2f", fcff)
		}
		if g > 0.0361 {
			t.Errorf("Taxa de perpetuidade g deve ser limitada a 0.0361, obteve %.4f", g)
		}

		t.Logf("DCF TEST3: Teto = R$ %.2f | WACC = %.2f%% | FCFF = R$ %.2fM | g = %.2f%%",
			teto, wacc*100, fcff/1e6, g*100)
	})

	t.Run("Setor Financeiro (Bancos)", func(t *testing.T) {
		banco := &domain.Ativo{
			Ticker:       "ITUB4",
			Classe:       domain.ClasseAcao,
			Setor:        "Financeiro",
			Subsetor:     "Intermediários Financeiros",
			Segmento:     "Bancos",
			PrecoAtual:   35.0,
			ValorMercado: 300_000_000_000.0,
		}

		teto, _, _, _, err := CalcularDCF(banco, cenario)
		if err != nil {
			t.Fatalf("Erro inesperado para banco: %v", err)
		}
		if teto != 0 {
			t.Errorf("DCF para setor financeiro deve retornar 0, obteve %.2f", teto)
		}
	})
}
