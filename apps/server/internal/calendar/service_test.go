package calendar

import (
	"context"
	"testing"

	"carteira_investimentos/server/internal/b3"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/storage"
)

func TestCalendarioProventosSazonalidade(t *testing.T) {
	store := storage.NovoInMemoryStore()
	b3Client := b3.NovoB3Client()

	vale3 := &domain.Ativo{
		Ticker:     "VALE3",
		Nome:       "VALE S.A.",
		Classe:     domain.ClasseAcao,
		Setor:      "Materiais Básicos",
		PrecoAtual: 55.00,
		Score:      88.0,
		DY:         8.5,
	}

	itub4 := &domain.Ativo{
		Ticker:     "ITUB4",
		Nome:       "ITAU UNIBANCO HOLDING S.A.",
		Classe:     domain.ClasseAcao,
		Setor:      "Financeiro",
		PrecoAtual: 33.00,
		Score:      92.0,
		DY:         6.8,
	}

	_ = store.Salvar(vale3)
	_ = store.Salvar(itub4)

	serv := NovoService(store, b3Client)

	filtro := FiltroCalendario{
		TipoData: "DATA_COM",
	}

	res, err := serv.ObterCalendario(context.Background(), filtro)
	if err != nil {
		t.Fatalf("Erro ao obter calendário: %v", err)
	}

	if len(res.Ativos) == 0 {
		t.Fatalf("Esperava ativos no resultado do calendário")
	}

	t.Logf("Total Ativos no Calendário: %d", res.TotalAtivos)

	for _, at := range res.Ativos {
		t.Logf("\n=== %s (%s) | Score: %.0f | Maior Prob: %.1f%% em %s ===",
			at.Ticker, at.Nome, at.Score, at.MaiorProbabilidade, at.MesMaisProvavel)

		mesesComPagamento := 0
		for _, m := range at.Meses {
			if m.Probabilidade > 0 {
				mesesComPagamento++
				t.Logf("  [%02d - %-10s]: %5.1f%% (anos: %v)", m.Mes, m.NomeMes, m.Probabilidade, m.AnosComProvento)
			}
		}

		if at.Ticker == "ITUB4" {
			// Itaú deve pagar em praticamente todos os 12 meses
			if mesesComPagamento < 10 {
				t.Errorf("ITUB4 deveria ter pelo menos 10 meses pagadores, obteve %d", mesesComPagamento)
			}
		}

		if at.Ticker == "VALE3" {
			// VALE3 deve ter pelo menos 2 a 4 meses com alta probabilidade (Março, Agosto, Dezembro)
			mar := at.Meses[2]  // Março
			ago := at.Meses[7]  // Agosto
			dez := at.Meses[11] // Dezembro
			t.Logf("VALE3 Março: %.1f%% | Agosto: %.1f%% | Dezembro: %.1f%%", mar.Probabilidade, ago.Probabilidade, dez.Probabilidade)
			if mar.Probabilidade == 0 && ago.Probabilidade == 0 && dez.Probabilidade == 0 {
				t.Errorf("VALE3 deveria ter pagamentos históricos registrados em Março, Agosto ou Dezembro")
			}
		}
	}
}
