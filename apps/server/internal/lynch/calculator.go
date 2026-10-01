package lynch

import (
	"carteira_investimentos/server/internal/domain"
	"errors"
)

// AnalisadorLynch implementa a interface domain.Analisador
type AnalisadorLynch struct{}

func Novo() *AnalisadorLynch {
	return &AnalisadorLynch{}
}

func (l *AnalisadorLynch) Nome() string {
	return "Peter Lynch (PEG Ratio)"
}

func (l *AnalisadorLynch) Executar(ativo *domain.Ativo) error {
	// Aplicável apenas a Ações com dados de lucro e crescimento
	if ativo.Classe != domain.ClasseAcao || ativo.LPA <= 0 || ativo.CrescimentoLucro5A <= 0 {
		return errors.New("dados insuficientes para análise de Peter Lynch")
	}

	pl := ativo.PrecoAtual / ativo.LPA
	ativo.PEGRatio = pl / ativo.CrescimentoLucro5A

	// Preço Justo de Lynch: LPA * Taxa de Crescimento (limitada a 25% para conservadorismo)
	crescimento := ativo.CrescimentoLucro5A
	if crescimento > 25.0 {
		crescimento = 25.0
	}
	ativo.PrecoJustoLynch = ativo.LPA * crescimento

	return nil
}
