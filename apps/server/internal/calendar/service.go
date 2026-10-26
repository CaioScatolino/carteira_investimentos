package calendar

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"carteira_investimentos/server/internal/b3"
	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/storage"
)

var nomesMeses = []string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

type Service struct {
	store      storage.SnapshotStore
	b3Client   *b3.B3Client
	cacheMu    sync.RWMutex
	cacheDados map[string]AtivoCalendario // chave: ticker + "_" + tipoData
	cacheTTL   time.Time
}

func NovoService(store storage.SnapshotStore, b3Client *b3.B3Client) *Service {
	if b3Client == nil {
		b3Client = b3.NovoB3Client()
	}
	return &Service{
		store:      store,
		b3Client:   b3Client,
		cacheDados: make(map[string]AtivoCalendario),
	}
}

// ObterCalendario processa os proventos históricos dos ativos (top volume / carteira) e calcula a matriz sazonal
func (s *Service) ObterCalendario(ctx context.Context, filtro FiltroCalendario) (*ResumoCalendario, error) {
	todosAtivos := s.store.ListarTodos()
	if len(todosAtivos) == 0 {
		return &ResumoCalendario{
			TipoData:            filtro.TipoData,
			PeriodoAnos:         5,
			ProbabilidadeMinima: filtro.ProbabilidadeMinima,
			MesesMaisFrequentes: make(map[string]int),
			Ativos:              []AtivoCalendario{},
		}, nil
	}

	tipoData := strings.ToUpper(strings.TrimSpace(filtro.TipoData))
	if tipoData != "DATA_PAGAMENTO" {
		tipoData = "DATA_COM"
	}

	limite := filtro.Limite
	if limite <= 0 {
		limite = 20
	} else if limite > 100 {
		limite = 100
	}

	offset := filtro.Offset
	if offset < 0 {
		offset = 0
	}

	mapaTickersCarteira := make(map[string]bool)
	for _, tk := range filtro.TickersCarteira {
		mapaTickersCarteira[strings.ToUpper(strings.TrimSpace(tk))] = true
	}

	anoAtual := time.Now().Year()
	anosReferencia := []int{anoAtual - 5, anoAtual - 4, anoAtual - 3, anoAtual - 2, anoAtual - 1}

	// 1. Filtragem preliminar em memória (sem chamadas I/O de rede)
	var candidatos []*domain.Ativo
	for _, a := range todosAtivos {
		if a.PrecoAtual <= 0 {
			continue
		}

		tkUpper := strings.ToUpper(strings.TrimSpace(a.Ticker))

		// Filtro Carteira
		naCarteira := mapaTickersCarteira[tkUpper]
		if filtro.ApenasCarteira && !naCarteira {
			continue
		}

		// Filtro Classe
		if filtro.Classe != "" && !strings.EqualFold(string(a.Classe), filtro.Classe) {
			continue
		}

		// Filtro Setor
		if filtro.Setor != "" && !strings.EqualFold(a.Setor, filtro.Setor) {
			continue
		}

		// Filtro Busca (Ticker ou Nome)
		if filtro.Busca != "" {
			termo := strings.ToLower(strings.TrimSpace(filtro.Busca))
			match := strings.Contains(strings.ToLower(a.Ticker), termo) || strings.Contains(strings.ToLower(a.Nome), termo)
			if !match {
				continue
			}
		}

		candidatos = append(candidatos, a)
	}

	// 2. Ordenação Institucional por Volume / Liquidez
	// Ativos na carteira têm prioridade no topo; em seguida, ordenados por Volume Total e Score
	sort.Slice(candidatos, func(i, j int) bool {
		tkI := strings.ToUpper(strings.TrimSpace(candidatos[i].Ticker))
		tkJ := strings.ToUpper(strings.TrimSpace(candidatos[j].Ticker))
		posI := mapaTickersCarteira[tkI]
		posJ := mapaTickersCarteira[tkJ]
		if posI != posJ {
			return posI
		}
		if candidatos[i].VolumeTotal != candidatos[j].VolumeTotal {
			return candidatos[i].VolumeTotal > candidatos[j].VolumeTotal
		}
		return candidatos[i].Score > candidatos[j].Score
	})

	totalCandidatos := len(candidatos)

	// 3. Paginação do slice de candidatos (ex: top 20 volume)
	start := offset
	if start > totalCandidatos {
		start = totalCandidatos
	}
	end := start + limite
	if end > totalCandidatos {
		end = totalCandidatos
	}
	pagina := candidatos[start:end]
	temMais := end < totalCandidatos

	// 4. Execução Concorrente com Limitação de Concorrência (8 workers paralelos)
	type itemResult struct {
		index int
		item  AtivoCalendario
	}
	resCh := make(chan itemResult, len(pagina))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for i, ativo := range pagina {
		wg.Add(1)
		go func(idx int, a *domain.Ativo) {
			defer wg.Done()
			tkUpper := strings.ToUpper(strings.TrimSpace(a.Ticker))
			chaveCache := tkUpper + "_" + tipoData

			s.cacheMu.RLock()
			itemCal, emCache := s.cacheDados[chaveCache]
			s.cacheMu.RUnlock()

			if !emCache {
				sem <- struct{}{}
				itemCal = s.calcularSazonalidadeAtivo(a, tipoData, anosReferencia)
				<-sem

				s.cacheMu.Lock()
				s.cacheDados[chaveCache] = itemCal
				s.cacheMu.Unlock()
			}

			itemCal.PossuiNaCarteira = mapaTickersCarteira[tkUpper]
			resCh <- itemResult{index: idx, item: itemCal}
		}(i, ativo)
	}

	wg.Wait()
	close(resCh)

	itensCalculados := make([]AtivoCalendario, len(pagina))
	for r := range resCh {
		itensCalculados[r.index] = r.item
	}

	// 5. Aplicar filtro de probabilidade mínima e agregar estatísticas
	var resultado []AtivoCalendario
	mesesFrequentes := make(map[string]int)

	for _, itemCal := range itensCalculados {
		if filtro.ProbabilidadeMinima > 0 && itemCal.MaiorProbabilidade < filtro.ProbabilidadeMinima {
			continue
		}
		resultado = append(resultado, itemCal)

		for _, m := range itemCal.Meses {
			if m.Probabilidade >= 50.0 {
				mesesFrequentes[m.NomeMes]++
			}
		}
	}

	return &ResumoCalendario{
		TotalAtivos:         totalCandidatos,
		Limite:              limite,
		Offset:              offset,
		TemMais:             temMais,
		TipoData:            tipoData,
		PeriodoAnos:         len(anosReferencia),
		ProbabilidadeMinima: filtro.ProbabilidadeMinima,
		MesesMaisFrequentes: mesesFrequentes,
		Ativos:              resultado,
	}, nil
}

