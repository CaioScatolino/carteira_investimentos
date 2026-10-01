package main

import (
	"context"
	"fmt"
	"time"

	"carteira_investimentos/server/internal/bazin"
	"carteira_investimentos/server/internal/catalog"
	"carteira_investimentos/server/internal/config"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/fii"
	"carteira_investimentos/server/internal/graham"
	"carteira_investimentos/server/internal/lynch"
	"carteira_investimentos/server/internal/provider"
	"carteira_investimentos/server/internal/storage"
	"carteira_investimentos/server/internal/worker"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("🚀 B3 CORE: AUDITORIA GERAL DE MERCADO (CVM + YAHOO + MYSQL)")
	fmt.Println("==================================================================")

	// 1. Carrega configurações do .env e conecta ao MySQL
	cfg := config.Carregar()
	db, err := storage.NovoMySQLConnection(cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if err != nil {
		fmt.Printf("❌ Falha na conexão com o MySQL: %s\n", err)
		return
	}
	defer db.Close()

	// 2. Confirma o catálogo de ativos no MySQL
	repo := catalog.NovoMySQLRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ativosDB, err := repo.ListarTodos(ctx)
	if err != nil {
		fmt.Printf("❌ Falha ao consultar catálogo: %s\n", err)
		return
	}
	fmt.Printf("🏛️ Catálogo MySQL Ativo: %d ativos cadastrados na base local.\n", len(ativosDB))

	// 3. Inicializa os 4 Motores Analíticos
	motores := []domain.Analisador{
		bazin.Novo(0.06), // Décio Bazin: Dividend Yield sustentável >= 6%
		graham.Novo(),    // Benjamin Graham: VI = sqrt(22.5 * LPA * VPA)
		lynch.Novo(),     // Peter Lynch: PEG Ratio e Preço Justo
		fii.Novo(0.065),  // FIIs: P/VP e Spread vs NTN-B (IPCA+ 6.5% a.a.)
	}

	// 4. Inicializa o Cache e o Provedor de Cotações
	store := storage.NovoInMemoryStore()
	provedor := provider.NovoYahooFinanceProvider()

	// 5. Dispara a Sincronização Dinâmica (Yahoo + CVM ITR/FII sem dados fixos)
	sincronizador := worker.NovoMarketSyncWorker(provedor, store, motores, 15*time.Minute)
	sincronizador.ExecutarSincronizacao()

	// 6. Lê os ativos auditados do Cache
	todosAtivos := store.ListarTodos()

	var compras, manter, alerta []*domain.Ativo
	for _, a := range todosAtivos {
		switch a.Status {
		case domain.StatusComprarMais:
			compras = append(compras, a)
		case domain.StatusManter:
			manter = append(manter, a)
		default:
			alerta = append(alerta, a)
		}
	}

	// 7. Painel Consolidado de Decisão (Semáforo Completo)
	fmt.Printf("\n🟢 [COMPRAR MAIS] - %d Ativos Qualificados com Margem de Segurança:\n", len(compras))
	fmt.Println("------------------------------------------------------------------")
	for _, a := range compras {
		imprimirLinhaAtivo("✅", a)
	}

	fmt.Printf("\n🟡 [MANTER] - %d Ativos Neutros (Próximos ao Preço Justo):\n", len(manter))
	fmt.Println("------------------------------------------------------------------")
	for _, a := range manter {
		imprimirLinhaAtivo("⚠️", a)
	}

	fmt.Printf("\n🔴 [ALERTA / EVITAR] - %d Ativos Sem Margem de Segurança:\n", len(alerta))
	fmt.Println("------------------------------------------------------------------")
	for _, a := range alerta {
		imprimirLinhaAtivo("🚨", a)
	}

	fmt.Println("\n==================================================================")
	fmt.Printf("📊 Auditoria Finalizada: %d Ativos Analisados com Sucesso!\n", len(todosAtivos))
	fmt.Println("==================================================================")
}

func imprimirLinhaAtivo(icone string, a *domain.Ativo) {
	fmt.Printf("%s %-6s (%-4s) | Cotação: R$ %6.2f | Proventos 12M: R$ %5.2f\n",
		icone, a.Ticker, a.Classe, a.PrecoAtual, a.Dividendos12M)
	if a.Classe == domain.ClasseAcao {
		fmt.Printf("   • Teto Bazin: R$ %6.2f | VI Graham: R$ %6.2f | Lynch: R$ %6.2f\n",
			a.PrecoTetoBazin, a.ValorGraham, a.PrecoJustoLynch)
		fmt.Printf("   • LPA (CVM): R$ %6.2f | VPA (CVM): R$ %6.2f\n",
			a.LPA, a.VPA)
	} else {
		fmt.Printf("   • Teto Bazin: R$ %6.2f | P/VP: %4.2f (VP Cota: R$ %6.2f) | Spread NTN-B: %+.2f%%\n",
			a.PrecoTetoBazin, a.PVP, a.VPCota, a.SpreadNTNB)
	}
}
