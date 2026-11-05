package b3import

import (
	"bytes"
	"os"
	"testing"
	"time"

	"carteira_investimentos/server/internal/portfolio"
	"github.com/xuri/excelize/v2"
)

func TestParserB3_AnalisarPlanilha(t *testing.T) {
	f := excelize.NewFile()
	sheet := "Movimentação"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Entrada/Saída", "Data", "Movimentação", "Produto", "Instituição", "Quantidade", "Preço unitário", "Valor da Operação",
	}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	dados := [][]interface{}{
		{"Debito", "20/10/2026", "Transferência - Liquidação", "ITSA4 - ITAUSA S/A", "NU INVEST", "10", "10.00", "100.00"},
		{"Credito", "15/05/2026", "Bonificação em Ativos", "ITSA4 - ITAUSA S/A", "NU INVEST", "2", "-", "-"},
		{"Debito", "10/04/2026", "Direitos de Subscrição - Exercido", "BTLG12 - BTG PACTUAL", "NU INVEST", "5", "-", "500.00"},
		{"Debito", "01/04/2026", "Direitos de Subscrição - Não Exercido", "XPML12 - XP MALLS", "NU INVEST", "0", "-", "-"},
		{"Credito", "15/03/2026", "Rendimento", "BTLG11 - BTG PACTUAL", "NU INVEST", "10", "0.80", "8.00"},
		{"Credito", "01/02/2026", "Transferência - Liquidação", "BTLG11 - BTG PACTUAL", "NU INVEST", "10", "100.00", "1000.00"},
		{"Credito", "10/01/2026", "Transferência - Liquidação", "ITSA4 - ITAUSA S/A", "NU INVEST", "20", "8.00", "160.00"},
	}

	for rowIdx, row := range dados {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("erro ao gravar planilha em memória: %v", err)
	}

	parser := NovoParserB3()
	relatorio, err := parser.AnalisarPlanilha(&buf)
	if err != nil {
		t.Fatalf("erro ao analisar planilha: %v", err)
	}

	if relatorio.TotalMovimentacoes != 7 {
		t.Errorf("esperava 7 movimentações, obteve %d", relatorio.TotalMovimentacoes)
	}

	if relatorio.TotalProventos != 8.00 {
		t.Errorf("esperava R$ 8.00 de proventos, obteve R$ %.2f", relatorio.TotalProventos)
	}

	var posBTLG *PosicaoImportada
	var posITSA *PosicaoImportada

	for i := range relatorio.Posicoes {
		p := &relatorio.Posicoes[i]
		if p.Ticker == "BTLG11" {
			posBTLG = p
		}
		if p.Ticker == "ITSA4" {
			posITSA = p
		}
	}

	if posBTLG == nil {
		t.Fatalf("BTLG11 não foi encontrado nas posições ativas")
	}
	if posBTLG.Quantidade != 15 {
		t.Errorf("BTLG11: esperava 15 cotas, obteve %.2f", posBTLG.Quantidade)
	}
	if posBTLG.PrecoMedio != 100.00 {
		t.Errorf("BTLG11: esperava PM de R$ 100.00, obteve R$ %.2f", posBTLG.PrecoMedio)
	}
	if posBTLG.Classe != portfolio.ClasseFII {
		t.Errorf("BTLG11: esperava classe FII, obteve %s", posBTLG.Classe)
	}

	if posITSA == nil {
		t.Fatalf("ITSA4 não foi encontrado nas posições ativas")
	}
	if posITSA.Quantidade != 12 {
		t.Errorf("ITSA4: esperava 12 cotas, obteve %.2f", posITSA.Quantidade)
	}
	if posITSA.PrecoMedio < 7.20 || posITSA.PrecoMedio > 7.30 {
		t.Errorf("ITSA4: esperava PM em torno de R$ 7.27, obteve R$ %.2f", posITSA.PrecoMedio)
	}
	if posITSA.Classe != portfolio.ClasseAcaoBR {
		t.Errorf("ITSA4: esperava classe ACAO_BR, obteve %s", posITSA.Classe)
	}

	t.Logf("✓ Teste aprovado! BTLG11 Qtd: %.0f PM: R$ %.2f | ITSA4 Qtd: %.0f PM: R$ %.2f",
		posBTLG.Quantidade, posBTLG.PrecoMedio, posITSA.Quantidade, posITSA.PrecoMedio)
}

func TestParserB3_ParseNumeroEData(t *testing.T) {
	parser := NovoParserB3()

	if v := parser.parseNumero("1.250,50"); v != 1250.50 {
		t.Errorf("esperava 1250.50, obteve %.2f", v)
	}
	if v := parser.parseNumero("14.72"); v != 14.72 {
		t.Errorf("esperava 14.72, obteve %.2f", v)
	}
	if v := parser.parseNumero("-"); v != 0 {
		t.Errorf("esperava 0 para '-', obteve %.2f", v)
	}

	dt := parser.parseData("28/10/2026")
	if dt.Year() != 2026 || dt.Month() != time.October || dt.Day() != 28 {
		t.Errorf("falha ao interpretar data '28/10/2026': %v", dt)
	}
}

func TestParserB3_ArquivoRealLocalSeExistir(t *testing.T) {
	caminhoReal := `c:\Users\caio\Downloads\movimentacao-2026-11-05-18-02-30.xlsx`
	if _, err := os.Stat(caminhoReal); os.IsNotExist(err) {
		t.Skip("Arquivo real não localizado em Downloads, pulando teste local.")
		return
	}

	f, err := os.Open(caminhoReal)
	if err != nil {
		t.Skipf("Não foi possível abrir arquivo: %v", err)
		return
	}
	defer f.Close()

	parser := NovoParserB3()
	rel, err := parser.AnalisarPlanilha(f)
	if err != nil {
		t.Fatalf("Erro ao analisar arquivo real da B3: %v", err)
	}

	t.Logf("✓ Arquivo real da B3 processado com sucesso!")
	t.Logf("• Total Movimentações: %d", rel.TotalMovimentacoes)
	t.Logf("• Período: %s a %s", rel.DataInicio.Format("02/01/2006"), rel.DataFim.Format("02/01/2006"))
	t.Logf("• Total Ativos Ativos: %d", rel.TotalAtivosAtivos)
	t.Logf("• Custo Total Investido: R$ %.2f", rel.TotalInvestido)
	t.Logf("• Proventos Totais: R$ %.2f", rel.TotalProventos)
	t.Logf("• Total Operações Reconstruídas: %d", len(rel.Operacoes))
	t.Logf("• Total Proventos Registros: %d", len(rel.Proventos))
}
