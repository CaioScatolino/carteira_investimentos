package b3

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseCOTAHIST descompacta em memória e lê o layout posicional linha a linha
func ParseCOTAHIST(zipBytes []byte) (map[string]*CotacaoDiaria, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("falha ao descompactar ZIP da B3: %w", err)
	}

	mapa := make(map[string]*CotacaoDiaria)

	for _, f := range zipReader.File {
		// Procura o arquivo texto principal COTAHIST_D...TXT
		if !strings.HasSuffix(f.Name, ".TXT") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		scanner := bufio.NewScanner(rc)

		for scanner.Scan() {
			linha := scanner.Text()
			// Linhas de detalhe possuem no mínimo 188 caracteres
			if len(linha) < 188 {
				continue
			}

			// TIPREG (Pos 0..2): "01" = Cotação de Ativo (ignora 00 Header e 99 Trailer)
			tipreg := linha[0:2]
			if tipreg != "01" {
				continue
			}

			// CODBDI (Pos 10..12): "02" = Lote Padrão Ações/Units, "12" = Fundos Imobiliários, "14" = ETFs
			codbdi := strings.TrimSpace(linha[10:12])
			if codbdi != "02" && codbdi != "12" && codbdi != "14" {
				continue // Ignora opções, fracionários e leilões especiais
			}

			// CODNEG (Pos 12..24): Ticker do ativo
			ticker := strings.TrimSpace(linha[12:24])
			nome := strings.TrimSpace(linha[27:39])

			// Data (Pos 2..10): AAAAMMDD
			dataStr := linha[2:10]
			dataPregao, _ := time.Parse("20060102", dataStr)

			// Helper para converter formato implícito v99 (divide por 100.0)
			parsePreco := func(inicio, fim int) float64 {
				val, _ := strconv.ParseFloat(strings.TrimSpace(linha[inicio:fim]), 64)
				return val / 100.0
			}

			precoAbe := parsePreco(56, 69)
			precoMax := parsePreco(69, 82)
			precoMin := parsePreco(82, 95)
			precoMed := parsePreco(95, 108)
			precoUlt := parsePreco(108, 121) // Preço de Fechamento oficial

			// Volume financeiro total (Pos 170..188): Formato 16v99
			volTot := parsePreco(170, 188)

			// Quantidade de negócios (Pos 147..152)
			qtdNeg, _ := strconv.Atoi(strings.TrimSpace(linha[147:152]))

			mapa[ticker] = &CotacaoDiaria{
				Data:            dataPregao,
				Ticker:          ticker,
				NomeResumido:    nome,
				PrecoAbertura:   precoAbe,
				PrecoMaximo:     precoMax,
				PrecoMinimo:     precoMin,
				PrecoMedio:      precoMed,
				PrecoFechamento: precoUlt,
				VolumeTotal:     volTot,
				QuantidadeNeg:   qtdNeg,
				CodigoBDI:       codbdi,
			}
		}
		break
	}

	return mapa, nil
}
