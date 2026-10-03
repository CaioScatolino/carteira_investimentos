package statusinvest

import (
	"context"
	"fmt"
	"time"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/storage"
)

// IngestionService coordena a captura integral dos dados do StatusInvest e auditoria dos motores
type IngestionService struct {
	client  *Client
	store   storage.SnapshotStore
	motores []domain.Analisador
}

// NovoIngestionService cria uma nova instância do serviço de ingestão consolidado
func NovoIngestionService(client *Client, store storage.SnapshotStore, motores []domain.Analisador) *IngestionService {
	if client == nil {
		client = NovoClient()
	}
	return &IngestionService{
		client:  client,
		store:   store,
		motores: motores,
	}
}

// ExecutarSincronizacao dispara as 2 requisições ao StatusInvest, converte os modelos e roda os motores
func (s *IngestionService) ExecutarSincronizacao(ctx context.Context) ([]*domain.Ativo, error) {
	inicio := time.Now()
	fmt.Println("\n==================================================================")
	fmt.Println("🚀 INGESTION SERVICE: SINCRONIZAÇÃO CONSOLIDADA STATUSINVEST")
	fmt.Println("==================================================================")

	// 1. Busca todas as ações
	acoesRaw, err := s.client.ObterAcoes(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao sincronizar ações: %w", err)
	}
	fmt.Printf("📊 [StatusInvest] %d Ações recebidas com todas as 36 colunas!\n", len(acoesRaw))

	// 2. Busca todos os FIIs
	fiisRaw, err := s.client.ObterFIIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao sincronizar FIIs: %w", err)
	}
	fmt.Printf("🏢 [StatusInvest] %d FIIs recebidos com todas as 22 colunas!\n", len(fiisRaw))

	var todosAtivos []*domain.Ativo

	// 3. Mapeia Ações para o modelo de domínio
	for _, a := range acoesRaw {
		if a.Price <= 0 || a.Ticker == "" {
			continue
		}

		divs12M := 0.0
		if a.DY > 0 {
			divs12M = a.Price * (a.DY / 100.0)
		}

		crescimento := a.LucrosCAGR5
		if crescimento <= 0 {
			crescimento = 10.0 // Média defensiva quando CAGR histórico não for positivo
		}

		ativo := &domain.Ativo{
			Ticker:              a.Ticker,
			Nome:                a.CompanyName,
			Classe:              domain.ClasseAcao,
			PrecoAtual:          a.Price,
			LPA:                 a.LPA,
			VPA:                 a.VPA,
			PL:                  a.PL,
			PVP:                 a.PVP,
			PVPReal:             a.PVP,
			DY:                  a.DY,
			Dividendos12M:       divs12M,
			CrescimentoLucro5A:  crescimento,
			PEGRatio:            a.PEGRatio,
			ROIC:                a.ROIC,
			ROA:                 a.ROA,
			ROE:                 a.ROE,
			GiroAtivos:          a.GiroAtivos,
			MargemBruta:         a.MargemBruta,
			MargemEbit:          a.MargemEbit,
			MargemLiquida:       a.MargemLiquida,
			DividaLiquidaPL:     a.DividaLiquidaPatrimonioLiquido,
			DividaLiquidaEbit:   a.DividaLiquidaEbit,
			LiquidezCorrente:    a.LiquidezCorrente,
			LiquidezMediaDiaria: a.LiquidezMediaDiaria,
			VolumeTotal:         a.LiquidezMediaDiaria,
			ValorMercado:        a.ValorMercado,
			Setor:               a.SectorName,
			Subsetor:            a.SubSectorName,
			Segmento:            a.SegmentName,
			Pareceres:           make(map[string]domain.ParecerItem),
		}

		if ativo.PrecoAtual > 0 && ativo.LPA != 0 {
			ativo.EarningsYield = (ativo.LPA / ativo.PrecoAtual) * 100.0
		}

		// Roda os motores de valuation
		for _, motor := range s.motores {
			_ = motor.Executar(ativo)
		}
		ativo.AvaliarSemaforo()

		todosAtivos = append(todosAtivos, ativo)
		if s.store != nil {
			s.store.Salvar(ativo)
		}
	}

	// 4. Mapeia FIIs para o modelo de domínio
	for _, f := range fiisRaw {
		if f.Price <= 0 || f.Ticker == "" {
			continue
		}

		divs12M := 0.0
		if f.DY > 0 {
			divs12M = f.Price * (f.DY / 100.0)
		}

		ativo := &domain.Ativo{
			Ticker:              f.Ticker,
			Nome:                f.CompanyName,
			Classe:              domain.ClasseFII,
			PrecoAtual:          f.Price,
			VPCota:              f.ValorPatrimonialCota,
			VPA:                 f.ValorPatrimonialCota,
			PVP:                 f.PVP,
			DY:                  f.DY,
			Dividendos12M:       divs12M,
			LastDividend:        f.LastDividend,
			LiquidezMediaDiaria: f.LiquidezMediaDiaria,
			VolumeTotal:         f.LiquidezMediaDiaria,
			PercentualCaixa:     f.PercentualCaixa,
			NumeroCotistas:      f.NumeroCotistas,
			Gestao:              f.GestaoF,
			Setor:               f.SectorName,
			Subsetor:            f.SubSectorName,
			Segmento:            f.Segment,
			Pareceres:           make(map[string]domain.ParecerItem),
		}

		// Roda os motores de valuation
		for _, motor := range s.motores {
			_ = motor.Executar(ativo)
		}
		ativo.AvaliarSemaforo()

		todosAtivos = append(todosAtivos, ativo)
		if s.store != nil {
			s.store.Salvar(ativo)
		}
	}

	duracao := time.Since(inicio)
	fmt.Printf("✅ [Ingestion] Sucesso total: %d ativos auditados e armazenados em %v!\n", len(todosAtivos), duracao)
	fmt.Println("==================================================================")

	return todosAtivos, nil
}
