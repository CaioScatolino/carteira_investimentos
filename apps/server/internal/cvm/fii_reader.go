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

// CNPJs oficiais registrados na CVM dos principais FIIs da B3
var CNPJPorTicker = map[string]string{
	"MXRF11": "13.318.527/0001-79",
	"HGLG11": "11.728.688/0001-47",
	"XPML11": "28.750.990/0001-74",
	"KNRI11": "12.005.956/0001-65",
	"BTLG11": "13.111.782/0001-84",
	"VISC11": "17.554.274/0001-25",
	"XPLG11": "26.502.794/0001-85",
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

// ObterValoresPatrimoniais baixa o ZIP oficial do ano corrente e extrai o VP/Cota mais recente
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

	// Lê o ZIP em memória (apenas ~1 MB)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	// Mapeia CNPJ -> Ticker inverso para busca rápida O(1)
	tickerPorCNPJ := make(map[string]string)
	for t, cnpj := range CNPJPorTicker {
		cnpjLimpo := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cnpj, ".", ""), "/", ""), "-", "")
		tickerPorCNPJ[cnpjLimpo] = t
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

		// Pula o cabeçalho
		_, _ = reader.Read()

		for {
			linha, err := reader.Read()
			if err != nil {
				break // Fim do ficheiro
			}

			if len(linha) <= 23 {
				continue
			}

			// Coluna 0: CNPJ | Coluna 23: Valor_Patrimonial_Cotas
			cnpjLinha := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(linha[0], ".", ""), "/", ""), "-", "")

			if ticker, encontrado := tickerPorCNPJ[cnpjLinha]; encontrado {
				vpStr := strings.TrimSpace(linha[23])
				if vp, err := strconv.ParseFloat(vpStr, 64); err == nil && vp > 0 {
					// Guarda o VP mais recente
					resultadoVP[ticker] = vp
				}
			}
		}
		break
	}

	return resultadoVP, nil
}
