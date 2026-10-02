package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"carteira_investimentos/server/internal/bazin"
	"carteira_investimentos/server/internal/catalog"
	"carteira_investimentos/server/internal/config"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/fii"
	"carteira_investimentos/server/internal/gordon"
	"carteira_investimentos/server/internal/graham"
	"carteira_investimentos/server/internal/lynch"
	"carteira_investimentos/server/internal/provider"
	"carteira_investimentos/server/internal/score"
	"carteira_investimentos/server/internal/storage"
	"carteira_investimentos/server/internal/worker"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("🚀 B3 CORE: AUDITORIA MULTI-MODELO DE MERCADO (CVM + B3 OFICIAL)")
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

	// 3. Inicializa os Motores de Valuation e Consolidação
	motores := []domain.Analisador{
		bazin.Novo(0.06),  // 1. Décio Bazin: Dividend Yield sustentável >= 6%
		graham.Novo(),     // 2. Benjamin Graham: Valor Intrínseco
		lynch.Novo(),      // 3. Peter Lynch: PEG Ratio e Preço Justo de Crescimento
		gordon.Novo(0.11), // 4. Gordon DDM: Preço Teto com crescimento sustentável (taxa desc. 11%)
		fii.Novo(0.065),   // 5. FIIs: P/VP e Spread real vs NTN-B (IPCA+ 6.5% a.a.)
		score.Novo(),      // 6. Score Fundamentalista (0 a 100) e Pareceres Individuais
	}

	// 4. Inicializa o Cache e o Provedor Oficial B3 (COTAHIST em Lote sem Yahoo)
	store := storage.NovoInMemoryStore()
	provedor := provider.NovoB3MarketProvider(nil, repo, 500_000.00)

	// 5. Scanner Geral de Mercado: Passando nil/vazio, o B3MarketProvider varre TODOS os ativos líquidos da B3
	var universo []string // Slice vazio ativa o modo Scanner de Mercado Aberto

	sincronizador := worker.NovoMarketSyncWorker(provedor, store, motores, 15*time.Minute)
	sincronizador.ExecutarSincronizacao(universo)

	// 6. Lê todos os ativos auditados e separa em 3 listas independentes
	todosAtivos := store.ListarTodos()

	var acoes, fiis, etfs []*domain.Ativo
	for _, a := range todosAtivos {
		switch a.Classe {
		case domain.ClasseFII:
			fiis = append(fiis, a)
		case domain.ClasseETF:
			etfs = append(etfs, a)
		default:
			acoes = append(acoes, a)
		}
	}

	// Ordenador por Score decrescente (e desempate por Liquidez / Volume)
	ordenarPorScore := func(slice []*domain.Ativo) {
		sort.Slice(slice, func(i, j int) bool {
			if slice[i].Score == slice[j].Score {
				return slice[i].VolumeTotal > slice[j].VolumeTotal
			}
			return slice[i].Score > slice[j].Score
		})
	}

	ordenarPorScore(fiis)
	ordenarPorScore(acoes)
	ordenarPorScore(etfs)

	// ==================================================================
	// 7. RANKING COMPLETO DE FUNDOS IMOBILIÁRIOS (FIIs)
	// ==================================================================
	fmt.Println("\n==================================================================")
	fmt.Printf("🏢 1. RANKING COMPLETO DE FUNDOS IMOBILIÁRIOS (%d FIIs) - ORDENADOS POR SCORE\n", len(fiis))
	fmt.Println("==================================================================")
	for i, f := range fiis {
		imprimirFII(i+1, f)
	}

	// ==================================================================
	// 8. RANKING COMPLETO DE AÇÕES
	// ==================================================================
	fmt.Println("\n==================================================================")
	fmt.Printf("📈 2. RANKING COMPLETO DE AÇÕES (%d Ações) - ORDENADOS POR SCORE\n", len(acoes))
	fmt.Println("==================================================================")
	for i, a := range acoes {
		imprimirAcao(i+1, a)
	}

	// ==================================================================
	// 9. RANKING COMPLETO DE ETFs
	// ==================================================================
	if len(etfs) > 0 {
		fmt.Println("\n==================================================================")
		fmt.Printf("🌐 3. RANKING COMPLETO DE ETFs (%d ETFs) - ORDENADOS POR LIQUIDEZ\n", len(etfs))
		fmt.Println("==================================================================")
		for i, e := range etfs {
			imprimirETF(i+1, e)
		}
	}

	fmt.Println("\n==================================================================")
	fmt.Printf("📊 Auditoria Finalizada: %d Ativos Analisados (Ações: %d | FIIs: %d | ETFs: %d)\n",
		len(todosAtivos), len(acoes), len(fiis), len(etfs))
	fmt.Println("==================================================================")
}

