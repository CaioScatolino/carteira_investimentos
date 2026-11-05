package b3import

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"carteira_investimentos/server/internal/portfolio"
	"github.com/xuri/excelize/v2"
)

// ParserB3 lê e decodifica planilhas oficiais de movimentação da B3
type ParserB3 struct{}

func NovoParserB3() *ParserB3 {
	return &ParserB3{}
}

// AnalisarPlanilha processa o arquivo XLSX e gera o relatório completo de custódia e proventos
func (p *ParserB3) AnalisarPlanilha(r io.Reader) (*RelatorioAnaliseB3, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir arquivo Excel da B3: %w", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("a planilha não contém nenhuma aba de dados")
	}

	sheetName := sheetList[0]
	for _, name := range sheetList {
		norm := strings.ToLower(name)
		if strings.Contains(norm, "movimenta") || strings.Contains(norm, "negocia") {
			sheetName = name
			break
		}
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler linhas da aba '%s': %w", sheetName, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("a planilha está vazia ou contém apenas cabeçalhos")
	}

	// 1. Identificar índices das colunas no cabeçalho
	headerRow := rows[0]
	idxES, idxData, idxMov, idxProd, idxInst, idxQtd, idxPreco, idxValor := -1, -1, -1, -1, -1, -1, -1, -1

	for i, col := range headerRow {
		c := strings.ToLower(strings.TrimSpace(col))
		switch {
		case strings.Contains(c, "entrada") || strings.Contains(c, "saída") || strings.Contains(c, "saida"):
			idxES = i
		case c == "data":
			idxData = i
		case strings.Contains(c, "movimenta"):
			idxMov = i
		case strings.Contains(c, "produto"):
			idxProd = i
		case strings.Contains(c, "institui"):
			idxInst = i
		case strings.Contains(c, "quantidade"):
			idxQtd = i
		case strings.Contains(c, "preço") || strings.Contains(c, "preco"):
			idxPreco = i
		case strings.Contains(c, "valor"):
			idxValor = i
		}
	}

	// Fallback para índices posicionais padrão se não encontrados por texto
	if idxES == -1 {
		idxES = 0
	}
	if idxData == -1 {
		idxData = 1
	}
	if idxMov == -1 {
		idxMov = 2
	}
	if idxProd == -1 {
		idxProd = 3
	}
	if idxInst == -1 {
		idxInst = 4
	}
	if idxQtd == -1 {
		idxQtd = 5
	}
	if idxPreco == -1 {
		idxPreco = 6
	}
	if idxValor == -1 {
		idxValor = 7
	}

	// 2. Extrair e sanitizar linhas brutas
	var linhasBrutas []LinhaMovimentacaoB3
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) == 0 {
			continue
		}

		getVal := func(idx int) string {
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		es := getVal(idxES)
		dataStr := getVal(idxData)
		mov := getVal(idxMov)
		prod := getVal(idxProd)
		inst := getVal(idxInst)
		qtdStr := getVal(idxQtd)
		precoStr := getVal(idxPreco)
		valorStr := getVal(idxValor)

		if prod == "" && mov == "" {
			continue
		}

		parsedDate := p.parseData(dataStr)
		if parsedDate.IsZero() {
			continue
		}

		ticker, _ := p.extrairTickerENome(prod)
		qtd := p.parseNumero(qtdStr)
		preco := p.parseNumero(precoStr)
		valor := p.parseNumero(valorStr)

		linhasBrutas = append(linhasBrutas, LinhaMovimentacaoB3{
			LinhaOrigem:   rowIdx + 1,
			EntradaSaida:  es,
			Data:          parsedDate,
			Movimentacao:  mov,
			Produto:       prod,
			Ticker:        ticker,
			Instituicao:   inst,
			Quantidade:    qtd,
			PrecoUnitario: preco,
			ValorOperacao: valor,
		})
	}

	if len(linhasBrutas) == 0 {
		return nil, fmt.Errorf("nenhuma linha válida com data e movimentação foi encontrada na planilha")
	}

	// 3. Ordenar cronologicamente (da mais antiga até a mais recente)
	// Como a planilha da B3 costuma vir do mais recente para o mais antigo, invertemos preservando a ordem intra-dia
	sort.SliceStable(linhasBrutas, func(i, j int) bool {
		if linhasBrutas[i].Data.Equal(linhasBrutas[j].Data) {
			// Na planilha original decrescente, linhas maiores no arquivo ocorreram antes no mesmo dia
			return linhasBrutas[i].LinhaOrigem > linhasBrutas[j].LinhaOrigem
		}
		return linhasBrutas[i].Data.Before(linhasBrutas[j].Data)
	})

	dataInicio := linhasBrutas[0].Data
	dataFim := linhasBrutas[len(linhasBrutas)-1].Data

	// 4. Executar Replay Cronológico da Custódia e Proventos
	type estadoCustodia struct {
		ticker         string
		razaoSocial    string
		classe         portfolio.ClasseAtivo
		quantidade     float64
		custoTotal     float64
		corretora      string
		totalCompras   int
		totalVendas    int
		totalProventos float64
	}

	custodiaMap := make(map[string]*estadoCustodia)
	var operacoesReconstruidas []OperacaoImportada
	var proventosIdentificados []ProventoImportado
	var totalProventosHistorico float64

	getOuCriarCustodia := func(ticker, razaoSocial, corretora string, classe portfolio.ClasseAtivo) *estadoCustodia {
		if ec, ok := custodiaMap[ticker]; ok {
			if ec.razaoSocial == "" && razaoSocial != "" {
				ec.razaoSocial = razaoSocial
			}
			if corretora != "" {
				ec.corretora = corretora
			}
			return ec
		}
		ec := &estadoCustodia{
			ticker:      ticker,
			razaoSocial: razaoSocial,
			classe:      classe,
			corretora:   corretora,
		}
		custodiaMap[ticker] = ec
		return ec
	}

	for _, l := range linhasBrutas {
		ticker, razao := p.extrairTickerENome(l.Produto)
		classe := p.classificarClasse(ticker, l.Produto)
		movNorm := strings.ToLower(l.Movimentacao)

		// -------------------------------------------------------------
		// A. PROVENTOS / RENDIMENTOS / JUROS / DIVIDENDOS
		// -------------------------------------------------------------
		if p.isProvento(movNorm) {
			// Proventos transferidos com débito representam estornos ou compensações
			if strings.EqualFold(l.EntradaSaida, "Debito") {
				continue
			}

			tipoProvento := p.mapearTipoProvento(movNorm)
			valorUnit := l.PrecoUnitario
			if valorUnit <= 0 && l.Quantidade > 0 && l.ValorOperacao > 0 {
				valorUnit = l.ValorOperacao / l.Quantidade
			}

			proventosIdentificados = append(proventosIdentificados, ProventoImportado{
				Ticker:        ticker,
				Tipo:          tipoProvento,
				DataPagamento: l.Data,
				Quantidade:    l.Quantidade,
				ValorUnitario: valorUnit,
				ValorTotal:    l.ValorOperacao,
				Corretora:     l.Instituicao,
			})
			totalProventosHistorico += l.ValorOperacao

			ec := getOuCriarCustodia(ticker, razao, l.Instituicao, classe)
			ec.totalProventos += l.ValorOperacao
			continue
		}

		// -------------------------------------------------------------
		// B. DIREITOS E RECIBOS DE SUBSCRIÇÃO
		// -------------------------------------------------------------
		if p.isSubscricao(ticker, movNorm) {
			// Se o direito foi EXERCIDO, o investidor pagou o valor da subscrição e adquiriu cotas!
			if strings.Contains(movNorm, "exercido") && l.ValorOperacao > 0 && l.Quantidade > 0 {
				tickerBase := p.converterDireitoParaAtivoBase(ticker)
				classeBase := p.classificarClasse(tickerBase, l.Produto)
				ec := getOuCriarCustodia(tickerBase, razao, l.Instituicao, classeBase)

				precoSubscricao := l.ValorOperacao / l.Quantidade
				ec.quantidade += l.Quantidade
				ec.custoTotal += l.ValorOperacao
				ec.totalCompras++

				operacoesReconstruidas = append(operacoesReconstruidas, OperacaoImportada{
					Ticker:        tickerBase,
					Classe:        classeBase,
					Tipo:          portfolio.TransacaoCompra,
					Data:          l.Data,
					Quantidade:    l.Quantidade,
					PrecoUnitario: precoSubscricao,
					ValorTotal:    l.ValorOperacao,
					Corretora:     l.Instituicao,
					Origem:        "B3_EXERCICIO_SUBSCRICAO",
					Notas:         fmt.Sprintf("Exercício de subscrição de %s (Qtd: %.0f cotas)", ticker, l.Quantidade),
				})
			}
			// Direitos emitidos, cessões, solicitações e recibos '13' são transitórios e não afetam o saldo base diretamente
			continue
		}

		// -------------------------------------------------------------
		// C. OPERAÇÕES DE COMPRA / VENDA / LIQUIDAÇÃO
		// -------------------------------------------------------------
		if strings.Contains(movNorm, "transferência - liquidação") ||
			strings.Contains(movNorm, "transferencia - liquidacao") ||
			movNorm == "compra" ||
			movNorm == "venda" {

			ec := getOuCriarCustodia(ticker, razao, l.Instituicao, classe)
			valorOp := l.ValorOperacao
			if valorOp <= 0 && l.Quantidade > 0 && l.PrecoUnitario > 0 {
				valorOp = l.Quantidade * l.PrecoUnitario
			}

			// COMPRA: Entrada de cotas na custódia
			if (strings.EqualFold(l.EntradaSaida, "Credito") && movNorm != "venda") || movNorm == "compra" {
				ec.quantidade += l.Quantidade
				ec.custoTotal += valorOp
				ec.totalCompras++

				operacoesReconstruidas = append(operacoesReconstruidas, OperacaoImportada{
					Ticker:        ticker,
					Classe:        classe,
					Tipo:          portfolio.TransacaoCompra,
					Data:          l.Data,
					Quantidade:    l.Quantidade,
					PrecoUnitario: l.PrecoUnitario,
					ValorTotal:    valorOp,
					Corretora:     l.Instituicao,
					Origem:        "B3_EXTRATO",
					Notas:         fmt.Sprintf("Importado da B3: %s", l.Movimentacao),
				})
			} else if strings.EqualFold(l.EntradaSaida, "Debito") || movNorm == "venda" {
				// VENDA: Saída de cotas da custódia (abate o custo proporcional pelo Preço Médio)
				pmAtual := 0.0
				if ec.quantidade > 0 {
					pmAtual = ec.custoTotal / ec.quantidade
				}

				custoAbatido := l.Quantidade * pmAtual
				ec.quantidade -= l.Quantidade
				ec.custoTotal -= custoAbatido
				ec.totalVendas++

				if ec.quantidade <= 0.0001 {
					ec.quantidade = 0
					ec.custoTotal = 0
				}

				operacoesReconstruidas = append(operacoesReconstruidas, OperacaoImportada{
					Ticker:        ticker,
					Classe:        classe,
					Tipo:          portfolio.TransacaoVenda,
					Data:          l.Data,
					Quantidade:    l.Quantidade,
					PrecoUnitario: l.PrecoUnitario,
					ValorTotal:    valorOp,
					Corretora:     l.Instituicao,
					Origem:        "B3_EXTRATO",
					Notas:         fmt.Sprintf("Importado da B3: %s", l.Movimentacao),
				})
			}
			continue
		}

		// -------------------------------------------------------------
		// D. EVENTOS CORPORATIVOS (Desdobro, Bonificação, Grupamento)
		// -------------------------------------------------------------
		if strings.Contains(movNorm, "bonificação em ativos") || strings.Contains(movNorm, "bonificacao") {
			ec := getOuCriarCustodia(ticker, razao, l.Instituicao, classe)
			ec.quantidade += l.Quantidade
			// Na bonificação de ações, as novas cotas são incorporadas sem desembolso financeiro
			continue
		}

		if strings.Contains(movNorm, "desdobro") {
			ec := getOuCriarCustodia(ticker, razao, l.Instituicao, classe)
			ec.quantidade += l.Quantidade
			continue
		}

		if strings.Contains(movNorm, "grupamento") {
			ec := getOuCriarCustodia(ticker, razao, l.Instituicao, classe)
			ec.quantidade -= l.Quantidade
			if ec.quantidade <= 0.0001 {
				ec.quantidade = 0
				ec.custoTotal = 0
			}
			continue
		}
	}

	// 5. Montar lista de posições ativas (saldo > 0.001 cotas)
	var posicoesResultantes []PosicaoImportada
	var totalInvestidoCusto float64

	for _, ec := range custodiaMap {
		if ec.quantidade > 0.001 {
			pm := 0.0
			if ec.quantidade > 0 {
				pm = ec.custoTotal / ec.quantidade
			}
			if pm < 0 {
				pm = 0
			}

			// Arredonda para 2 casas
			pm = math.Round(pm*100) / 100
			custo := math.Round(ec.custoTotal*100) / 100

			posicoesResultantes = append(posicoesResultantes, PosicaoImportada{
				Ticker:         ec.ticker,
				RazaoSocial:    ec.razaoSocial,
				Classe:         ec.classe,
				Quantidade:     ec.quantidade,
				PrecoMedio:     pm,
				ValorAplicado:  custo,
				Corretora:      ec.corretora,
				TotalCompras:   ec.totalCompras,
				TotalVendas:    ec.totalVendas,
				TotalProventos: ec.totalProventos,
			})
			totalInvestidoCusto += custo
		}
	}

	// Ordena as posições por valor aplicado decrescente
	sort.Slice(posicoesResultantes, func(i, j int) bool {
		return posicoesResultantes[i].ValorAplicado > posicoesResultantes[j].ValorAplicado
	})

	return &RelatorioAnaliseB3{
		TotalMovimentacoes: len(linhasBrutas),
		DataInicio:         dataInicio,
		DataFim:            dataFim,
		TotalAtivosAtivos:  len(posicoesResultantes),
		TotalInvestido:     math.Round(totalInvestidoCusto*100) / 100,
		TotalProventos:     math.Round(totalProventosHistorico*100) / 100,
		Posicoes:           posicoesResultantes,
		Operacoes:          operacoesReconstruidas,
		Proventos:          proventosIdentificados,
	}, nil
}

