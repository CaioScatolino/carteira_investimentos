package b3

import (
	"testing"
)

func TestObterAnaliseProventos5A(t *testing.T) {
	client := NovoB3Client()

	testes := []struct {
		ticker            string
		tradingName       string
		preco             float64
		divStatusInvest12 float64
	}{
		{"PETR4", "PETROLEO BRASILEIRO S.A. PETROBRAS", 51.17, 3.67},
		{"VALE3", "VALE S.A.", 72.10, 5.61},
		{"BBAS3", "BCO BRASIL S.A.", 23.72, 0.69},
		{"ABCB4", "BCO ABC BRASIL S.A.", 26.15, 2.70},
		{"PSSA3", "PORTO SEGURO S.A.", 38.50, 2.10},
		{"CMIG4", "COMPANHIA ENERGETICA DE MINAS GERAIS - CEMIG", 11.50, 1.30},
		{"RANI3", "IRANI PAPEL E EMBALAGENS S.A.", 7.80, 0.70},
		{"KNIP11", "KINEA INDICE DE PRECOS FII", 98.20, 10.18},
		{"MXRF11", "MAXI RENDA FII", 10.15, 1.17},
		{"HGLG11", "CSHG LOGISTICA FII", 162.50, 13.20},
	}

	for _, tt := range testes {
		analise := client.ObterAnaliseProventos5A(tt.tradingName, tt.ticker, tt.preco, tt.divStatusInvest12)
		if analise == nil {
			t.Fatalf("[%s] Analise retornou nil", tt.ticker)
		}

		t.Logf("\n=== %s (%s) ===", analise.Ticker, analise.TradingName)
		t.Logf("  Total Eventos: %d", analise.TotalEventos)
		t.Logf("  Histórico Anual: %+v", analise.ProventosPorAno)
		t.Logf("  Média 5 Anos: R$ %.2f", analise.MediaDividendos5A)
		t.Logf("  Teto Bazin 5A: R$ %.2f (Margem: %+.1f%%)", analise.PrecoTetoBazin5A, analise.MargemBazin5A)
		t.Logf("  B3 12M Bruto: R$ %.2f | StatusInvest 12M: R$ %.2f", analise.Dividendos12MB3Bruto, analise.Dividendos12MStatusInvest)
		t.Logf("  Auditoria: %s (Aderência: %.1f%%)", analise.AderenciaStatusInvest, analise.AderenciaPercentual)

		if analise.TotalEventos == 0 {
			t.Errorf("[%s] Esperava eventos para ticker conhecido", tt.ticker)
		}
	}
}

func TestCotahistTradingNames(t *testing.T) {
	client := NovoB3Client()
	zipBytes, _, err := client.ObterArquivoMaisRecente()
	if err != nil {
		t.Fatalf("Erro ao obter COTAHIST: %v", err)
	}
	mapa, err := ParseCOTAHIST(zipBytes)
	if err != nil {
		t.Fatalf("Erro ao fazer parse do COTAHIST: %v", err)
	}

	t.Logf("Total de tickers únicos no COTAHIST: %d", len(mapa))

	exemplos := []string{"ABCB4", "PETR4", "VALE3", "BBAS3", "ITUB4", "KNIP11", "HGLG11", "MXRF11", "XPML11", "BTLG11", "VISC11"}
	for _, tk := range exemplos {
		if cot, ok := mapa[tk]; ok {
			t.Logf("  [%s] -> '%s'", tk, cot.NomeResumido)
		} else {
			t.Logf("  [%s] NÃO ENCONTRADO", tk)
		}
	}
}

func TestNormalizarProventos5A(t *testing.T) {
	// Caso PETR4 com a mega-distribuição de 2022 (R$ 16.55)
	provPetr4 := map[int]float64{
		2021: 5.49,
		2022: 16.55, // outlier extraordinário
		2023: 7.10,
		2024: 7.65,
		2025: 3.01,
	}
	anos := []int{2021, 2022, 2023, 2024, 2025}

	mediaBruta, mediaNormalizada, teveOutlier, obs := NormalizarProventos5A(provPetr4, anos)
	if !teveOutlier {
		t.Fatalf("Esperava que PETR4 detectasse outlier em 2022")
	}
	if mediaNormalizada >= mediaBruta {
		t.Fatalf("Média normalizada (%.2f) deveria ser menor que média bruta (%.2f)", mediaNormalizada, mediaBruta)
	}
	t.Logf("PETR4 Normalizado: Bruta=R$ %.2f -> Normalizada=R$ %.2f | Obs: %s", mediaBruta, mediaNormalizada, obs)

	// Caso Normal sem outliers (crescimento orgânico estável)
	provNormal := map[int]float64{
		2021: 1.00,
		2022: 1.20,
		2023: 1.40,
		2024: 1.60,
		2025: 1.80,
	}
	_, _, teveOutlierNormal, _ := NormalizarProventos5A(provNormal, anos)
	if teveOutlierNormal {
		t.Fatalf("Não deveria detectar outlier em série com crescimento orgânico normal")
	}
}
