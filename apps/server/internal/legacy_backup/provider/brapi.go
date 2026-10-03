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

// Estrutura para mapear a resposta JSON da API de mercado
type brapiQuoteResponse struct {
	Results []struct {
		Symbol             string  `json:"symbol"`
		RegularMarketPrice float64 `json:"regularMarketPrice"`
		DividendsData      *struct {
			CashDividends []struct {
				Rate float64 `json:"rate"`
			} `json:"cashDividends"`
		} `json:"dividendsData"`
	} `json:"results"`
}

type BrapiProvider struct {
	client *http.Client
}

// NovoBrapiProvider cria um cliente HTTP com Timeout de 10s (Boa prática de produção)
func NovoBrapiProvider() *BrapiProvider {
	return &BrapiProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BuscarAtivo faz a chamada HTTP para obter cotação e proventos reais
func (p *BrapiProvider) BuscarAtivo(ticker string) (*domain.Ativo, error) {
	url := fmt.Sprintf("https://brapi.dev/api/quote/%s?dividends=true", ticker)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar ticker %s: %w", ticker, err)
	}
	defer resp.Body.Close() // Garante que a conexão TCP é fechada ao sair da função

	var dados brapiQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON para %s: %w", ticker, err)
	}

	if len(dados.Results) == 0 {
		return nil, fmt.Errorf("nenhum dado retornado para %s", ticker)
	}

	item := dados.Results[0]

	// Somamos os últimos proventos informados
	totalProventos := 0.0
	if item.DividendsData != nil {
		// Limitamos a soma aos proventos mais recentes
		for i, d := range item.DividendsData.CashDividends {
			if i >= 6 { // Considera os dividendos mais recentes
				break
			}
			totalProventos += d.Rate
		}
	}

	// Identificamos se é Ação ou FII com base no final do código (11 = FII/ETF, 3/4 = Ações)
	classe := domain.ClasseAcao
	if strings.HasSuffix(item.Symbol, "11") {
		classe = domain.ClasseFII
	}

	return &domain.Ativo{
		Ticker:        item.Symbol,
		Classe:        classe,
		PrecoAtual:    item.RegularMarketPrice,
		Dividendos12M: totalProventos,
	}, nil
}

// BuscarEmLote utiliza Goroutines, WaitGroup e Channels para buscar em paralelo
func (p *BrapiProvider) BuscarEmLote(tickers []string) []*domain.Ativo {
	var wg sync.WaitGroup
	// Criamos um Canal com buffer para receber os ativos encontrados
	canalAtivos := make(chan *domain.Ativo, len(tickers))

	for _, t := range tickers {
		wg.Add(1) // Avisa ao WaitGroup que uma nova goroutine vai começar

		// Disparamos a Goroutine com a palavra 'go'
		// Passamos 't' como parâmetro para evitar o bug de closure de loops
		go func(ticker string) {
			defer wg.Done() // Avisa ao WaitGroup quando a goroutine termina

			ativo, err := p.BuscarAtivo(ticker)
			if err != nil {
				fmt.Printf("⚠️ [HTTP Error] %s\n", err)
				return
			}

			// Enviamos o ativo encontrado para dentro do canal com segurança
			canalAtivos <- ativo
		}(t)
	}

	// Goroutine auxiliar para fechar o canal assim que todos os downloads terminarem
	go func() {
		wg.Wait()
		close(canalAtivos)
	}()

	// Consumimos os ativos que vão chegando pelo canal de forma segura
	var carteira []*domain.Ativo
	for ativo := range canalAtivos {
		carteira = append(carteira, ativo)
	}

	return carteira
}
