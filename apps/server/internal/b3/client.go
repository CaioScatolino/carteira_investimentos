package b3

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

type B3Client struct {
	httpClient *http.Client
}

func NovoB3Client() *B3Client {
	return &B3Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// ObterArquivoMaisRecente busca retroativamente o último pregão disponível da B3
func (c *B3Client) ObterArquivoMaisRecente() ([]byte, string, error) {
	hoje := time.Now()

	// Tenta até 7 dias úteis anteriores (cobrindo feriados prolongados como Carnaval/Páscoa)
	for i := 0; i < 7; i++ {
		diaTentativa := hoje.AddDate(0, 0, -i)

		// Pula sábado e domingo antes mesmo de bater na rede
		if diaTentativa.Weekday() == time.Saturday || diaTentativa.Weekday() == time.Sunday {
			continue
		}

		// Padrão oficial da B3: DDMMAAAA (ex: 01102026)
		dataStr := diaTentativa.Format("02012006")
		url := fmt.Sprintf("https://bvmf.bmfbovespa.com.br/InstDados/SerHist/COTAHIST_D%s.ZIP", dataStr)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, "", err
			}

			dataFormatada := diaTentativa.Format("02/01/2006")
			return body, dataFormatada, nil
		}
		resp.Body.Close()
	}

	return nil, "", fmt.Errorf("nenhum pregão recente encontrado nos últimos 7 dias nos servidores da B3")
}
