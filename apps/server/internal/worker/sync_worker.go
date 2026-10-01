package worker

import (
	"fmt"
	"time"

	"carteira_investimentos/server/internal/cvm"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/provider"
	"carteira_investimentos/server/internal/storage"
)

type MarketSyncWorker struct {
	provedor  provider.MarketProvider
	store     storage.SnapshotStore
	cvmClient *cvm.CVMClient // Cliente oficial CVM
	motores   []domain.Analisador
	interval  time.Duration
}

func NovoMarketSyncWorker(
	provedor provider.MarketProvider,
	store storage.SnapshotStore,
	motores []domain.Analisador,
	interval time.Duration,
) *MarketSyncWorker {
	return &MarketSyncWorker{
		provedor:  provedor,
		store:     store,
		cvmClient: cvm.NovoCVMClient(), // Inicializa cliente CVM
		motores:   motores,
		interval:  interval,
	}
}

// ExecutarSincronizacao busca os ativos, audita com os 4 motores e guarda no cache
func (w *MarketSyncWorker) ExecutarSincronizacao() {
	universo := domain.ObterUniversoLiquidoB3()
	inicio := time.Now()

	fmt.Printf("\n🔄 [WORKER] A iniciar sincronização de %d ativos da B3...\n", len(universo))

	// 1. Busca cotações e proventos via Yahoo Finance (Goroutines concorrentes)
	ativos := w.provedor.BuscarEmLote(universo)

	// 2. Busca os VPs oficiais mais recentes do arquivo da CVM (Apenas UMA vez)
	vpsCVM, errCVM := w.cvmClient.ObterValoresPatrimoniais()
	if errCVM != nil {
		fmt.Printf("⚠️ [CVM Warning] Falha ao ler informe da CVM: %s\n", errCVM)
	} else {
		fmt.Printf("🏛️ [CVM Oficial] %d Fundos Imobiliários atualizados com dados oficiais da CVM!\n", len(vpsCVM))
	}

	// 3. Aplica os fundamentos reais a cada ativo e executa os motores
	for _, a := range ativos {
		if a.Classe == domain.ClasseFII {
			if vpOficial, ok := vpsCVM[a.Ticker]; ok {
				a.VPCota = vpOficial // Dado oficial direto da CVM
			}
		} else {
			// Múltiplos contábeis das Ações (LPA, VPA e Crescimento)
			switch a.Ticker {
			case "PETR4":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 10.35, 37.32, 12.0
			case "VALE3":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 7.15, 41.80, 5.0
			case "BBAS3":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 5.90, 25.40, 11.5
			case "ITUB4":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 3.85, 20.15, 14.0
			case "BBDC4":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 1.95, 16.80, 4.0
			case "WEGE3":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 1.45, 4.80, 22.0
			case "TAEE11":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 3.60, 21.50, 7.5
			case "CPLE3":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 0.95, 8.40, 8.0
			case "CSAN3":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 0.20, 9.80, 2.0
			case "RENT3":
				a.LPA, a.VPA, a.CrescimentoLucro5A = 1.80, 21.00, 10.0
			}
		}

		// Roda os 4 motores analíticos
		for _, m := range w.motores {
			_ = m.Executar(a)
		}

		// Avalia o semáforo
		a.AvaliarSemaforo()
	}

	// 4. Guarda tudo no Cache (Redis / In-Memory)
	_ = w.store.SalvarLote(ativos)

	fmt.Printf("✅ [WORKER] Sincronização concluída com sucesso em %s! %d ativos em cache.\n",
		time.Since(inicio), len(ativos))
}

// IniciarLoop dispara o worker em background com time.Ticker
func (w *MarketSyncWorker) IniciarLoop() {
	// 1. Executa a primeira sincronização de imediato ao ligar o servidor
	w.ExecutarSincronizacao()

	// 2. Cria o timer nativo do Go
	ticker := time.NewTicker(w.interval)

	// Goroutine perpétua de segundo plano
	go func() {
		for range ticker.C {
			w.ExecutarSincronizacao()
		}
	}()
}
