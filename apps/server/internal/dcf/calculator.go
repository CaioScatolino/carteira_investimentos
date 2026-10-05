package dcf

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/macro"
)

// IsSetorFinanceiro identifica se o ativo pertence ao setor bancário, seguros ou serviços financeiros.
// Bancos e instituições financeiras não possuem EBIT/Dívida Líquida padrão da firma (FCFF/WACC),
// sendo avaliados exclusivamente por Dividendos/Patrimônio (Bazin 5A, Gordon e Lynch).
func IsSetorFinanceiro(setor, subsetor, segmento string) bool {
	s := strings.ToUpper(setor + " " + subsetor + " " + segmento)
	return strings.Contains(s, "FINANC") ||
		strings.Contains(s, "BANCO") ||
		strings.Contains(s, "SEGURO") ||
		strings.Contains(s, "PREVIDÊNCIA") ||
		strings.Contains(s, "PREVIDENCIA") ||
		strings.Contains(s, "INTERMEDIÁRIOS") ||
		strings.Contains(s, "INTERMEDIARIOS") ||
		strings.Contains(s, "SERVIÇOS FINANCEIROS") ||
		strings.Contains(s, "SERVICOS FINANCEIROS")
}

// CalcularDCF calcula o Preço Teto por Ação baseado em Fluxo de Caixa Descontado (FCFF / WACC Damodaran Proxy).
func CalcularDCF(ativo *domain.Ativo, cenario macro.CenarioMacro) (float64, float64, float64, float64, error) {
	if ativo.Classe != domain.ClasseAcao {
		return 0, 0, 0, 0, nil
	}

	// Setores Financeiros: exclusão estrita do DCF da firma
	if IsSetorFinanceiro(ativo.Setor, ativo.Subsetor, ativo.Segmento) {
		ativo.PrecoTetoDCF = 0
		return 0, 0, 0, 0, nil
	}

	if ativo.PrecoAtual <= 0 || ativo.ValorMercado <= 0 {
		return 0, 0, 0, 0, errors.New("preço atual ou valor de mercado inválido para DCF")
	}

	// 1. Reconstituição de Balanço & Operacional
	numeroAcoes := ativo.ValorMercado / ativo.PrecoAtual
	if numeroAcoes <= 0 {
		return 0, 0, 0, 0, errors.New("número de ações inconsistente")
	}

	plContabil := 0.0
	if ativo.PVPReal > 0 {
		plContabil = ativo.ValorMercado / ativo.PVPReal
	} else if ativo.VPA > 0 {
		plContabil = numeroAcoes * ativo.VPA
	} else {
		return 0, 0, 0, 0, errors.New("patrimônio líquido contábil indisponível")
	}

	dividaLiquida := ativo.DividaLiquidaPL * plContabil
	enterpriseValue := ativo.ValorMercado + dividaLiquida
	if enterpriseValue <= 0 {
		enterpriseValue = ativo.ValorMercado
	}

	ebit := 0.0
	if ativo.PEbit > 0 {
		// (preco_atual / p_ebit) * numero_acoes == valor_mercado / p_ebit
		ebit = (ativo.PrecoAtual / ativo.PEbit) * numeroAcoes
	} else if ativo.EVEbit > 0 && enterpriseValue > 0 {
		ebit = enterpriseValue / ativo.EVEbit
	} else if ativo.LPA > 0 {
		// Fallback operacional
		ebit = (ativo.LPA * numeroAcoes) / (1.0 - 0.34)
	}

	if ebit <= 0 {
		return 0, 0, 0, 0, errors.New("EBIT não positivo ou indisponível")
	}

	nopat := ebit * (1.0 - 0.34) // Alíquota padrão IRPJ/CSLL de 34%

	// 2. Cálculo de WACC e Taxas
	selicDec := cenario.TaxaSelic / 100.0
	if selicDec <= 0 {
		selicDec = 0.1375
	}
	ntnbDec := cenario.TaxaNTNB / 100.0
	if ntnbDec <= 0 {
		ntnbDec = 0.0650
	}
	erpDec := cenario.EquityRiskPremium / 100.0
	if erpDec <= 0 {
		erpDec = 0.0550
	}

	// Kd: custo da dívida após benefício fiscal com spread bancário de 1.75% a.a.
	kd := (selicDec + 0.0175) * (1.0 - 0.34)
	// Ke: custo de capital próprio ancorado em NTN-B real + Equity Risk Premium Brasil
	ke := ntnbDec + erpDec

	we := ativo.ValorMercado / enterpriseValue
	wd := math.Max(0.0, dividaLiquida) / enterpriseValue
	somaPesos := we + wd
	if somaPesos > 0 {
		we /= somaPesos
		wd /= somaPesos
	} else {
		we = 1.0
		wd = 0.0
	}

	wacc := (we * ke) + (wd * kd)
	if wacc < 0.08 {
		wacc = 0.08 // Piso de segurança prudencial para custo de capital no Brasil
	}

	// 3. Reinvestimento e FCFF (Damodaran Proxy)
	rr := 0.0
	if ativo.ROIC > 0 {
		rr = ativo.CrescimentoLucro5A / ativo.ROIC
	}
	payoutDec := math.Min(1.0, math.Max(0.0, ativo.Payout/100.0))
	retencao := 1.0 - payoutDec

	// Trava de segurança: Se RR > 0.70 ou RR <= 0, adote RR = min(0.70, max(0.10, 1 - payout))
	if rr > 0.70 || rr <= 0.0 {
		rr = math.Min(0.70, math.Max(0.10, retencao))
	}

	fcff := nopat * (1.0 - rr)
	if fcff <= 0 {
		return 0, 0, 0, 0, errors.New("FCFF não positivo")
	}

	// 4. Perpetuidade & Preço Teto por Ação
	crescDec := math.Max(0.0, ativo.CrescimentoLucro5A/100.0)
	g := math.Min(crescDec*0.5, 0.0361) // Travada no teto real do PIB de 3,61% a.a.

	denominador := wacc - g
	if denominador < 0.02 {
		denominador = 0.02 // Evita singularidades matemáticas
	}

	evIntrinseco := (fcff * (1.0 + g)) / denominador
	equityIntrinseco := evIntrinseco - dividaLiquida
	if equityIntrinseco <= 0 {
		return 0, wacc, fcff, g, nil
	}

	precoTetoDCF := math.Max(0.0, equityIntrinseco/numeroAcoes)
	precoTetoDCF = math.Round(precoTetoDCF*100) / 100
	ativo.PrecoTetoDCF = precoTetoDCF

	return precoTetoDCF, wacc, fcff, g, nil
}

