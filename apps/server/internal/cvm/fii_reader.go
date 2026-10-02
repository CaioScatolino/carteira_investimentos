package cvm

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CNPJs oficiais dos principais FIIs de referência na CVM
var CNPJPorTicker = map[string]string{
	"MXRF11": "97521225000125",
	"HGLG11": "11728688000147",
	"XPML11": "28757546000100",
	"KNRI11": "12005956000165",
	"BTLG11": "11839593000109",
	"VISC11": "17554274000125",
	"XPLG11": "26502794000185",
}

type CVMClient struct {
	httpClient *http.Client
}

func NovoCVMClient() *CVMClient {
	return &CVMClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ObterValoresPatrimoniais baixa o ZIP oficial do ano corrente e extrai o VP/Cota de todos os FIIs
func (c *CVMClient) ObterValoresPatrimoniais() (map[string]float64, error) {
	ano := time.Now().Year()
	url := fmt.Sprintf("https://dados.cvm.gov.br/dados/FII/DOC/INF_MENSAL/DADOS/inf_mensal_fii_%d.zip", ano)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar à CVM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVM retornou HTTP %d para o ano %d", resp.StatusCode, ano)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	// Mapeia CNPJ -> Ticker
	tickerPorCNPJ := make(map[string]string)
	for t, cnpj := range CNPJPorTicker {
		cnpjLimpo := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cnpj, ".", ""), "/", ""), "-", "")
		tickerPorCNPJ[cnpjLimpo] = t
	}

	cleanCNPJ := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), "/", ""), "-", "")
	}

	nomeCSV := fmt.Sprintf("inf_mensal_fii_complemento_%d.csv", ano)
	resultadoVP := make(map[string]float64)

	for _, file := range zipReader.File {
		if file.Name != nomeCSV {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		reader := csv.NewReader(rc)
		reader.Comma = ';'
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1

		header, _ := reader.Read()
		idxVP := 23
		for i, col := range header {
			if col == "Valor_Patrimonial_Cotas" {
				idxVP = i
				break
			}
		}

		for {
			linha, err := reader.Read()
			if err != nil {
				break
			}
			if len(linha) > idxVP {
				cnpj := cleanCNPJ(linha[0])
				if ticker, ok := tickerPorCNPJ[cnpj]; ok {
					if vp, err := strconv.ParseFloat(linha[idxVP], 64); err == nil && vp > 0 {
						resultadoVP[ticker] = vp
					}
				}
			}
		}
		break
	}

	return resultadoVP, nil
}