// calcularSazonalidadeAtivo analisa o histórico oficial de proventos da B3/CVM e calcula os 12 meses
func (s *Service) calcularSazonalidadeAtivo(a *domain.Ativo, tipoData string, anosRef []int) AtivoCalendario {
	tipoAlvo := "ON"
	tkUpper := strings.ToUpper(strings.TrimSpace(a.Ticker))
	if strings.HasSuffix(tkUpper, "4") || strings.HasSuffix(tkUpper, "5") || strings.HasSuffix(tkUpper, "6") {
		tipoAlvo = "PN"
	} else if strings.HasSuffix(tkUpper, "11") && a.Classe == domain.ClasseAcao {
		tipoAlvo = "UNT"
	}

	// Mapa auxiliar: mês (1 a 12) -> lista de anos únicos com provento
	anosPorMes := make(map[int]map[int]bool)
	valoresPorMes := make(map[int][]float64)
	tiposPorMes := make(map[int]map[string]bool)

	for m := 1; m <= 12; m++ {
		anosPorMes[m] = make(map[int]bool)
		tiposPorMes[m] = make(map[string]bool)
	}

	totalEventos := 0

	// CASO 1: AÇÕES (Base Oficial de Eventos em Dinheiro B3)
	if a.Classe == domain.ClasseAcao {
		eventos, _, _ := s.b3Client.ConsultarEventosBrutosB3(a.Nome, a.Ticker)
		totalEventos = len(eventos)

		for _, it := range eventos {
			// Filtra classe da ação (ON, PN, UNT)
			if it.TypeStock != "" && !strings.Contains(it.TypeStock, tipoAlvo) {
				continue
			}

			// Parse da data de referência
			dtStr := it.LastDateTimePriorEx
			if dtStr == "" {
				dtStr = it.LastDatePriorEx
			}
			if dtStr == "" {
				dtStr = it.DateApproval
			}

			var dt time.Time
			if strings.Contains(dtStr, "T") {
				dt, _ = time.Parse("2006-01-02T15:04:05", dtStr)
			} else if strings.Contains(dtStr, "/") && len(dtStr) >= 10 {
				dt, _ = time.Parse("02/01/2006", dtStr[:10])
			}

			if dt.IsZero() {
				continue
			}

			// Se a modalidade solicitada for Data Pagamento:
			// Na ausência da data exata no endpoint B3, projeta o mês de pagamento habitual (+1 mês após Data COM)
			if tipoData == "DATA_PAGAMENTO" {
				dt = dt.AddDate(0, 1, 0)
			}

			ano := dt.Year()
			mes := int(dt.Month())

			// Verifica se o ano está na janela de análise
			anoValido := false
			for _, ar := range anosRef {
				if ar == ano {
					anoValido = true
					break
				}
			}
			// Também aceita anos recentes caso o ativo tenha histórico um pouco mais curto
			if !anoValido && ano >= anosRef[0] {
				anoValido = true
			}

			if anoValido && mes >= 1 && mes <= 12 {
				anosPorMes[mes][ano] = true
				if it.CorporateAction != "" {
					tiposPorMes[mes][it.CorporateAction] = true
				}

				valStr := strings.ReplaceAll(strings.ReplaceAll(it.ValueCash, ".", ""), ",", ".")
				if val, err := strconv.ParseFloat(valStr, 64); err == nil && val > 0 {
					valoresPorMes[mes] = append(valoresPorMes[mes], val)
				}
			}
		}
	} else if a.Classe == domain.ClasseFII {
		// CASO 2: FUNDOS IMOBILIÁRIOS (FIIs via Histórico de Rendimentos)
		siResp, err := s.b3Client.ObterProventosStatusInvest(a.Ticker, true)
		if err == nil && siResp != nil {
			totalEventos = len(siResp.AssetEarningsModels)
			for _, it := range siResp.AssetEarningsModels {
				dataStr := it.Ed
				if tipoData == "DATA_PAGAMENTO" && it.Pd != "" {
					dataStr = it.Pd
				}

				dt, err := time.Parse("02/01/2006", dataStr)
				if err != nil {
					continue
				}

				ano := dt.Year()
				mes := int(dt.Month())

				anoValido := false
				for _, ar := range anosRef {
					if ar == ano {
						anoValido = true
						break
					}
				}
				if !anoValido && ano >= anosRef[0] {
					anoValido = true
				}

				if anoValido && mes >= 1 && mes <= 12 {
					anosPorMes[mes][ano] = true
					tiposPorMes[mes]["Rendimento"] = true
					if it.V > 0 {
						valoresPorMes[mes] = append(valoresPorMes[mes], it.V)
					}
				}
			}
		}
	}

	// Monta os 12 meses consolidados
	meses := make([]MesProvento, 12)
	totalMesesComProb := 0
	maiorProb := 0.0
	mesMaisProv := "—"
	totalAnosRef := len(anosRef)
	if totalAnosRef == 0 {
		totalAnosRef = 5
	}

	for i := 0; i < 12; i++ {
		numMes := i + 1
		anosSet := anosPorMes[numMes]

		var anosLista []int
		for ano := range anosSet {
			anosLista = append(anosLista, ano)
		}
		sort.Ints(anosLista)

		ocorrencias := len(anosLista)
		prob := (float64(ocorrencias) / float64(totalAnosRef)) * 100.0
		if prob > 100.0 {
			prob = 100.0
		}
		prob = math.Round(prob*10) / 10

		somaValores := 0.0
		for _, v := range valoresPorMes[numMes] {
			somaValores += v
		}
		valorMedio := 0.0
		if len(valoresPorMes[numMes]) > 0 {
			valorMedio = math.Round((somaValores/float64(len(valoresPorMes[numMes])))*100) / 100
		}

		var tiposLista []string
		for t := range tiposPorMes[numMes] {
			tiposLista = append(tiposLista, t)
		}
		sort.Strings(tiposLista)

		meses[i] = MesProvento{
			Mes:             numMes,
			NomeMes:         nomesMeses[i],
			Probabilidade:   prob,
			Ocorrencias:     ocorrencias,
			TotalAnos:       totalAnosRef,
			ValorMedio:      valorMedio,
			AnosComProvento: anosLista,
			Tipos:           tiposLista,
		}

		if prob > 0 {
			totalMesesComProb++
		}
		if prob > maiorProb {
			maiorProb = prob
			mesMaisProv = nomesMeses[i]
		}
	}

	return AtivoCalendario{
		Ticker:                a.Ticker,
		Nome:                  a.Nome,
		Classe:                string(a.Classe),
		Setor:                 a.Setor,
		Segmento:              a.Segmento,
		PrecoAtual:            a.PrecoAtual,
		DY:                    a.DY,
		Score:                 a.Score,
		Porte:                 a.Porte,
		Status:                string(a.Status),
		Meses:                 meses,
		TotalMesesPagadores:   totalMesesComProb,
		MesMaisProvavel:       mesMaisProv,
		MaiorProbabilidade:    maiorProb,
		TotalEventosHistorico: totalEventos,
	}
}