func getIconeStatus(status domain.StatusRecomendacao) string {
	switch status {
	case domain.StatusComprarMais:
		return "🟢 [COMPRAR MAIS]"
	case domain.StatusManter:
		return "🟡 [MANTER]"
	default:
		return "🔴 [ALERTA / EVITAR]"
	}
}

func getIconeParecer(status domain.StatusParecer) string {
	switch status {
	case domain.ParecerAprovado:
		return "✅"
	case domain.ParecerAtencao:
		return "⚠️"
	default:
		return "❌"
	}
}

func imprimirAcao(posicao int, a *domain.Ativo) {
	fmt.Printf("\n%3dº %-6s (%-12s) | Cotação: R$ %6.2f | SCORE: %3.0f/100 %s\n",
		posicao, a.Ticker, a.Nome, a.PrecoAtual, a.Score, getIconeStatus(a.Status))
	fmt.Printf("     • Indicadores: P/L: %4.1fx | P/VP: %4.2fx | ROE: %4.1f%% | Yield: %4.1f%% (R$ %.2f)\n",
		a.PL, a.PVPReal, a.ROE, a.DY, a.Dividendos12M)

	if p, ok := a.Pareceres["Bazin"]; ok {
		fmt.Printf("     ├── 💰 Bazin:      %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := a.Pareceres["Graham"]; ok {
		fmt.Printf("     ├── 📐 Graham:     %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := a.Pareceres["Lynch"]; ok {
		fmt.Printf("     ├── 🧠 Lynch:      %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := a.Pareceres["Gordon"]; ok {
		fmt.Printf("     ├── 📈 Gordon:     %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := a.Pareceres["Greenblatt"]; ok {
		fmt.Printf("     └── ⚡ Greenblatt: %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
}

func imprimirFII(posicao int, f *domain.Ativo) {
	fmt.Printf("\n%3dº %-6s (%-12s) | Cotação: R$ %6.2f | SCORE: %3.0f/100 %s\n",
		posicao, f.Ticker, f.Nome, f.PrecoAtual, f.Score, getIconeStatus(f.Status))
	fmt.Printf("     • Indicadores: P/VP: %4.2f (VP: R$ %.2f) | Proventos 12M: R$ %5.2f (DY: %4.1f%%)\n",
		f.PVP, f.VPCota, f.Dividendos12M, f.DY)

	if p, ok := f.Pareceres["PVP"]; ok {
		fmt.Printf("     ├── 🏢 CVM P/VP:       %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := f.Pareceres["SpreadNTNB"]; ok {
		fmt.Printf("     ├── 📈 Prêmio NTN-B:   %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := f.Pareceres["Bazin"]; ok {
		fmt.Printf("     └── 💰 Teto Bazin:     %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
}

func imprimirETF(posicao int, e *domain.Ativo) {
	fmt.Printf("\n%3dº %-6s (%-12s) | Cotação: R$ %6.2f | SCORE: %3.0f/100 %s\n",
		posicao, e.Ticker, e.Nome, e.PrecoAtual, e.Score, getIconeStatus(e.Status))
	if p, ok := e.Pareceres["Liquidez"]; ok {
		fmt.Printf("     └── 💧 Liquidez B3:    %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
}
