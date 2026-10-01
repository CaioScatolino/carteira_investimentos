package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"carteira_investimentos/server/internal/domain"
)

// Estrutura que mapeia a resposta JSON do Yahoo Finance
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol             string  `json:"symbol"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
			Events struct {
				Dividends map[string]struct {
					Amount float64 `json:"amount"`
				} `json:"dividends"`
			} `json:"events"`
		} `json:"result"`
	} `json:"chart"`
}

type YahooFinanceProvider struct {
	client *http.Client
}

func NovoYahooFinanceProvider() *YahooFinanceProvider {
	return &YahooFinanceProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BuscarAtivo consulta a cotação e todos os proventos dos últimos 12 meses
func (y *YahooFinanceProvider) BuscarAtivo(ticker string) (*domain.Ativo, error) {
	// A B3 no Yahoo Finance utiliza o sufixo .SA (ex: PETR4.SA)
	tickerFormatado := ticker
	if !strings.HasSuffix(tickerFormatado, ".SA") {
		tickerFormatado = ticker + ".SA"
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?events=div&range=1y&interval=1d", tickerFormatado)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Header essencial para evitar bloqueio por WAF
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro de rede para %s: %w", ticker, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status HTTP %d retornado para %s", resp.StatusCode, ticker)
	}

	var dados yahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		return nil, fmt.Errorf("falha no decode JSON para %s: %w", ticker, err)
	}

	if len(dados.Chart.Result) == 0 {
		return nil, fmt.Errorf("nenhum resultado encontrado para %s", ticker)
	}

	item := dados.Chart.Result[0]

	// Somamos todos os proventos reais pagos nos últimos 12 meses
	totalProventos := 0.0
	for _, div := range item.Events.Dividends {
		totalProventos += div.Amount
	}

	// Por isto (Regra inteligente B3: Units de ações NÃO são FIIs):
	classe := domain.ClasseAcao
	if strings.HasSuffix(ticker, "11") {
		// Units de Ações consagradas na B3
		unitsAcoes := map[string]bool{
			"TAEE11": true,
			"SAPR11": true,
			"KLBN11": true,
			"ALUP11": true,
			"BPAC11": true,
			"SANB11": true,
		}
		if !unitsAcoes[ticker] {
			classe = domain.ClasseFII
		}
	}

	return &domain.Ativo{
		Ticker:        ticker,
		Classe:        classe,
		PrecoAtual:    item.Meta.RegularMarketPrice,
		Dividendos12M: totalProventos,
	}, nil
}

// BuscarEmLote dispara Goroutines concorrentes com WaitGroup e Channel
func (y *YahooFinanceProvider) BuscarEmLote(tickers []string) []*domain.Ativo {
	var wg sync.WaitGroup
	canal := make(chan *domain.Ativo, len(tickers))

	for _, t := range tickers {
		wg.Add(1)

		go func(ticker string) {
			defer wg.Done()

			ativo, err := y.BuscarAtivo(ticker)
			if err != nil {
				fmt.Printf("⚠️ [Yahoo Error] %s\n", err)
				return
			}

			canal <- ativo
		}(t)
	}

	go func() {
		wg.Wait()
		close(canal)
	}()

	var lista []*domain.Ativo
	for ativo := range canal {
		lista = append(lista, ativo)
	}

	return lista
}