// ----------------------------------------------------------------------------
// Helpers de Sanitização e Classificação
// ----------------------------------------------------------------------------

func (p *ParserB3) extrairTickerENome(produto string) (string, string) {
	prod := strings.TrimSpace(produto)
	if prod == "" {
		return "", ""
	}

	// Casos especiais de Tesouro Direto
	if strings.HasPrefix(strings.ToUpper(prod), "TESOURO") {
		// ex: "Tesouro IPCA+ com Juros Semestrais 2035" -> Ticker: "Tesouro IPCA+ 2035"
		return prod, "Tesouro Nacional / Renda Fixa"
	}

	parts := strings.SplitN(prod, " - ", 2)
	tickerBruto := strings.TrimSpace(parts[0])
	razao := ""
	if len(parts) > 1 {
		razao = strings.TrimSpace(parts[1])
	}

	// Remove eventual sufixo 'F' de mercado fracionário para ações (ex: PETR4F -> PETR4)
	tickerLimpo := strings.ToUpper(tickerBruto)
	if len(tickerLimpo) == 6 && strings.HasSuffix(tickerLimpo, "F") {
		tickerLimpo = tickerLimpo[:5]
	}

	return tickerLimpo, razao
}

func (p *ParserB3) classificarClasse(ticker, produto string) portfolio.ClasseAtivo {
	t := strings.ToUpper(strings.TrimSpace(ticker))
	pUpper := strings.ToUpper(produto)

	if strings.HasPrefix(t, "TESOURO") || strings.Contains(pUpper, "TESOURO") || strings.Contains(pUpper, "CDB") {
		return portfolio.ClasseRendaFixa
	}

	// BDRs (terminados em 34, 35, 39) ou Ativos Internacionais Globais
	if strings.HasSuffix(t, "34") || strings.HasSuffix(t, "35") || strings.HasSuffix(t, "39") ||
		t == "IVVB11" || t == "WRLD11" || t == "SPYI11" || t == "ACWI11" || t == "COIN11" {
		return portfolio.ClasseInternacional
	}

	// Units conhecidas de Ações que terminam em 11
	acoesUnits := map[string]bool{
		"TAEE11": true, "SAPR11": true, "ALUP11": true, "SANB11": true,
		"BPAC11": true, "KLBN11": true, "ENGI11": true, "CPLE11": true,
		"SULA11": true, "TIET11": true, "MODL11": true, "PSSA11": true,
		"RANI11": true, "ROMI11": true,
	}

	if acoesUnits[t] {
		return portfolio.ClasseAcaoBR
	}

	// Fundos Imobiliários e Fiagros (terminam em 11)
	if strings.HasSuffix(t, "11") {
		return portfolio.ClasseFII
	}

	// Ações Brasileiras (terminam em 3, 4, 5, 6, 7, 8)
	if strings.HasSuffix(t, "3") || strings.HasSuffix(t, "4") ||
		strings.HasSuffix(t, "5") || strings.HasSuffix(t, "6") ||
		strings.HasSuffix(t, "7") || strings.HasSuffix(t, "8") {
		return portfolio.ClasseAcaoBR
	}

	return portfolio.ClasseAcaoBR
}