// AnalisadorDCF implementa a interface domain.Analisador
type AnalisadorDCF struct{}

func Novo() *AnalisadorDCF {
	return &AnalisadorDCF{}
}

func (d *AnalisadorDCF) Nome() string {
	return "Fluxo de Caixa Descontado (DCF Proxy)"
}

func (d *AnalisadorDCF) Executar(ativo *domain.Ativo) error {
	if ativo.Classe != domain.ClasseAcao {
		return nil
	}

	if ativo.Pareceres == nil {
		ativo.Pareceres = make(map[string]domain.ParecerItem)
	}

	if IsSetorFinanceiro(ativo.Setor, ativo.Subsetor, ativo.Segmento) {
		ativo.PrecoTetoDCF = 0
		ativo.Pareceres["DCF"] = domain.ParecerItem{
			Status:  domain.ParecerNaoAplica,
			Metrica: "Inaplicável p/ Financeiro",
			Detalhe: "Instituições financeiras são avaliadas por Bazin 5A, Gordon e Lynch",
		}
		return nil
	}

	cenario := macro.ObterCenario()
	tetoDCF, wacc, fcff, g, err := CalcularDCF(ativo, cenario)
	if err != nil || tetoDCF <= 0 {
		ativo.PrecoTetoDCF = 0
		ativo.Pareceres["DCF"] = domain.ParecerItem{
			Status:  domain.ParecerNaoAplica,
			Metrica: "DCF: Em revisão",
			Detalhe: "Métricas operacionais insuficientes ou FCFF inconsistente",
		}
		return nil
	}

	status := domain.ParecerReprovado
	if ativo.PrecoAtual <= tetoDCF {
		status = domain.ParecerAprovado
	}

	margem := 0.0
	if tetoDCF > 0 {
		margem = ((tetoDCF - ativo.PrecoAtual) / tetoDCF) * 100.0
	}

	ativo.Pareceres["DCF"] = domain.ParecerItem{
		Status:  status,
		Metrica: fmt.Sprintf("Teto DCF: R$ %6.2f | WACC: %4.1f%%", tetoDCF, wacc*100.0),
		Detalhe: fmt.Sprintf("FCFF: R$ %5.1fM | g: %3.1f%% (Margem: %+.1f%%)", fcff/1e6, g*100.0, margem),
	}

	return nil
}
