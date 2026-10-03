package main

import (
	"context"
	"fmt"
	"sort"
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
	fmt.Println("==================================================================")
	fmt.Println("⚡ STATUSINVEST CONSOLIDATED INGESTION & AUDITORIA MULTI-MODELO")
	fmt.Println("==================================================================")

	// 1. Motores de Valuation
	motores := []domain.Analisador{
		bazin.Novo(0.06),      // 1. Décio Bazin: Preço Teto para DY >= 6%
		graham.Novo(),         // 2. Benjamin Graham: Valor Intrínseco sqrt(22.5 * LPA * VPA)
		lynch.Novo(),          // 3. Peter Lynch: PEG Ratio e Preço Justo
		gordon.Novo(0.11),     // 4. Gordon DDM: Preço Teto de Crescimento Sustentável
		greenblatt.Novo(),     // 5. Joel Greenblatt: The Magic Formula
		piotroski.Novo(),      // 6. Joseph Piotroski: F-Score (0 a 9)
		fii.Novo(0.065),       // 7. FIIs: Segmentação, Cap Rate e Spread NTN-B
		score.Novo(),          // 8. Score Fundamentalista (0 a 100)
	}

	// 2. Snapshot Store em memória (preparado para Redis)
	store := storage.NovoInMemoryStore()

	// 3. Ingestion Service
	service := statusinvest.NovoIngestionService(nil, store, motores)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ativos, err := service.ExecutarSincronizacao(ctx)
	if err != nil {
		fmt.Printf("❌ Erro durante a sincronização: %v\n", err)
		return
	}

	// Separa Ações e FIIs
	var acoes, fiis []*domain.Ativo
	for _, a := range ativos {
		if a.Classe == domain.ClasseFII {
			fiis = append(fiis, a)
		} else {
			acoes = append(acoes, a)
		}
	}

	ordenarScore := func(lista []*domain.Ativo) {
		sort.Slice(lista, func(i, j int) bool {
			if lista[i].Score == lista[j].Score {
				return lista[i].VolumeTotal > lista[j].VolumeTotal
			}
			return lista[i].Score > lista[j].Score
		})
	}

	ordenarScore(acoes)
	ordenarScore(fiis)

	fmt.Println("\n🏆 TOP 5 AÇÕES POR SCORE FUNDAMENTALISTA:")
	fmt.Printf("%-8s | %-28s | %-9s | %-7s | %-6s | %-6s | %-8s | %-12s\n",
		"TICKER", "EMPRESA", "PREÇO", "SCORE", "P/L", "P/VP", "DY", "SEMÁFORO")
	fmt.Println("---------------------------------------------------------------------------------------------------------")
	for i := 0; i < 5 && i < len(acoes); i++ {
		a := acoes[i]
		fmt.Printf("%-8s | %-28s | R$ %6.2f | %4.1f/100 | %5.1f | %5.2f | %5.2f%% | %s\n",
			a.Ticker, truncar(a.Nome, 28), a.PrecoAtual, a.Score, a.PL, a.PVP, a.DY, a.Status)
	}

	fmt.Println("\n🏢 TOP 5 FIIs POR SCORE FUNDAMENTALISTA:")
	fmt.Printf("%-8s | %-28s | %-9s | %-7s | %-6s | %-8s | %-12s | %-12s\n",
		"TICKER", "FUNDO", "PREÇO", "SCORE", "P/VP", "DY", "SEGMENTO", "SEMÁFORO")
	fmt.Println("---------------------------------------------------------------------------------------------------------")
	for i := 0; i < 5 && i < len(fiis); i++ {
		f := fiis[i]
		fmt.Printf("%-8s | %-28s | R$ %6.2f | %4.1f/100 | %5.2f | %5.2f%% | %-12s | %s\n",
			f.Ticker, truncar(f.Nome, 28), f.PrecoAtual, f.Score, f.PVP, f.DY, truncar(f.Segmento, 12), f.Status)
	}

	// Exibe detalhes completos de um ativo consolidado (ex: BBAS3, CMIG4 ou PETR4)
	destaques := []string{"BBAS3", "CMIG4", "PETR4", "HGLG11"}
	fmt.Println("\n🔍 RAIO-X DE ATIVOS CONSOLIDADOS DO STATUSINVEST:")
	for _, ticker := range destaques {
		for _, a := range ativos {
			if a.Ticker == ticker {
				fmt.Printf("\n--- [%s] %s (%s) ---\n", a.Ticker, a.Nome, a.Classe)
				fmt.Printf("  Cotação: R$ %.2f | DY: %.2f%% | P/L: %.2f | P/VP: %.2f | LPA: R$ %.2f | VPA: R$ %.2f\n",
					a.PrecoAtual, a.DY, a.PL, a.PVP, a.LPA, a.VPA)
				if a.Classe == domain.ClasseAcao {
					fmt.Printf("  ROE: %.2f%% | ROIC: %.2f%% | ROA: %.2f%% | Margem Líquida: %.2f%% | Dív.Líq/PL: %.2f\n",
						a.ROE, a.ROIC, a.ROA, a.MargemLiquida, a.DividaLiquidaPL)
					fmt.Printf("  Setor: %s | Subsetor: %s | Segmento: %s\n",
						a.Setor, a.Subsetor, a.Segmento)
				} else {
					fmt.Printf("  VP/Cota: R$ %.2f | Segmento: %s | Gestão: %s | Cotistas: %.0f\n",
						a.VPCota, a.Segmento, a.Gestao, a.NumeroCotistas)
				}
				fmt.Printf("  Valuation: Teto Bazin: R$ %.2f | Graham: R$ %.2f | Lynch: R$ %.2f | Score: %.1f/100 (%s)\n",
					a.PrecoTetoBazin, a.ValorGraham, a.PrecoJustoLynch, a.Score, a.Status)
				break
			}
		}
	}
	fmt.Println("\n🏁 Concluído com sucesso!")
}

func truncar(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
