package main

import (
	"fmt"

	"carteira_investimentos/server/internal/bazin"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/graham"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("📊 AUDITORIA B3: SEMÁFORO INTELIGENTE (BAZIN + GRAHAM)")
	fmt.Println("================================================================")

	carteira := []*domain.Ativo{
		{
			Ticker:        "PETR4",
			Classe:        domain.ClasseAcao,
			PrecoAtual:    38.50,
			Dividendos12M: 5.20,
			LPA:           8.10,
			VPA:           31.40,
		},
		{
			Ticker:        "VALE3",
			Classe:        domain.ClasseAcao,
			PrecoAtual:    62.00,
			Dividendos12M: 4.80,
			LPA:           7.20,
			VPA:           42.50,
		},
		{
			Ticker:        "MXRF11",
			Classe:        domain.ClasseFII,
			PrecoAtual:    10.15,
			Dividendos12M: 1.10,
		},
		{
			Ticker:        "CARO3", // Exemplo de ativo esticado/caro
			Classe:        domain.ClasseAcao,
			PrecoAtual:    95.00,
			Dividendos12M: 1.50, // Yield muito baixo (~1.5%)
			LPA:           2.00,
			VPA:           15.00,
		},
	}

	yieldBazin := 0.06 // 6% ao ano

	for _, ativo := range carteira {
		// 1. Motor Décio Bazin
		tetoBazin, errBazin := bazin.Calcular(ativo.Dividendos12M, yieldBazin)
		if errBazin == nil {
			ativo.PrecoTetoBazin = tetoBazin
		}

		// 2. Motor Benjamin Graham (apenas para ações)
		if ativo.Classe == domain.ClasseAcao {
			viGraham, errGraham := graham.Calcular(ativo.LPA, ativo.VPA)
			if errGraham == nil {
				ativo.ValorGraham = viGraham
			}
		}

		// 3. Execução do Semáforo Decisório
		ativo.AvaliarSemaforo()

		// 4. Formatação Visual do Semáforo
		var icone string
		switch ativo.Status {
		case domain.StatusComprarMais:
			icone = "✅"
		case domain.StatusManter:
			icone = "⚠️"
		default:
			icone = "🚨"
		}

		fmt.Printf("[%s - %s] Cotação Atual: R$ %.2f\n", ativo.Ticker, ativo.Classe, ativo.PrecoAtual)
		fmt.Printf("   Teto Bazin (6%%): R$ %.2f", ativo.PrecoTetoBazin)
		if ativo.ValorGraham > 0 {
			fmt.Printf(" | VI Graham: R$ %.2f", ativo.ValorGraham)
		}
		fmt.Println()
		fmt.Printf("   Recomendação:    %s %s\n", icone, ativo.Status)
		fmt.Println("----------------------------------------------------------------")
	}

	fmt.Println("================================================================")
}
