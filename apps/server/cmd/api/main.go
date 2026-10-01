package main

import (
	"fmt"
	"time"

	"carteira_investimentos/server/internal/bazin"
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
	fmt.Println("🚀 PAINEL GERAL DE AUDITORIA B3 (SNAPSHOT CONSOLIDADO)")
	fmt.Println("==================================================================")

	store := storage.NovoInMemoryStore()
	provedor := provider.NovoYahooFinanceProvider()

	motores := []domain.Analisador{
		bazin.Novo(0.06),
		graham.Novo(),
		lynch.Novo(),
		fii.Novo(0.065), // Taxa de referência Tesouro IPCA+: 6.5% a.a.
	}

	sincronizador := worker.NovoMarketSyncWorker(provedor, store, motores, 15*time.Minute)
	sincronizador.IniciarLoop()

	todosAtivos := store.ListarTodos()

	// Separação por categorias do Semáforo
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

	// 1. OPORTUNIDADES: COMPRAR MAIS
	fmt.Printf("\n🟢 [COMPRAR MAIS] - %d Ativos Qualificados com Margem de Segurança:\n", len(compras))
	fmt.Println("------------------------------------------------------------------")
	for _, a := range compras {
		imprimirLinhaAtivo("✅", a)
	}

	// 2. MANTER / NEUTROS
	fmt.Printf("\n🟡 [MANTER] - %d Ativos Neutros (Preço Próximo ao Justo):\n", len(manter))
	fmt.Println("------------------------------------------------------------------")
	for _, a := range manter {
		imprimirLinhaAtivo("⚠️", a)
	}

	// 3. ALERTA / AGUARDAR
	fmt.Printf("\n🔴 [ALERTA] - %d Ativos Caros / Sem Margem de Segurança:\n", len(alerta))
	fmt.Println("------------------------------------------------------------------")
	for _, a := range alerta {
		imprimirLinhaAtivo("🚨", a)
	}

	fmt.Println("==================================================================")
	fmt.Printf("📊 Total de Ativos Auditados: %d | Base de Dados 100%% Atualizada\n", len(todosAtivos))
	fmt.Println("==================================================================")
}

func imprimirLinhaAtivo(icone string, a *domain.Ativo) {
	fmt.Printf("%s %-6s (%-4s) | Cotação: R$ %6.2f | Proventos 12M: R$ %5.2f\n",
		icone, a.Ticker, a.Classe, a.PrecoAtual, a.Dividendos12M)
	fmt.Printf("   • Teto Bazin: R$ %6.2f", a.PrecoTetoBazin)
	if a.Classe == domain.ClasseAcao {
		fmt.Printf(" | VI Graham: R$ %6.2f | Lynch: R$ %6.2f\n", a.ValorGraham, a.PrecoJustoLynch)
	} else {
		fmt.Printf(" | P/VP: %4.2f | Spread NTN-B: %+.2f%%\n", a.PVP, a.SpreadNTNB)
	}
}
