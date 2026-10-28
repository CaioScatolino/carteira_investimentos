package technical

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"carteira_investimentos/server/internal/storage"
)

// YahooFinanceChartResponse mapeia a estrutura v8 do Yahoo Finance
type YahooFinanceChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				Symbol             string  `json:"symbol"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				LongName           string  `json:"longName"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					High   []*float64 `json:"high"`
					Low    []*float64 `json:"low"`
					Close  []*float64 `json:"close"`
					Volume []*float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

type Service struct {
	httpClient *http.Client
	cache      storage.CacheStore
	store      storage.SnapshotStore
}

func NovoService(cache storage.CacheStore, store storage.SnapshotStore) *Service {
	return &Service{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: cache,
		store: store,
	}
}

// ObterDadosGrafico busca os candles no Yahoo Finance com cache no Redis e calcula indicadores
func (s *Service) ObterDadosGrafico(ctx context.Context, ticker, intervalo, periodo string) (*RespostaGrafico, error) {
	tickerLimpo := strings.ToUpper(strings.TrimSpace(ticker))
	if tickerLimpo == "" {
		return nil, errors.New("ticker não pode ser vazio")
	}

	// Normaliza intervalo
	if intervalo != "1d" && intervalo != "1wk" {
		intervalo = "1d"
	}

	// Normaliza período
	if periodo != "3mo" && periodo != "6mo" && periodo != "1y" && periodo != "2y" && periodo != "5y" {
		periodo = "1y"
	}

	chaveCache := fmt.Sprintf("b3:chart:v3:%s:%s:%s", tickerLimpo, intervalo, periodo)

	// 1. Tenta recuperar do Cache Redis / In-Memory
	if s.cache != nil {
		if cachedBytes, err := s.cache.ObterCache(ctx, chaveCache); err == nil && len(cachedBytes) > 0 {
			var respCache RespostaGrafico
			if err := json.Unmarshal(cachedBytes, &respCache); err == nil {
				return &respCache, nil
			}
		}
	}

	// 2. Busca no Yahoo Finance v8 API
	symbolYahoo := tickerLimpo
	if !strings.HasSuffix(symbolYahoo, ".SA") {
		symbolYahoo += ".SA"
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=%s&range=%s", symbolYahoo, intervalo, periodo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição para o Yahoo Finance: %w", err)
	}

	// Header User-Agent obrigatório para evitar 429 Too Many Requests no Yahoo
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro de rede ao consultar Yahoo Finance (%s): %w", symbolYahoo, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		corpo, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("yahoo finance retornou status %d: %s", res.StatusCode, string(corpo))
	}

	var yfResp YahooFinanceChartResponse
	if err := json.NewDecoder(res.Body).Decode(&yfResp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON do Yahoo Finance: %w", err)
	}

	if len(yfResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("nenhum dado gráfico retornado para %s", tickerLimpo)
	}

	resultado := yfResp.Chart.Result[0]
	if len(resultado.Timestamp) == 0 || len(resultado.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("série temporal vazia para %s", tickerLimpo)
	}

	quotes := resultado.Indicators.Quote[0]
	numPontos := len(resultado.Timestamp)

	var candles []Candle
	for i := 0; i < numPontos; i++ {
		// Valida se os campos não são nulos no ponto
		if i >= len(quotes.Open) || i >= len(quotes.High) || i >= len(quotes.Low) || i >= len(quotes.Close) {
			continue
		}
		if quotes.Open[i] == nil || quotes.High[i] == nil || quotes.Low[i] == nil || quotes.Close[i] == nil {
			continue
		}

		cClose := *quotes.Close[i]
		if cClose <= 0 {
			continue
		}

		cOpen := *quotes.Open[i]
		cHigh := *quotes.High[i]
		cLow := *quotes.Low[i]
		cVolume := 0.0
		if i < len(quotes.Volume) && quotes.Volume[i] != nil {
			cVolume = *quotes.Volume[i]
		}

		candles = append(candles, Candle{
			Time:   resultado.Timestamp[i],
			Open:   cOpen,
			High:   cHigh,
			Low:    cLow,
			Close:  cClose,
			Volume: cVolume,
		})
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("nenhum candle válido processado para %s", tickerLimpo)
	}

	// Ordena cronologicamente
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Time < candles[j].Time
	})

	// 3. Calcula os Indicadores Técnicos
	mma8 := CalcularMMA(candles, 8)
	mma21 := CalcularMMA(candles, 21)
	mma200 := CalcularMMA(candles, 200)
	rsi14 := CalcularRSI(candles, 14)

	// 4. Cruzamento Fundamentalista com Snapshot B3 do Ativo
	precoAtual := candles[len(candles)-1].Close
	var precoTetoConsolidado, precoTetoBazin, valorGraham, margemSeguranca float64
	var modelosTeto []string
	nomeAtivo := tickerLimpo

	if s.store != nil {
		if ativo, err := s.store.Obter(tickerLimpo); err == nil && ativo != nil {
			if ativo.Nome != "" {
				nomeAtivo = ativo.Nome
			}
			if ativo.PrecoAtual > 0 {
				precoAtual = ativo.PrecoAtual
			}
			precoTetoConsolidado = ativo.PrecoTetoConsolidado
			modelosTeto = ativo.ModelosTetoConsolidado
			precoTetoBazin = ativo.PrecoTetoBazin
			if ativo.PrecoTetoBazin5A > 0 {
				precoTetoBazin = ativo.PrecoTetoBazin5A
			}
			// Fallback caso PrecoTetoConsolidado ainda não tenha sido computado
			if precoTetoConsolidado == 0 {
				if precoTetoBazin > 0 {
					precoTetoConsolidado = precoTetoBazin
				} else if ativo.ValorGraham > 0 {
					precoTetoConsolidado = ativo.ValorGraham
				}
			}
			valorGraham = ativo.ValorGraham
			margemSeguranca = ativo.PremioDescontoPercentual
			if margemSeguranca == 0 && ativo.MargemSegurancaConsolidada != 0 {
				margemSeguranca = ativo.MargemSegurancaConsolidada
			}
			if margemSeguranca == 0 && precoTetoConsolidado > 0 && precoAtual > 0 {
				margemSeguranca = ((precoTetoConsolidado - precoAtual) / precoAtual) * 100
			}
		}
	}

	// 5. Avaliação do Sinal Técnico
	var ultimoRSI, ultimaMMA21, ultimaMMA200 float64
	if len(rsi14) > 0 {
		ultimoRSI = rsi14[len(rsi14)-1].Value
	}
	if len(mma21) > 0 {
		ultimaMMA21 = mma21[len(mma21)-1].Value
	}
	if len(mma200) > 0 {
		ultimaMMA200 = mma200[len(mma200)-1].Value
	}

	tetoReferencia := precoTetoConsolidado
	if tetoReferencia == 0 {
		tetoReferencia = precoTetoBazin
	}

	sinal, desc := AvaliarSinalTecnico(precoAtual, tetoReferencia, ultimoRSI, ultimaMMA21, ultimaMMA200)

	resp := RespostaGrafico{
		Ticker:               tickerLimpo,
		Nome:                 nomeAtivo,
		Intervalo:            intervalo,
		Periodo:              periodo,
		PrecoAtual:           precoAtual,
		PrecoTetoConsolidado: precoTetoConsolidado,
		ModelosTeto:          modelosTeto,
		PrecoTetoBazin:       precoTetoBazin,
		ValorGraham:          valorGraham,
		MargemSeguranca:      margemSeguranca,
		SinalTecnico:         sinal,
		DescricaoSinal:       desc,
		Candles:              candles,
		MMA8:                 mma8,
		MMA21:                mma21,
		MMA200:               mma200,
		RSI14:                rsi14,
	}

	// 6. Grava no Cache Redis / In-Memory por 4 horas
	if s.cache != nil {
		if dadosBytes, err := json.Marshal(resp); err == nil {
			_ = s.cache.SalvarCache(ctx, chaveCache, dadosBytes, 4*time.Hour)
		}
	}

	return &resp, nil
}
