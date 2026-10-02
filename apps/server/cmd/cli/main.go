package main

import (
	"flag"
	"fmt"
	"sort"
	"time"

	"carteira_investimentos/server/internal/b3"
)

func main() {
	syncB3 := flag.Bool("sync-b3", false, "Dispara a sincronização das cotações diárias oficiais da B3")
	flag.Parse()

	fmt.Println("==================================================================")
	fmt.Println("⚡ B3 CORE CLI: FERRAMENTAS ADMINISTRATIVAS & SCANNER EM LOTE")
	fmt.Println("==================================================================")

	if !*syncB3 {
		fmt.Println("Comando não especificado. Para sincronizar as cotações oficiais da B3, execute:")
		fmt.Println("👉 go run cmd/cli/main.go --sync-b3")
		return
	}

	fmt.Println("\n🔄 Conectando aos servidores da B3 para localizar o pregão mais recente...")
	inicio := time.Now()

	client := b3.NovoB3Client()
	zipBytes, dataPregao, err := client.ObterArquivoMaisRecente()
	if err != nil {
		fmt.Printf("❌ Erro ao baixar dados da B3: %s\n", err)
		return
	}

	fmt.Printf("📦 Arquivo oficial do pregão de %s baixado (%.2f KB) em %s!\n",
		dataPregao, float64(len(zipBytes))/1024, time.Since(inicio))

	// Executa o parser posicional em memória
	inicioParser := time.Now()
	cotacoes, err := b3.ParseCOTAHIST(zipBytes)
	if err != nil {
		fmt.Printf("❌ Erro ao processar arquivo: %s\n", err)
		return
	}

	fmt.Printf("⚡ Processamento concluído em %s! %d ativos negociados encontrados.\n",
		time.Since(inicioParser), len(cotacoes))

	// Ordena os 10 ativos com maior volume financeiro no dia
	type itemVolume struct {
		Ticker string
		Cot    *b3.CotacaoDiaria
	}
	var ranking []itemVolume
	volumeTotalGeral := 0.0

	for t, c := range cotacoes {
		ranking = append(ranking, itemVolume{Ticker: t, Cot: c})
		volumeTotalGeral += c.VolumeTotal
	}

	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].Cot.VolumeTotal > ranking[j].Cot.VolumeTotal
	})

	fmt.Println("\n🏆 TOP 10 MAIORES VOLUMES NEGOCIADOS NA B3 NO PREGÃO:")
	fmt.Println("------------------------------------------------------------------")
	for i := 0; i < 10 && i < len(ranking); i++ {
		item := ranking[i].Cot
		fmt.Printf("%2d. %-6s (%-12s) | Fechamento: R$ %6.2f | Volume: R$ %13.2f\n",
			i+1, item.Ticker, item.NomeResumido, item.PrecoFechamento, item.VolumeTotal)
	}

	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("📊 Volume Financeiro Total do Pregão: R$ %.2f Bilhões\n", volumeTotalGeral/1e9)
	fmt.Printf("⏱️ Tempo Total da Operação: %s (Zero requisições externas ao Yahoo!)\n", time.Since(inicio))
	fmt.Println("==================================================================")
}
