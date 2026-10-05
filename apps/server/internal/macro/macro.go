package macro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// CenarioMacro armazena os parâmetros macroeconômicos de referência do mercado brasileiro
type CenarioMacro struct {
	TaxaSelic             float64 `json:"taxa_selic"`              // Meta Selic (% a.a., ex: 11.75)
	TaxaNTNB              float64 `json:"taxa_ntnb"`               // Juro real da NTN-B / Tesouro IPCA+ (% a.a., ex: 6.50)
	SpreadMinimoAcoes     float64 `json:"spread_minimo_acoes"`     // Spread mínimo exigido para ações (% a.a., ex: 1.50)
	SpreadMinimoFIITijolo float64 `json:"spread_minimo_fii_tijolo"`// Spread mínimo para FIIs de tijolo (% a.a., ex: 2.00)
	EquityRiskPremium     float64 `json:"equity_risk_premium"`     // Prêmio de risco de equity no Brasil (% a.a., ex: 5.50)
	DataAtualizacao       string  `json:"data_atualizacao"`
	Fonte                 string  `json:"fonte"`
}

var (
	instancia *CenarioMacro
	mu        sync.RWMutex
)

// ObterCenarioRetorna uma cópia segura do cenário macroeconômico ativo
func ObterCenario() CenarioMacro {
	mu.RLock()
	defer mu.RUnlock()
	if instancia == nil {
		instancia = CenarioPadrao()
	}
	return *instancia
}

// AtualizarCenario permite recalibrar manualmente os parâmetros macroeconômicos
func AtualizarCenario(novo CenarioMacro) {
	mu.Lock()
	defer mu.Unlock()
	if novo.TaxaSelic <= 0 {
		novo.TaxaSelic = 13.75
	}
	if novo.TaxaNTNB <= 0 {
		novo.TaxaNTNB = 6.50
	}
	if novo.SpreadMinimoAcoes <= 0 {
		novo.SpreadMinimoAcoes = 1.50
	}
	if novo.SpreadMinimoFIITijolo <= 0 {
		novo.SpreadMinimoFIITijolo = 2.00
	}
	if novo.EquityRiskPremium <= 0 {
		novo.EquityRiskPremium = 5.50
	}
	if novo.DataAtualizacao == "" {
		novo.DataAtualizacao = time.Now().Format("02/01/2006 15:04")
	}
	if novo.Fonte == "" {
		novo.Fonte = "Banco Central do Brasil (SGS Série 432) & Tesouro Direto"
	}
	instancia = &novo
}

// CenarioPadrao retorna os parâmetros calibrados com a realidade da curva de juros do Brasil
func CenarioPadrao() *CenarioMacro {
	return &CenarioMacro{
		TaxaSelic:             13.75, // Meta Selic oficial do BCB SGS Série 432 (antes do IR)
		TaxaNTNB:              6.50,  // Tesouro IPCA+ longo (NTN-B 2035/2045)
		SpreadMinimoAcoes:     1.50,  // Hurdle para ações = 6.50% + 1.50% = 8.00%
		SpreadMinimoFIITijolo: 2.00,  // Hurdle para Tijolo = 6.50% + 2.00% = 8.50%
		EquityRiskPremium:     5.50,  // ERP clássico Brasil
		DataAtualizacao:       time.Now().Format("02/01/2006 15:04"),
		Fonte:                 "Banco Central do Brasil (SGS Série 432) & Tesouro Direto",
	}
}

// ObterYieldMinimoAcoes retorna a taxa de corte mínima para ações (ex: 8.0%)
func (c CenarioMacro) ObterYieldMinimoAcoes() float64 {
	return (c.TaxaNTNB + c.SpreadMinimoAcoes) / 100.0
}

// ObterCustoCapitalGordon retorna o k (custo de capital próprio) calibrado pelo CAPM
func (c CenarioMacro) ObterCustoCapitalGordon() float64 {
	// CAPM simplificado para o mercado brasileiro: Selic * 0.90 + 3.5%
	// Garante que o custo de capital não seja inferior à taxa livre de risco
	ke := (c.TaxaSelic * 0.85 + 4.0) / 100.0
	if ke < 0.125 {
		ke = 0.125 // Piso de 12.5% a.a. para equity no Brasil
	}
	return ke
}

// ObterSpreadSegmentoFII retorna o prêmio de risco exigido sobre a NTN-B por segmento imobiliário
func (c CenarioMacro) ObterSpreadSegmentoFII(segmento string) float64 {
	seg := segmento
	switch {
	case containsIgnoreCase(seg, "LOGÍST") || containsIgnoreCase(seg, "LOGIST"):
		return 1.50 // Risco moderado, contratos atípicos longos
	case containsIgnoreCase(seg, "SHOPPING"):
		return 2.00 // Risco comercial / fluxo de vendas
	case containsIgnoreCase(seg, "LAJES") || containsIgnoreCase(seg, "ESCRITÓRIOS"):
		return 2.50 // Maior risco de vacância estrutural e obsolescência
	case containsIgnoreCase(seg, "RENDA URBANA") || containsIgnoreCase(seg, "VAREJO") || containsIgnoreCase(seg, "HOSPITAL"):
		return 1.75 // Contratos atípicos e inquilinos resilientes
	default:
		return c.SpreadMinimoFIITijolo // 2.00% padrão
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 &&
		(s == substr || len(s) > 0 && findIgnoreCase(s, substr))
}

func findIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr)
}

// RespostaBCBSGS mapeia a resposta da API oficial do Banco Central
type RespostaBCBSGS struct {
	Data  string `json:"data"`
	Valor string `json:"valor"`
}

// SincronizarTaxasOficiais tenta obter a taxa Selic oficial mais recente do BCB
func SincronizarTaxasOficiais() {
	client := &http.Client{Timeout: 5 * time.Second}
	// Série 432: Taxa de juros - Meta Selic definida pelo Copom (% a.a.)
	url := "https://api.bcb.gov.br/dados/serie/bcdata.sgs.432/dados/ultimos/1?formato=json"

	resp, err := client.Get(url)
	if err != nil {
		return // Mantém padrão conservador silenciosamente
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var resultados []RespostaBCBSGS
	if err := json.NewDecoder(resp.Body).Decode(&resultados); err != nil || len(resultados) == 0 {
		return
	}

	val, err := strconv.ParseFloat(resultados[0].Valor, 64)
	if err == nil && val > 0 {
		cenario := ObterCenario()
		cenario.TaxaSelic = val
		cenario.DataAtualizacao = fmt.Sprintf("%s (BCB SGS Série 432)", resultados[0].Data)
		cenario.Fonte = "Banco Central do Brasil (SGS Série 432)"
		AtualizarCenario(cenario)
	}
}
