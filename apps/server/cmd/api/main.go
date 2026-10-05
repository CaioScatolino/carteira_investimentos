package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"carteira_investimentos/server/internal/api"
	"carteira_investimentos/server/internal/b3"
	"carteira_investimentos/server/internal/bazin"
	"carteira_investimentos/server/internal/catalog"
	"carteira_investimentos/server/internal/config"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/fii"
	"carteira_investimentos/server/internal/gordon"
	"carteira_investimentos/server/internal/graham"
	"carteira_investimentos/server/internal/greenblatt"
	"carteira_investimentos/server/internal/lynch"
	"carteira_investimentos/server/internal/macro"
	"carteira_investimentos/server/internal/piotroski"
	"carteira_investimentos/server/internal/score"
	"carteira_investimentos/server/internal/statusinvest"
	"carteira_investimentos/server/internal/storage"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("🚀 B3 CORE: AUDITORIA MULTI-MODELO DE MERCADO (CVM + B3 OFICIAL)")
	fmt.Println("==================================================================")

	// 1. Sincroniza parâmetros macroeconômicos oficiais (BCB SGS / Tesouro Direto)
	macro.SincronizarTaxasOficiais()
	cenario := macro.ObterCenario()
	fmt.Printf("🏛️ CENÁRIO MACRO ATIVO (%s):\n", cenario.DataAtualizacao)
	fmt.Printf("   • Taxa Selic Meta: %.2f%% a.a. | NTN-B (IPCA+): %.2f%% a.a.\n", cenario.TaxaSelic, cenario.TaxaNTNB)
	fmt.Printf("   • Hurdle Mínimo Ações: %.2f%% a.a. | Hurdle FIIs: %.2f%% a.a.\n",
		cenario.TaxaNTNB+cenario.SpreadMinimoAcoes, cenario.TaxaNTNB+cenario.SpreadMinimoFIITijolo)
	fmt.Println("==================================================================")

	// 2. Carrega configurações do .env e conecta ao MySQL (com fallback gracioso)
	cfg := config.Carregar()
	db, err := storage.NovoMySQLConnection(cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if err != nil {
		fmt.Printf("⚠️ MySQL local offline: operando em modo Resiliente de Alta Disponibilidade (In-Memory + StatusInvest + B3)\n")
	} else {
		defer db.Close()
		// Confirma catálogo de ativos no MySQL se conectado
		repo := catalog.NovoMySQLRepository(db)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		ativosDB, err := repo.ListarTodos(ctx)
		cancel()
		if err == nil {
			fmt.Printf("🏛️ Catálogo MySQL Ativo: %d ativos cadastrados na base local.\n", len(ativosDB))
		}
	}

	// 3. Inicializa os Motores de Valuation e Consolidação Calibrados ao Macro
	motores := []domain.Analisador{
		bazin.NovoDinamico(),  // 1. Décio Bazin: Hurdle dinâmico via NTN-B + Spread
		graham.Novo(),         // 2. Benjamin Graham: Valor Intrínseco
		lynch.Novo(),          // 3. Peter Lynch: PEG Ratio e Preço Justo de Crescimento
		gordon.NovoDinamico(), // 4. Gordon DDM: Custo de capital calibrado via CAPM / Selic
		greenblatt.Novo(),     // 5. Joel Greenblatt: The Magic Formula (EV/EBIT + ROIC com adaptação p/ bancos)
		piotroski.Novo(),      // 6. Joseph Piotroski: F-Score de Solvência & Saúde Contábil (0 a 9)
		fii.Novo(0),           // 7. FIIs: Segmentação (Tijolo/Papel), Cap Rate Implícito e Spread NTN-B dinâmico
		score.Novo(),          // 8. Score Fundamentalista Composto (0 a 100) com Veredito de Risco vs Renda Fixa
	}

	// 4. Inicializa o Cache e o Ingestion Service Consolidado (StatusInvest + B3)
	store := storage.NovoInMemoryStore()
	b3Client := b3.NovoB3Client()
	ingestionService := statusinvest.NovoIngestionService(nil, b3Client, store, motores)

	// 5. Ingestão consolidada em 2 requisições ultra-rápidas (<1s para 100% dos ativos da B3)
	ctxIngest, cancelIngest := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelIngest()

	if _, err := ingestionService.ExecutarSincronizacao(ctxIngest); err != nil {
		fmt.Printf("⚠️ Erro na sincronização inicial do StatusInvest: %v\n", err)
	}

	// Inicia rotina periódica em background a cada 1 hora
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			ctxBg, cancelBg := context.WithTimeout(context.Background(), 60*time.Second)
			_, _ = ingestionService.ExecutarSincronizacao(ctxBg)
			cancelBg()
		}
	}()

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

	// ==================================================================
	// 10. INICIALIZAÇÃO DA API REST (HTTP SERVER)
	// ==================================================================
	apiHandler := api.NovoHandler(store, b3Client)
	router := apiHandler.ConfigurarRotas()

	servidor := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	fmt.Println("\n==================================================================")
	fmt.Println("🌐 SERVIDOR HTTP ATIVO EM http://localhost:8080")
	fmt.Println("   • Health Check: GET http://localhost:8080/health")
	fmt.Println("   • Rankings:     GET http://localhost:8080/api/v1/rankings")
	fmt.Println("   • Ativo Único:  GET http://localhost:8080/api/v1/ativos/PETR4")
	fmt.Println("==================================================================")

	if err := servidor.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Erro fatal no servidor HTTP: %s\n", err)
	}

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

	if a.PrecoTetoConsolidado > 0 {
		fmt.Printf("     ├── 🎯 Preço Teto: R$ %6.2f | Prêmio: %+.1f%% (Consenso)\n", a.PrecoTetoConsolidado, a.PremioDescontoPercentual)
	}
	if p, ok := a.Pareceres["VereditoRisco"]; ok {
		fmt.Printf("     ├── ⚖️ Risco NTN-B:  %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
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

	if f.PrecoTetoConsolidado > 0 {
		fmt.Printf("     ├── 🎯 Preço Teto: R$ %6.2f | Prêmio: %+.1f%% (Consenso)\n", f.PrecoTetoConsolidado, f.PremioDescontoPercentual)
	}
	if p, ok := f.Pareceres["VereditoRisco"]; ok {
		fmt.Printf("     ├── ⚖️ Risco NTN-B:    %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := f.Pareceres["PVP"]; ok {
		fmt.Printf("     ├── 🏢 CVM P/VP:       %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
	if p, ok := f.Pareceres["Bazin"]; ok {
		fmt.Printf("     └── 💰 Teto Renda:     %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
}

func imprimirETF(posicao int, e *domain.Ativo) {
	fmt.Printf("\n%3dº %-6s (%-12s) | Cotação: R$ %6.2f | SCORE: %3.0f/100 %s\n",
		posicao, e.Ticker, e.Nome, e.PrecoAtual, e.Score, getIconeStatus(e.Status))
	if p, ok := e.Pareceres["Liquidez"]; ok {
		fmt.Printf("     └── 💧 Liquidez B3:    %s %s (%s)\n", getIconeParecer(p.Status), p.Metrica, p.Detalhe)
	}
}
