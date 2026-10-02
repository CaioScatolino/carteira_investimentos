package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"carteira_investimentos/server/internal/b3"
	"carteira_investimentos/server/internal/catalog"
	"carteira_investimentos/server/internal/domain"
)

// B3MarketProvider implementa a interface provider.MarketProvider
// utilizando exclusivamente dados oficiais do arquivo diário da B3 (COTAHIST_D)
// e o catálogo cadastral do MySQL local.
//
// VANTAGENS:
// 1. Elimina bloqueios de IP (HTTP 429 / 999) do Yahoo Finance em varreduras em lote.
// 2. Um único download HTTP diário (~450 KB comprimido) contém 100% dos ativos negociados.
// 3. Fornece o volume financeiro oficial real para filtrar ativos sem liquidez ("micos").
type B3MarketProvider struct {
	client        *b3.B3Client
	repo          catalog.Repository // Repositório MySQL para mapear Ticker -> CNPJ e Classe
	minVolume     float64            // Filtro de corte de liquidez em R$ (ex: 500_000.00)
	mu            sync.RWMutex       // Mutex para leitura e escrita segura no cache em memória
	cacheCotacoes map[string]*b3.CotacaoDiaria
	dataPregao    string
	ultimaCarga   time.Time
}

// NovoB3MarketProvider inicializa o provedor oficial com o cliente B3 e o repositório MySQL.
// minVolume define o corte financeiro diário para descartar ativos sem liquidez (ex: 500000.00 = R$ 500k/dia).
func NovoB3MarketProvider(client *b3.B3Client, repo catalog.Repository, minVolume float64) *B3MarketProvider {
	if client == nil {
		client = b3.NovoB3Client()
	}
	return &B3MarketProvider{
		client:    client,
		repo:      repo,
		minVolume: minVolume,
	}
}

// garantirCotacoesCarregadas verifica se as cotações do pregão mais recente já estão em memória.
// Se não estiverem (ou se expiraram há mais de 12 horas), faz o download oficial da B3 e o parse.
func (p *B3MarketProvider) garantirCotacoesCarregadas() error {
	p.mu.RLock()
	carregado := p.cacheCotacoes != nil && time.Since(p.ultimaCarga) < 12*time.Hour
	p.mu.RUnlock()

	if carregado {
		return nil
	}

	// Lock exclusivo para escrita no cache
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check após adquirir o lock de escrita
	if p.cacheCotacoes != nil && time.Since(p.ultimaCarga) < 12*time.Hour {
		return nil
	}

	// 1. Baixa o arquivo ZIP do pregão mais recente da B3
	zipBytes, dataFormatada, err := p.client.ObterArquivoMaisRecente()
	if err != nil {
		return fmt.Errorf("falha ao baixar COTAHIST da B3: %w", err)
	}

	// 2. Faz o parse em streaming com bufio.Scanner (leva < 15ms)
	cotacoes, err := b3.ParseCOTAHIST(zipBytes)
	if err != nil {
		return fmt.Errorf("falha ao processar arquivo posicional da B3: %w", err)
	}

	p.cacheCotacoes = cotacoes
	p.dataPregao = dataFormatada
	p.ultimaCarga = time.Now()

	return nil
}

