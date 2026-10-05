package statusinvest

import (
	"context"
	"fmt"
	"math"
	"time"

	"carteira_investimentos/server/internal/b3"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/storage"
)

// IngestionService coordena a captura integral dos dados do StatusInvest e auditoria dos motores
type IngestionService struct {
	client   *Client
	b3Client *b3.B3Client
	store    storage.SnapshotStore
	motores  []domain.Analisador
}

// NovoIngestionService cria uma nova instância do serviço de ingestão consolidado
func NovoIngestionService(client *Client, b3Client *b3.B3Client, store storage.SnapshotStore, motores []domain.Analisador) *IngestionService {
	if client == nil {
		client = NovoClient()
	}
	if b3Client == nil {
		b3Client = b3.NovoB3Client()
	}
	return &IngestionService{
		client:   client,
		b3Client: b3Client,
		store:    store,
		motores:  motores,
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

		payout := 0.0
		if a.LPA > 0 && divs12M > 0 {
			payout = (divs12M / a.LPA) * 100.0
		} else if a.LPA <= 0 && divs12M > 0 {
			payout = 999.0 // Payout com prejuízo
		}

		isAtipico := false
		alertaRisco := ""

		// Critérios Institucionais de Yield Trap em Ações:
		if a.DY >= 18.0 {
			isAtipico = true
			alertaRisco = fmt.Sprintf("Yield Atípico (%4.1f%%): Provento extraordinário não perpétuo", a.DY)
		} else if payout > 115.0 && a.DY >= 8.0 {
			isAtipico = true
			alertaRisco = fmt.Sprintf("Payout Excessivo (%4.0f%%): Distribuição acima do lucro anual", payout)
		} else if a.LPA <= 0 && divs12M > 0 {
			isAtipico = true
			alertaRisco = "Prejuízo Operacional: Provento pago com reservas ou desinvestimento"
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
			Payout:              payout,
			IsProventoAtipico:   isAtipico,
			AlertaRisco:         alertaRisco,
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

		// Décio Bazin: Preço Teto e Média Histórica de 5 Anos
		media5A := divs12M
		if a.LucrosCAGR5 > 0 && a.LucrosCAGR5 < 100 {
			fator := 1.0 / (1.0 + (a.LucrosCAGR5/100.0)*0.4)
			media5A = divs12M * fator
		}
		ativo.MediaDividendos5A = math.Round(media5A*100) / 100
		if media5A > 0 {
			ativo.PrecoTetoBazin5A = math.Round((media5A/0.06)*100) / 100
			if ativo.PrecoAtual > 0 {
				ativo.MargemBazin5A = math.Round(((ativo.PrecoTetoBazin5A-ativo.PrecoAtual)/ativo.PrecoTetoBazin5A)*1000) / 10
			}
		}

		// Se o ticker for Blue Chip / pagadora mapeada na B3, enriquece com a série oficial da B3
		if s.b3Client != nil && b3.IsTickerMapeado(ativo.Ticker) {
			analise := s.b3Client.ObterAnaliseProventos5A(a.CompanyName, a.Ticker, a.Price, divs12M)
			if analise != nil && analise.TotalEventos > 0 {
				ativo.MediaDividendos5A = analise.MediaDividendos5A
				ativo.MediaDividendos5ANormalizada = analise.MediaDividendos5ANormalizada
				ativo.PrecoTetoBazin5A = analise.PrecoTetoBazin5A
				ativo.MargemBazin5A = analise.MargemBazin5A
				ativo.Dividendos12MB3 = analise.Dividendos12MB3Liquido
				ativo.HistoricoDividendosAnual = analise.ProventosPorAno
				ativo.Diferenca12MB3StatusInvest = analise.Diferenca12M
				ativo.AderenciaStatusInvest = analise.AderenciaStatusInvest
				ativo.TeveOutlier5A = analise.TeveOutlier5A
				ativo.ObservacaoOutlier = analise.ObservacaoOutlier
			}
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

		isAtipicoFII := false
		alertaRiscoFII := ""
		if f.DY >= 18.0 {
			isAtipicoFII = true
			alertaRiscoFII = fmt.Sprintf("Amortização Extraordinária (DY %4.1f%%): Fundo em liquidação ou devolução de capital", f.DY)
		} else if f.PVP > 0 && f.PVP < 0.35 && divs12M > 0 {
			isAtipicoFII = true
			alertaRiscoFII = "Risco Crítico: Cotação em colapso / Possível liquidação judicial"
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
			IsProventoAtipico:   isAtipicoFII,
			AlertaRisco:         alertaRiscoFII,
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

		// Décio Bazin FII: Média Histórica de 5 Anos (estimativa prudente padrão)
		media5AFII := divs12M
		if f.DividendCAGR > 0 && f.DividendCAGR < 100 {
			fator := 1.0 / (1.0 + (f.DividendCAGR/100.0)*0.4)
			media5AFII = divs12M * fator
		}
		ativo.MediaDividendos5A = math.Round(media5AFII*100) / 100
		if media5AFII > 0 {
			ativo.PrecoTetoBazin5A = math.Round((media5AFII/0.06)*100) / 100
			if ativo.PrecoAtual > 0 {
				ativo.MargemBazin5A = math.Round(((ativo.PrecoTetoBazin5A-ativo.PrecoAtual)/ativo.PrecoTetoBazin5A)*1000) / 10
			}
		}

		// Se o FII for consolidado/relevante, enriquece com a série histórica auditada CVM/StatusInvest
		if s.b3Client != nil && (f.NumeroCotistas >= 50000 || f.LiquidezMediaDiaria >= 3000000) {
			analise := s.b3Client.ObterAnaliseProventos5A(f.CompanyName, f.Ticker, f.Price, divs12M)
			if analise != nil && analise.TotalEventos > 0 {
				ativo.MediaDividendos5A = analise.MediaDividendos5A
				ativo.MediaDividendos5ANormalizada = analise.MediaDividendos5ANormalizada
				ativo.PrecoTetoBazin5A = analise.PrecoTetoBazin5A
				ativo.MargemBazin5A = analise.MargemBazin5A
				ativo.Dividendos12MB3 = analise.Dividendos12MB3Liquido
				ativo.HistoricoDividendosAnual = analise.ProventosPorAno
				ativo.Diferenca12MB3StatusInvest = analise.Diferenca12M
				ativo.AderenciaStatusInvest = analise.AderenciaStatusInvest
				ativo.TeveOutlier5A = analise.TeveOutlier5A
				ativo.ObservacaoOutlier = analise.ObservacaoOutlier
			}
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
