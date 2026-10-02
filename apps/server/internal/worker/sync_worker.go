package worker

import (
	"fmt"
	"strings"
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
// ExecutarSincronizacao busca os ativos, audita com os 4 motores e guarda no cache
func (w *MarketSyncWorker) ExecutarSincronizacao(universo []string) {
	inicio := time.Now()

	if len(universo) > 0 {
		fmt.Printf("\n🔄 [WORKER] A iniciar auditoria restrita de %d ativos da B3...\n", len(universo))
	} else {
		fmt.Println("\n🔄 [WORKER] A iniciar SCANNER GERAL DE MERCADO em lote (100% dos ativos da B3)...")
	}

	// 1. Busca cotações e proventos ao vivo (Yahoo Finance via Goroutines)
	ativos := w.provedor.BuscarEmLote(universo)

	// 2. Busca VPs oficiais de FIIs da CVM (em memória, 1x)
	vpsCVM, errFII := w.cvmClient.ObterValoresPatrimoniais()
	if errFII != nil {
		fmt.Printf("⚠️ [CVM FII Warning] %s\n", errFII)
	} else {
		fmt.Printf("🏛️ [CVM Oficial] %d FIIs com Valor Patrimonial oficial carregados!\n", len(vpsCVM))
	}

	// 3. Busca Balanços Oficiais de Ações da CVM (LPA e VPA dinâmicos sem mocks)
	fundamentosCVM, errAcoes := w.cvmClient.ObterFundamentosAcoes()
	if errAcoes != nil {
		fmt.Printf("⚠️ [CVM Ações Warning] %s\n", errAcoes)
	} else {
		fmt.Printf("🏛️ [CVM Oficial] %d Balanços contábeis de companhias abertas carregados!\n", len(fundamentosCVM))
	}

	// 4. Aplica os fundamentos calculados a cada ativo e executa os 4 motores
	for _, a := range ativos {
		if a.Classe == domain.ClasseFII {
			if vpOficial, ok := vpsCVM[a.Ticker]; ok {
				a.VPCota = vpOficial // Dado oficial da CVM
			}
		} else {
			// Atribui LPA e VPA calculados dinamicamente das demonstrações contábeis oficiais da CVM
			if len(fundamentosCVM) > 0 {
				cnpjAlvo := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(a.CNPJ, ".", ""), "/", ""), "-", "")

				// Se o CNPJ for vazio ou um identificador provisório (ex: 'B3-PETR4'), recorre ao mapa de CNPJs oficiais
				if len(cnpjAlvo) != 14 || strings.HasPrefix(cnpjAlvo, "B3") {
					cnpjsOficiais := map[string]string{
						"PETR4": "33000167000101", "VALE3": "33592510000154",
						"BBAS3": "00000000000191", "ITUB4": "60872504000123",
						"BBDC4": "60746948000112", "WEGE3": "84429695000111",
						"TAEE11": "07859971000130", "CPLE3": "76483817000120",
						"CSAN3": "50746577000115", "RENT3": "16670085000155",
					}
					if realCNPJ, ok := cnpjsOficiais[a.Ticker]; ok {
						cnpjAlvo = realCNPJ
					}
				}

				if f, encontrado := fundamentosCVM[cnpjAlvo]; encontrado {
					if f.LPA > 0 {
						a.LPA = f.LPA
					}
					if f.VPA > 0 {
						a.VPA = f.VPA
					}
					a.CrescimentoLucro5A = 10.0 // Média conservadora de 10% a.a. para PEG Ratio (Lynch)
				}
			}
		}

		// Roda os 4 motores analíticos
		for _, m := range w.motores {
			_ = m.Executar(a)
		}

		// Avalia o semáforo
		a.AvaliarSemaforo()
	}

	// 5. Guarda tudo no Cache (Redis / In-Memory)
	_ = w.store.SalvarLote(ativos)

	fmt.Printf("✅ [WORKER] Sincronização concluída com sucesso em %s! %d ativos auditados.\n",
		time.Since(inicio), len(ativos))
}

// IniciarLoop dispara o worker em background com time.Ticker
func (w *MarketSyncWorker) IniciarLoop(universo []string) {
	// 1. Executa a primeira sincronização de imediato ao ligar o servidor
	w.ExecutarSincronizacao(universo)

	// 2. Cria o timer nativo do Go
	ticker := time.NewTicker(w.interval)

	// Goroutine perpétua de segundo plano
	go func() {
		for range ticker.C {
			w.ExecutarSincronizacao(universo)
		}
	}()
}
