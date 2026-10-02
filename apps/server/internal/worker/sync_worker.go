package worker

import (
	"fmt"
	"strings"
	"time"

	"carteira_investimentos/server/internal/b3"
	"carteira_investimentos/server/internal/cvm"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/provider"
	"carteira_investimentos/server/internal/storage"
)

type MarketSyncWorker struct {
	provedor  provider.MarketProvider
	store     storage.SnapshotStore
	cvmClient *cvm.CVMClient // Cliente oficial CVM
	b3Client  *b3.B3Client   // Cliente oficial B3 para eventos corporativos em dinheiro
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
		b3Client:  b3.NovoB3Client(),   // Inicializa cliente B3
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

	// 2. Busca Dados Oficiais de FIIs da CVM (VP/Cota + Rendimentos 12M em memória, 1x)
	dadosFII, errFII := w.cvmClient.ObterDadosFII()
	if errFII != nil {
		fmt.Printf("⚠️ [CVM FII Warning] %s\n", errFII)
	} else {
		fmt.Printf("🏛️ [CVM Oficial] %d FIIs com Valor Patrimonial e Proventos carregados!\n", len(dadosFII))
	}

	// 3. Busca Balanços Oficiais de Ações da CVM (LPA e VPA dinâmicos sem mocks)
	fundamentosCVM, errAcoes := w.cvmClient.ObterFundamentosAcoes()
	if errAcoes != nil {
		fmt.Printf("⚠️ [CVM Ações Warning] %s\n", errAcoes)
	} else {
		fmt.Printf("🏛️ [CVM Oficial] %d Balanços contábeis de companhias abertas carregados!\n", len(fundamentosCVM))
	}

	// Carrega cadastro oficial de companhias da CVM para resolução dinâmica de CNPJ
	cadastroCVM, _ := w.cvmClient.ObterCadastroCVM()
	if cadastroCVM != nil {
		fmt.Println("📋 [CVM Cadastro] Cadastro oficial de companhias abertas carregado com sucesso!")
	}

	// 4. Aplica os fundamentos calculados a cada ativo e executa os 4 motores
	for _, a := range ativos {
		if a.Classe == domain.ClasseFII {
			// Atribui VP/Cota e Proventos 12M apurados diretamente da CVM
			if f, ok := dadosFII[a.Ticker]; ok {
				a.VPCota = f.VPCota
				a.Dividendos12M = f.Dividendos12M
			} else {
				// Resolução dinâmica por similaridade de nome para FIIs com código divergente (ex: KNUQ11)
				limpoB3 := strings.TrimPrefix(strings.TrimPrefix(strings.ToUpper(a.Nome), "FII "), "FDO ")
				palavrasB3 := strings.Fields(limpoB3)
				for _, f := range dadosFII {
					if len(palavrasB3) >= 2 && f.Nome != "" {
						palavrasCVM := strings.Fields(strings.ToUpper(f.Nome))
						todasBatem := true
						for _, pB3 := range palavrasB3 {
							bateu := false
							for _, pCVM := range palavrasCVM {
								if strings.HasPrefix(pCVM, pB3) {
									bateu = true
									break
								}
							}
							if !bateu {
								todasBatem = false
								break
							}
						}
						if todasBatem {
							a.VPCota = f.VPCota
							a.Dividendos12M = f.Dividendos12M
							a.CNPJ = f.CNPJ
							dadosFII[a.Ticker] = f
							break
						}
					}
				}
			}
		} else {

			// Proventos Oficiais 12M da B3 (Eventos Corporativos em Dinheiro com desconto de IR em JCP - Bazin)
			if provs := w.b3Client.ObterProventos12MAcao(a.Nome, a.Ticker); provs > 0 {
				a.Dividendos12M = provs
			}

			// Atribui LPA e VPA calculados dinamicamente das demonstrações contábeis oficiais da CVM
			if len(fundamentosCVM) > 0 {
				cnpjAlvo := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(a.CNPJ, ".", ""), "/", ""), "-", "")

				// Se não tem CNPJ cadastrado ou é genérico, resolve dinamicamente na base da CVM
				if (len(cnpjAlvo) != 14 || strings.HasPrefix(cnpjAlvo, "B3")) && cadastroCVM != nil {
					cnpjAlvo = cadastroCVM.ResolverCNPJ(a.Ticker, a.Nome)
					if cnpjAlvo != "" {
						a.CNPJ = cnpjAlvo
					}
				}

				if f, encontrado := fundamentosCVM[cnpjAlvo]; encontrado {
					if f.LPA > 0 {
						a.LPA = f.LPA
					}
					if f.VPA > 0 {
						a.VPA = f.VPA
					}
					a.CrescimentoLucro5A = 10.0 // Média de 10% a.a. para PEG Ratio (Lynch)
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
