package graham

import (
	"carteira_investimentos/server/internal/domain"
	"errors"
	"math" // Biblioteca matemática padrão do Go (para raiz quadrada math.Sqrt)
)

const MultiplicadorGraham = 22.5

// Calcular determina o Preço Justo / Valor Intrínseco pelo modelo de Benjamin Graham
func Calcular(lpa float64, vpa float64) (float64, error) {
	// Graham exige lucros e patrimônio positivos
	if lpa <= 0 || vpa <= 0 {
		return 0, errors.New("LPA e VPA devem ser estritamente positivos para o método de Graham")
	}
	valorIntrinseco := math.Sqrt(MultiplicadorGraham * lpa * vpa)
	return valorIntrinseco, nil
}

// AnalisadorGraham implementa domain.Analisador
type AnalisadorGraham struct{}

func Novo() *AnalisadorGraham {
	return &AnalisadorGraham{}
}

func (g *AnalisadorGraham) Nome() string {
	return "Benjamin Graham"
}

func (g *AnalisadorGraham) Executar(ativo *domain.Ativo) error {
	if ativo.Classe != domain.ClasseAcao {
		return nil
	}
	vi, err := Calcular(ativo.LPA, ativo.VPA)
	if err != nil {
		return err
	}
	ativo.ValorGraham = vi
	return nil
}
