package fii

import (
	"carteira_investimentos/server/internal/domain"
	"errors"
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
	return "Valuation FII (P/VP e Spread NTN-B)"
}

func (f *AnalisadorFII) Executar(ativo *domain.Ativo) error {
	if ativo.Classe != domain.ClasseFII || ativo.PrecoAtual <= 0 {
		return errors.New("método aplicável apenas a Fundos Imobiliários")
	}

	// 1. Cálculo do P/VP
	if ativo.VPCota > 0 {
		ativo.PVP = ativo.PrecoAtual / ativo.VPCota
	}

	// 2. Cálculo do Spread sobre o Tesouro IPCA+
	dividendYield := ativo.Dividendos12M / ativo.PrecoAtual
	ativo.SpreadNTNB = (dividendYield - f.TaxaNTNB) * 100.0 // em pontos percentuais

	return nil
}
