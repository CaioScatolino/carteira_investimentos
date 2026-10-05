package fii

import (
	"fmt"
	"math"
	"strings"

	"carteira_investimentos/server/internal/domain"
)

type AnalisadorFII struct {
	TaxaNTNB float64 // Exemplo: 0.065 para Tesouro IPCA+ de 6.5% a.a.
}

func Novo(taxaNTNB float64) *AnalisadorFII {
	return &AnalisadorFII{
		TaxaNTNB: taxaNTNB,
	}
}

func (f *AnalisadorFII) Nome() string {
	return "Valuation FII (P/VP, Cap Rate e Spread NTN-B)"
}

func (f *AnalisadorFII) Executar(ativo *domain.Ativo) error {
	if ativo.Classe != domain.ClasseFII || ativo.PrecoAtual <= 0 {
		return nil
	}

	if ativo.Pareceres == nil {
		ativo.Pareceres = make(map[string]domain.ParecerItem)
	}

	// 1. Cálculo do P/VP e Cap Rate Implícito
	if ativo.VPCota > 0 {
		ativo.PVP = ativo.PrecoAtual / ativo.VPCota
	}
	if ativo.PVP > 0 && ativo.DY > 0 {
		ativo.CapRateImplicito = math.Round((ativo.DY/ativo.PVP)*100) / 100
	}

	// 2. Cálculo do Spread sobre o Tesouro IPCA+
	if ativo.PrecoAtual > 0 {
		dividendYield := ativo.Dividendos12M / ativo.PrecoAtual
		ativo.SpreadNTNB = (dividendYield - f.TaxaNTNB) * 100.0 // em pontos percentuais
	}

	// 3. Classificação de Tipologia: Papel vs Tijolo vs FOF
	segmentoUpper := strings.ToUpper(ativo.Segmento)
	setorUpper := strings.ToUpper(ativo.Setor)
	isPapel := strings.Contains(segmentoUpper, "PAP") ||
		strings.Contains(segmentoUpper, "CRI") ||
		strings.Contains(segmentoUpper, "RECEB") ||
		strings.Contains(setorUpper, "PAP")

	isTijolo := strings.Contains(segmentoUpper, "SHOPPING") ||
		strings.Contains(segmentoUpper, "LOGÍST") ||
		strings.Contains(segmentoUpper, "LOGIST") ||
		strings.Contains(segmentoUpper, "LAJES") ||
		strings.Contains(segmentoUpper, "IMÓVEIS") ||
		strings.Contains(segmentoUpper, "IMOVEIS") ||
		strings.Contains(segmentoUpper, "HOTEL") ||
		strings.Contains(segmentoUpper, "HOSPITAL") ||
		strings.Contains(setorUpper, "TIJOLO")

	// Parecer 1: Segmentação & Preço Justo Patrimonial
	if isPapel {
		// Regra de Ouro para FIIs de Papel: PROIBIDO comprar com ágio (> 1.02x)
		if ativo.PVP > 1.02 {
			ativo.AlertaRisco = "Ágio em FII de Papel"
			ativo.VereditoRisco = "RISCO_DESCOMPENSADO"
			ativo.JustificativaRisco = fmt.Sprintf("🔴 Ágio prejudicial em FII de Papel (P/VP %.2fx). Comprar crédito acima do VP corrói o retorno.", ativo.PVP)
			ativo.Pareceres["FII Tipo & Ágio"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("P/VP: %4.2fx (Papel/CRI)", ativo.PVP),
				Detalhe: "Alerta de Ágio: Comprar CRI acima do VP corrói o retorno",
			}
		} else if ativo.PVP >= 0.85 && ativo.PVP <= 1.01 {
			ativo.Pareceres["FII Tipo & Ágio"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("P/VP: %4.2fx | Spread: %+.1f%%", ativo.PVP, ativo.SpreadNTNB),
				Detalhe: "Papel com preço justo/desconto seguro e sem ágio",
			}
		} else {
			ativo.Pareceres["FII Tipo & Ágio"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("P/VP: %4.2fx (Papel)", ativo.PVP),
				Detalhe: "Desconto acentuado: Verificar risco de inadimplência/calote",
			}
		}
	} else if isTijolo {
		// Regra de Ouro para FIIs de Tijolo: Desconto patrimonial + Cap Rate Implícito
		capRate := 0.0
		if ativo.PVP > 0 {
			capRate = ativo.DY / ativo.PVP
		}

		if ativo.PVP > 0 && ativo.PVP <= 0.95 && ativo.DY >= 8.0 {
			ativo.Pareceres["FII Tijolo & Cap Rate"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("P/VP: %4.2fx | Cap Rate: %4.1f%%", ativo.PVP, capRate),
				Detalhe: "Tijolo com desconto real sobre custo de reposição",
			}
		} else if ativo.PVP > 0 && ativo.PVP <= 1.02 {
			ativo.Pareceres["FII Tijolo & Cap Rate"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("P/VP: %4.2fx | Cap Rate: %4.1f%%", ativo.PVP, capRate),
				Detalhe: "Tijolo negociando próximo do valor patrimonial contábil",
			}
		} else {
			ativo.Pareceres["FII Tijolo & Cap Rate"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("P/VP: %4.2fx", ativo.PVP),
				Detalhe: "Tijolo negociando acima do valor patrimonial",
			}
		}
	} else {
		// FIIs Híbridos / FOFs / Outros
		if ativo.PVP > 0 && ativo.PVP <= 0.95 {
			ativo.Pareceres["FII Híbrido/FOF"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("P/VP: %4.2fx", ativo.PVP),
				Detalhe: "Desconto patrimonial atraente",
			}
		}
	}

	// Parecer 2: Estabilidade e Recorrência de Dividendos
	if ativo.IsProventoAtipico || ativo.DY >= 18.0 || (ativo.PVP > 0 && ativo.PVP < 0.35) {
		ativo.Pareceres["FII Proventos"] = domain.ParecerItem{
			Status:  domain.ParecerReprovado,
			Metrica: fmt.Sprintf("DY 12M: %4.1f%% (Atípico)", ativo.DY),
			Detalhe: "⚠️ Amortização Extraordinária: Fundo em devolução de capital ou liquidação",
		}
	} else if ativo.Dividendos12M > 0 && ativo.LastDividend > 0 && ativo.PrecoAtual > 0 {
		mediaMensal := ativo.Dividendos12M / 12.0
		razaoUltimo := ativo.LastDividend / mediaMensal

		if razaoUltimo < 0.70 {
			ativo.Pareceres["FII Proventos"] = domain.ParecerItem{
				Status:  domain.ParecerReprovado,
				Metrica: fmt.Sprintf("Último: R$ %4.2f (Média: R$ %4.2f)", ativo.LastDividend, mediaMensal),
				Detalhe: "Alerta: Último dividendo bem abaixo da média histórica dos 12M",
			}
		} else if razaoUltimo >= 0.90 && razaoUltimo <= 1.25 {
			ativo.Pareceres["FII Proventos"] = domain.ParecerItem{
				Status:  domain.ParecerAprovado,
				Metrica: fmt.Sprintf("Último: R$ %4.2f | Média: R$ %4.2f", ativo.LastDividend, mediaMensal),
				Detalhe: "Proventos regulares e estáveis",
			}
		} else {
			ativo.Pareceres["FII Proventos"] = domain.ParecerItem{
				Status:  domain.ParecerAtencao,
				Metrica: fmt.Sprintf("Último: R$ %4.2f | Média: R$ %4.2f", ativo.LastDividend, mediaMensal),
				Detalhe: "Provento volátil ou atípico em relação à média",
			}
		}
	} else if ativo.Dividendos12M <= 0 {
		ativo.Pareceres["FII Proventos"] = domain.ParecerItem{
			Status:  domain.ParecerReprovado,
			Metrica: "DY 12M: 0,0%",
			Detalhe: "Fundo sem histórico de proventos nos últimos 12 meses",
		}
	}

	return nil
}
