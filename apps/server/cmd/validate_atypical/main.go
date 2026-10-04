package main

import (
	"context"
	"fmt"
	"time"

	"carteira_investimentos/server/internal/bazin"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/fii"
	"carteira_investimentos/server/internal/gordon"
	"carteira_investimentos/server/internal/graham"
	"carteira_investimentos/server/internal/greenblatt"
	"carteira_investimentos/server/internal/lynch"
	"carteira_investimentos/server/internal/piotroski"
	"carteira_investimentos/server/internal/score"
	"carteira_investimentos/server/internal/statusinvest"
	"carteira_investimentos/server/internal/storage"
)

func main() {
	motores := []domain.Analisador{
		bazin.Novo(0.06),
		graham.Novo(),
		lynch.Novo(),
		gordon.Novo(0.11),
		greenblatt.Novo(),
		piotroski.Novo(),
		fii.Novo(0.065),
		score.Novo(),
	}

	store := storage.NovoInMemoryStore()
	service := statusinvest.NovoIngestionService(nil, store, motores)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := service.ExecutarSincronizacao(ctx)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	tickers := []string{"RIAA3", "GRND3", "HBRE3", "SCAR3", "LOGG3", "PATL11", "BBFI11", "HGPO11", "JCIN11"}
	fmt.Printf("\n%-7s | %-12s | %-8s | %-8s | %-10s | %-10s | %-6s | %-12s | %s\n",
		"TICKER", "CLASSE", "PRECO", "DY", "BAZIN", "PAYOUT", "SCORE", "STATUS", "ALERTA")
	fmt.Println("-------------------------------------------------------------------------------------------------------------------------")

	for _, t := range tickers {
		ativo, err := store.Obter(t)
		if err != nil {
			fmt.Printf("%-7s | NÃO ENCONTRADO\n", t)
			continue
		}
		fmt.Printf("%-7s | %-12s | R$%6.2f | %6.1f%% | R$%8.2f | %7.1f%% | %5.0f  | %-12s | %s\n",
			ativo.Ticker, ativo.Classe, ativo.PrecoAtual, ativo.DY, ativo.PrecoTetoBazin, ativo.Payout, ativo.Score, ativo.Status, ativo.AlertaRisco)
	}
}
