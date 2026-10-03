package provider

import "carteira_investimentos/server/internal/domain"

type MarketProvider interface {
	BuscarAtivo(ticker string) (*domain.Ativo, error)
	BuscarEmLote(tickers []string) []*domain.Ativo
}
