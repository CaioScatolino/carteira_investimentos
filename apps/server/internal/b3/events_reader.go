package b3

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"carteira_investimentos/server/internal/domain"
)

// B3ProventoItem mapeia a estrutura oficial de evento corporativo em dinheiro da B3
type B3ProventoItem struct {
	TypeStock           string `json:"typeStock"`           // "ON", "PN", etc.
	ValueCash           string `json:"valueCash"`           // Valor por ação (ex: "0,0242425")
	CorporateAction     string `json:"corporateAction"`     // "DIVIDENDO", "JRS CAP PROPRIO"
	LastDatePriorEx     string `json:"lastDatePriorEx"`     // Data COM (DD/MM/YYYY)
	LastDateTimePriorEx string `json:"lastDateTimePriorEx"` // Data COM formato ISO
	DateApproval        string `json:"dateApproval"`        // Data de Deliberação
}

type B3ProventosResponse struct {
	Results []B3ProventoItem `json:"results"`
}

var (
	cacheProventosAcoes   = make(map[string]float64)
	cacheProventosAcoesMu sync.RWMutex
)

// ObterProventos12MAcao busca dinamicamente o histórico oficial de eventos em dinheiro da B3
// e calcula a somatória líquida dos últimos 12 meses (aplicando 15% de IR sobre JCP - Regra Décio Bazin)
func (c *B3Client) ObterProventos12MAcao(tradingName, ticker string) float64 {
	cacheProventosAcoesMu.RLock()
	if val, ok := cacheProventosAcoes[ticker]; ok {
		cacheProventosAcoesMu.RUnlock()
		return val
	}
	cacheProventosAcoesMu.RUnlock()

	nome := strings.TrimSpace(tradingName)
	if nome == "" && len(ticker) >= 4 {
		nome = ticker[:4]
	}

	payload := fmt.Sprintf(`{"tradingName":"%s","language":"pt-br"}`, nome)
	b64 := base64.StdEncoding.EncodeToString([]byte(payload))
	url := fmt.Sprintf("https://sistemaswebb3-listados.b3.com.br/listedCompaniesProxy/CompanyCall/GetListedCashDividends/%s", b64)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0.0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := c.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return 0.0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0
	}

	var b3Resp B3ProventosResponse
	if err := json.Unmarshal(body, &b3Resp); err != nil || len(b3Resp.Results) == 0 {
		return 0.0
	}

	// Identifica classe da ação pelo ticker B3:
	// Ticker final 3 = ON (Ordinária)
	// Ticker final 4, 5, 6 = PN (Preferencial)
	// Ticker final 11 = UNT (Unit)
	tipoAlvo := "ON"
	if strings.HasSuffix(ticker, "4") || strings.HasSuffix(ticker, "5") || strings.HasSuffix(ticker, "6") {
		tipoAlvo = "PN"
	} else if strings.HasSuffix(ticker, "11") {
		tipoAlvo = "UNT"
	}

	var eventos []domain.EventoProvento
	for _, it := range b3Resp.Results {
		// Filtra classe da ação se não for Unit
		if it.TypeStock != "" && tipoAlvo != "UNT" && !strings.Contains(it.TypeStock, tipoAlvo) {
			continue
		}

		// Prioriza Data COM para a janela de 12 meses
		dtStr := it.LastDateTimePriorEx
		if dtStr == "" {
			dtStr = it.LastDatePriorEx
		}
		if dtStr == "" {
			dtStr = it.DateApproval
		}

		var dt time.Time
		if strings.Contains(dtStr, "T") {
			dt, _ = time.Parse("2006-01-02T15:04:05", dtStr)
		} else if strings.Contains(dtStr, "/") && len(dtStr) >= 10 {
			dt, _ = time.Parse("02/01/2006", dtStr[:10])
		}

		if dt.IsZero() {
			continue
		}

		valStr := strings.ReplaceAll(strings.ReplaceAll(it.ValueCash, ".", ""), ",", ".")
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil || val <= 0 {
			continue
		}

		tipo := domain.ProventoDividendo
		if strings.Contains(strings.ToUpper(it.CorporateAction), "JRS") || strings.Contains(strings.ToUpper(it.CorporateAction), "JUROS") {
			tipo = domain.ProventoJCP
		}

		eventos = append(eventos, domain.EventoProvento{
			Identificador: ticker,
			Tipo:          tipo,
			DataPagamento: dt,
			ValorUnitario: val,
		})
	}

	// Calcula proventos líquidos nos últimos 365 dias
	soma := domain.CalcularProventos12M(eventos, time.Now())

	cacheProventosAcoesMu.Lock()
	cacheProventosAcoes[ticker] = soma
	cacheProventosAcoesMu.Unlock()

	return soma
}
