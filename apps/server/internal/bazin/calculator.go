package bazin

import (
	"errors"
)

// Calcular aplica a fórmula de Décio Bazin: Preço Teto = Proventos / Yield Mínimo

func Calcular(dividendos12M float64, yieldMinimo float64) (float64, error) {
	if yieldMinimo <= 0 {
		return 0, errors.New("o yield mínimo deve ser maior que zero (ex: 0.06 para 6%)")
	}
	precoTeto := dividendos12M / yieldMinimo
	return precoTeto, nil
}

// CalcularMargem calcula a margem percentual em relação à cotação atual
func CalcularMargem(precoAtual float64, precoTeto float64) float64 {
	if precoAtual <= 0 {
		return 0
	}
	return ((precoTeto - precoAtual) / precoAtual) * 100.0
}
