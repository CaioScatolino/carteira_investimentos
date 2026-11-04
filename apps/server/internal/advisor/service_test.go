package advisor_test

import (
	"context"
	"testing"

	"carteira_investimentos/server/internal/advisor"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/portfolio"
	"carteira_investimentos/server/internal/storage"
)

func TestAdvisor_AnalisarAporte(t *testing.T) {
	store := storage.NovoInMemoryStore()
	store.Salvar(&domain.Ativo{
		Ticker:                     "PETR4",
		Nome:                       "Petrobras",
		Classe:                     domain.ClasseAcao,
		PrecoAtual:                 30.0,
		PrecoTetoConsolidado:       45.0,
		MargemSegurancaConsolidada: 33.3,
		Score:                      95,
		DY:                         8.5,
		Porte:                      "BLUE_CHIP",
	})
	store.Salvar(&domain.Ativo{
		Ticker:                     "HGLG11",
		Nome:                       "CSHG Logística",
		Classe:                     domain.ClasseFII,
		PrecoAtual:                 150.0,
		PrecoTetoConsolidado:       180.0,
		MargemSegurancaConsolidada: 16.7,
		Score:                      90,
		PVPReal:                    0.85,
		DY:                         9.2,
		Porte:                      "FII_GIGANTE",
	})

	portfolioService := portfolio.NovoService(nil, store)
	advisorService := advisor.NovoService(portfolioService, store, "")

	req := advisor.RequisicaoAporte{
		ValorAporte:         1000.0,
		Perfil:              advisor.PerfilEquilibrado,
		PermitirNovosAtivos: true,
	}

	diag, err := advisorService.AnalisarAporte(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("erro inesperado ao analisar aporte: %v", err)
	}

	if diag.ValorAporteInformado != 1000.0 {
		t.Errorf("esperava valor de aporte 1000, obteve %f", diag.ValorAporteInformado)
	}

	if len(diag.DiagnosticoMacro) == 0 {
		t.Errorf("esperava diagnóstico macro preenchido")
	}

	if diag.ParecerSintese == "" {
		t.Errorf("esperava parecer síntese preenchido")
	}
}