// carregarMapaCatalogo carrega em lote o mapa Ticker -> Asset do MySQL
// para cruzar rapidamente a classe e o CNPJ de cada ticker sem fazer N queries individuais.
func (p *B3MarketProvider) carregarMapaCatalogo() map[string]*catalog.Asset {
	mapa := make(map[string]*catalog.Asset)
	if p.repo == nil {
		return mapa
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ativos, err := p.repo.ListarTodos(ctx)
	if err != nil {
		fmt.Printf("⚠️ [B3Provider Warning] Falha ao carregar catálogo MySQL: %s\n", err)
		return mapa
	}

	for _, a := range ativos {
		mapa[a.Ticker] = a
	}
	return mapa
}

// converterParaDominio transforma o registro bruto do COTAHIST em uma entidade domain.Ativo,
// enriquecendo com os dados cadastrais (CNPJ e Classe) vindos do catálogo MySQL.
func (p *B3MarketProvider) converterParaDominio(c *b3.CotacaoDiaria, cat *catalog.Asset) *domain.Ativo {
	classe := domain.ClasseAcao
	cnpj := ""
	dividendos := 0.0 // Valor padrão caso não tenha no catálogo

	if cat != nil {
		if cat.Classe == "FII" {
			classe = domain.ClasseFII
		}
		cnpj = cat.CNPJ
		dividendos = cat.Dividendos12M // <-- PUXA OS PROVENTOS DO MYSQL!
	} else {
		if c.CodigoBDI == "12" || strings.HasSuffix(c.Ticker, "11") {
			unitsAcoes := map[string]bool{
				"TAEE11": true, "SAPR11": true, "KLBN11": true,
				"ALUP11": true, "BPAC11": true, "SANB11": true,
			}
			if !unitsAcoes[c.Ticker] {
				classe = domain.ClasseFII
			}
		}
	}

	return &domain.Ativo{
		Ticker:        c.Ticker,
		Nome:          c.NomeResumido, // <-- Nome da empresa do pregão da B3
		Classe:        classe,
		CNPJ:          cnpj,
		VolumeTotal:   c.VolumeTotal,
		PrecoAtual:    c.PrecoFechamento, // Preço oficial PREULT do pregão
		Dividendos12M: dividendos,        // Proventos oficiais do catálogo local!
	}

}

// BuscarAtivo consulta a cotação oficial do ativo no pregão da B3
func (p *B3MarketProvider) BuscarAtivo(ticker string) (*domain.Ativo, error) {
	if err := p.garantirCotacoesCarregadas(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	c, existe := p.cacheCotacoes[ticker]
	p.mu.RUnlock()

	if !existe {
		return nil, fmt.Errorf("ativo %s não foi negociado no último pregão da B3", ticker)
	}

	var cat *catalog.Asset
	if p.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cat, _ = p.repo.BuscarPorTicker(ctx, ticker)
	}

	return p.converterParaDominio(c, cat), nil
}

// BuscarEmLote obtém as cotações em memória de forma instantânea (< 1ms).
// COMPORTAMENTO INTELIGENTE:
//   - Se 'tickers' contiver uma lista específica, retorna apenas os ativos dessa lista.
//   - Se 'tickers' for vazio (nil ou len == 0), retorna TODOS os ativos negociados que
//     possuem volume diário >= minVolume (Modo Scanner de Mercado Geral).
func (p *B3MarketProvider) BuscarEmLote(tickers []string) []*domain.Ativo {
	if err := p.garantirCotacoesCarregadas(); err != nil {
		fmt.Printf("⚠️ [B3Provider Error] %s\n", err)
		return nil
	}

	mapaCatalogo := p.carregarMapaCatalogo()

	p.mu.RLock()
	defer p.mu.RUnlock()

	var resultado []*domain.Ativo

	// Caso 1: Lista restrita solicitada explicitamente
	if len(tickers) > 0 {
		for _, t := range tickers {
			if c, existe := p.cacheCotacoes[t]; existe {
				resultado = append(resultado, p.converterParaDominio(c, mapaCatalogo[t]))
			}
		}
		return resultado
	}

	// Caso 2: Scanner Aberto - Filtra por liquidez mínima (VolumeTotal >= minVolume)
	for ticker, c := range p.cacheCotacoes {
		// Ignora ativos com liquidez insignificante para evitar distorções de preço
		if p.minVolume > 0 && c.VolumeTotal < p.minVolume {
			continue
		}
		resultado = append(resultado, p.converterParaDominio(c, mapaCatalogo[ticker]))
	}

	return resultado
}
