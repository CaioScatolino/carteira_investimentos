package statusinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL            = "https://statusinvest.com.br/category/advancedsearchresultpaginated"
	defaultUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	defaultTimeout     = 30 * time.Second
	CategoryTypeAcoes  = 1
	CategoryTypeFIIs   = 2
)

// Client gerencia as requisições otimizadas para a API interna do StatusInvest
type Client struct {
	httpClient *http.Client
}

// NovoClient instancia o cliente com configurações de resiliência e headers padrão de navegador
func NovoClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				IdleConnTimeout:     60 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

// ObterAcoes consulta todos os ativos de ações brasileiras disponíveis no StatusInvest
func (c *Client) ObterAcoes(ctx context.Context) ([]AcaoStatusInvest, error) {
	var resp QueryResponse[AcaoStatusInvest]
	referer := "https://statusinvest.com.br/acoes/busca-avancada"

	if err := c.executarBusca(ctx, CategoryTypeAcoes, referer, &resp); err != nil {
		return nil, fmt.Errorf("falha ao obter ações do StatusInvest: %w", err)
	}

	return resp.List, nil
}

// ObterFIIs consulta todos os fundos imobiliários disponíveis no StatusInvest
func (c *Client) ObterFIIs(ctx context.Context) ([]FIIStatusInvest, error) {
	var resp QueryResponse[FIIStatusInvest]
	referer := "https://statusinvest.com.br/fundos-imobiliarios/busca-avancada"

	if err := c.executarBusca(ctx, CategoryTypeFIIs, referer, &resp); err != nil {
		return nil, fmt.Errorf("falha ao obter FIIs do StatusInvest: %w", err)
	}

	return resp.List, nil
}

// executarBusca dispara o POST codificado em form-urlencoded com paginação ampla (take=1000)
func (c *Client) executarBusca(ctx context.Context, categoryType int, referer string, destino any) error {
	form := url.Values{}
	form.Set("search", "{}")
	form.Set("CategoryType", strconv.Itoa(categoryType))
	form.Set("page", "0")
	form.Set("take", "1000") // Garante o universo completo em apenas 1 requisição

	reqURL := fmt.Sprintf("%s?CategoryType=%d", baseURL, categoryType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	// Headers idênticos aos de uma requisição legítima do navegador via AJAX
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", "https://statusinvest.com.br")

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		corpo, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("StatusInvest retornou HTTP %d: %s", httpResp.StatusCode, string(corpo))
	}

	decoder := json.NewDecoder(httpResp.Body)
	if err := decoder.Decode(destino); err != nil {
		return fmt.Errorf("erro ao decodificar JSON do StatusInvest: %w", err)
	}

	return nil
}
