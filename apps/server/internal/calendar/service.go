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

// ObterCalendario processa os proventos históricos de todos os ativos ou da carteira e calcula a matriz sazonal
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

	mapaTickersCarteira := make(map[string]bool)
	for _, tk := range filtro.TickersCarteira {
		mapaTickersCarteira[strings.ToUpper(strings.TrimSpace(tk))] = true
	}

	anoAtual := time.Now().Year()
	anosReferencia := []int{anoAtual - 5, anoAtual - 4, anoAtual - 3, anoAtual - 2, anoAtual - 1}

	var resultado []AtivoCalendario
	mesesFrequentes := make(map[string]int)

	// Processa os ativos da base
	for _, a := range todosAtivos {
		if a.PrecoAtual <= 0 {
			continue
		}

		tkUpper := strings.ToUpper(strings.TrimSpace(a.Ticker))

		// 1. Filtro: Apenas Carteira
		naCarteira := mapaTickersCarteira[tkUpper]
		if filtro.ApenasCarteira && !naCarteira {
			continue
		}

		// 2. Filtro: Classe
		if filtro.Classe != "" && !strings.EqualFold(string(a.Classe), filtro.Classe) {
			continue
		}

		// 3. Filtro: Setor
		if filtro.Setor != "" && !strings.EqualFold(a.Setor, filtro.Setor) {
			continue
		}

		// 4. Filtro: Busca por texto (Ticker ou Nome)
		if filtro.Busca != "" {
			termo := strings.ToLower(strings.TrimSpace(filtro.Busca))
			match := strings.Contains(strings.ToLower(a.Ticker), termo) || strings.Contains(strings.ToLower(a.Nome), termo)
			if !match {
				continue
			}
		}

		// Consulta ou recupera do cache o cálculo deste ativo
		chaveCache := tkUpper + "_" + tipoData
		s.cacheMu.RLock()
		itemCal, emCache := s.cacheDados[chaveCache]
		s.cacheMu.RUnlock()

		if !emCache {
			itemCal = s.calcularSazonalidadeAtivo(a, tipoData, anosReferencia)
			s.cacheMu.Lock()
			s.cacheDados[chaveCache] = itemCal
			s.cacheMu.Unlock()
		}

		itemCal.PossuiNaCarteira = naCarteira

		// 5. Filtro: Probabilidade Mínima
		// Se o filtro exigir prob mínima (ex: >= 50%), o ativo precisa ter pelo menos um mês com prob >= filtro
		if filtro.ProbabilidadeMinima > 0 && itemCal.MaiorProbabilidade < filtro.ProbabilidadeMinima {
			continue
		}

		resultado = append(resultado, itemCal)

		// Contabiliza frequência geral para estatística agregada
		for _, m := range itemCal.Meses {
			if m.Probabilidade >= 50.0 {
				mesesFrequentes[m.NomeMes]++
			}
		}
	}

	// Ordenação Institucional:
	// 1. Posições que o usuário possui na carteira aparecem primeiro
	// 2. Ativos ordenados pelo Score Fundamentalista e Maior Probabilidade
	sort.Slice(resultado, func(i, j int) bool {
		if resultado[i].PossuiNaCarteira != resultado[j].PossuiNaCarteira {
			return resultado[i].PossuiNaCarteira // Carteira no topo
		}
		if resultado[i].Score != resultado[j].Score {
			return resultado[i].Score > resultado[j].Score
		}
		return resultado[i].MaiorProbabilidade > resultado[j].MaiorProbabilidade
	})

	return &ResumoCalendario{
		TotalAtivos:         len(resultado),
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