func (p *ParserB3) isProvento(mov string) bool {
	return strings.Contains(mov, "rendimento") ||
		strings.Contains(mov, "dividendo") ||
		strings.Contains(mov, "juros sobre capital") ||
		strings.Contains(mov, "jcp") ||
		strings.Contains(mov, "restituição de capital") ||
		strings.Contains(mov, "restituicao de capital") ||
		mov == "juros"
}

func (p *ParserB3) mapearTipoProvento(mov string) string {
	switch {
	case strings.Contains(mov, "rendimento"):
		return "RENDIMENTO"
	case strings.Contains(mov, "dividendo"):
		return "DIVIDENDO"
	case strings.Contains(mov, "juros sobre capital") || strings.Contains(mov, "jcp"):
		return "JCP"
	case strings.Contains(mov, "restituição") || strings.Contains(mov, "restituicao"):
		return "RESTITUICAO_CAPITAL"
	case strings.Contains(mov, "juros"):
		return "JUROS"
	default:
		return "RENDIMENTO"
	}
}

func (p *ParserB3) isSubscricao(ticker, mov string) bool {
	t := strings.ToUpper(ticker)
	return strings.Contains(mov, "subscri") ||
		strings.Contains(mov, "cessão de direitos") ||
		strings.Contains(mov, "cessao de direitos") ||
		strings.HasSuffix(t, "12") || strings.HasSuffix(t, "13") ||
		strings.HasSuffix(t, "14") || strings.HasSuffix(t, "15")
}

func (p *ParserB3) converterDireitoParaAtivoBase(ticker string) string {
	t := strings.ToUpper(ticker)
	for _, suf := range []string{"12", "13", "14", "15"} {
		if strings.HasSuffix(t, suf) {
			return strings.TrimSuffix(t, suf) + "11"
		}
	}
	return t
}

func (p *ParserB3) parseData(s string) time.Time {
	s = strings.TrimSpace(s)
	formatos := []string{
		"02/01/2006",
		"2006-01-02",
		"02/01/06",
		"02/01/2006 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, f := range formatos {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (p *ParserB3) parseNumero(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" || s == "N/A" {
		return 0.0
	}

	// Caso de string brasileira: "1.250,50" -> "1250.50"
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}
