package statusinvest

import (
	"context"
	"testing"
	"time"
)

func TestClientObterAcoes(t *testing.T) {
	c := NovoClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	acoes, err := c.ObterAcoes(ctx)
	if err != nil {
		t.Fatalf("Erro ao obter acoes: %v", err)
	}

	if len(acoes) == 0 {
		t.Fatalf("Nenhuma acao retornada")
	}

	for _, a := range acoes {
		if a.Ticker == "PETR4" || a.Ticker == "VALE3" || a.Ticker == "BBAS3" {
			div12M := a.Price * (a.DY / 100.0)
			t.Logf("[%s] Preco: %.2f | DY: %.2f%% | Div12M: R$ %.2f | Nome: %s", a.Ticker, a.Price, a.DY, div12M, a.CompanyName)
		}
	}
}
