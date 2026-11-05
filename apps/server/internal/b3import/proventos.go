package b3import

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

type ProventoRecebidoLinha struct {
	ID            int64     `json:"id"`
	Ticker        string    `json:"ticker"`
	Tipo          string    `json:"tipo"`
	DataPagamento string    `json:"data_pagamento"`
	Quantidade    float64   `json:"quantidade"`
	ValorUnitario float64   `json:"valor_unitario"`
	ValorLiquido  float64   `json:"valor_liquido"`
	Corretora     string    `json:"corretora"`
}

type TotalAno struct {
	Ano   int     `json:"ano"`
	Valor float64 `json:"valor"`
}

type SazonalidadeMes struct {
	Mes   int     `json:"mes"`
	Nome  string  `json:"nome"`
	Valor float64 `json:"valor"`
}

type AtivoPagador struct {
	Ticker         string  `json:"ticker"`
	TotalRecebido  float64 `json:"total_recebido"`
	Percentual     float64 `json:"percentual"`
	TotalEventos   int     `json:"total_eventos"`
}

type ResumoProventos struct {
	TotalAcumulado      float64                 `json:"total_acumulado"`
	MediaMensal12M      float64                 `json:"media_mensal_12m"`
	TotalRecebido12M    float64                 `json:"total_recebido_12m"`
	TotalPorAno         []TotalAno              `json:"total_por_ano"`
	SazonalidadeMensal  []SazonalidadeMes       `json:"sazonalidade_mensal"`
	TopAtivosPagadores  []AtivoPagador          `json:"top_ativos_pagadores"`
	UltimosPagamentos   []ProventoRecebidoLinha `json:"ultimos_pagamentos"`
}

type ProventoFuturo struct {
	Ticker           string  `json:"ticker"`
	Tipo             string  `json:"tipo"`
	DataCom          string  `json:"data_com"`
	DataPagamento    string  `json:"data_pagamento"`
	ValorUnitario    float64 `json:"valor_unitario"`
	QuantidadeCotas  float64 `json:"quantidade_cotas"`
	ValorPrevisto    float64 `json:"valor_previsto"`
	Status           string  `json:"status"` // "DATA_COM_IMINENTE", "AGUARDANDO_PAGAMENTO", "CREDITADO"
}

var nomesMesesPT = []string{
	"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

// ObterHistoricoProventos retorna o panorama analítico de renda passiva do investidor
func (s *Service) ObterHistoricoProventos(ctx context.Context, usuarioID int64) (*ResumoProventos, error) {
	if s.db == nil {
		return nil, fmt.Errorf("banco de dados indisponível")
	}

	query := `
		SELECT id, ticker, tipo, data_pagamento, quantidade, valor_unitario, valor_liquido, corretora
		FROM proventos_recebidos
		WHERE usuario_id = ?
		ORDER BY data_pagamento DESC, valor_liquido DESC
	`
	rows, err := s.db.QueryContext(ctx, query, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar proventos: %w", err)
	}
	defer rows.Close()

	var todos []ProventoRecebidoLinha
	totalAcumulado := 0.0
	totalPorAnoMap := make(map[int]float64)
	sazonalidadeMap := make(map[int]float64)
	ativosMap := make(map[string]*AtivoPagador)

	agora := time.Now()
	limite12M := agora.AddDate(-1, 0, 0)
	totalRecebido12M := 0.0

	for rows.Next() {
		var p ProventoRecebidoLinha
		var dt time.Time
		if err := rows.Scan(&p.ID, &p.Ticker, &p.Tipo, &dt, &p.Quantidade, &p.ValorUnitario, &p.ValorLiquido, &p.Corretora); err != nil {
			return nil, err
		}
		p.DataPagamento = dt.Format("02/01/2006")
		todos = append(todos, p)

		totalAcumulado += p.ValorLiquido
		ano := dt.Year()
		mes := int(dt.Month())
		totalPorAnoMap[ano] += p.ValorLiquido
		sazonalidadeMap[mes] += p.ValorLiquido

		if dt.After(limite12M) {
			totalRecebido12M += p.ValorLiquido
		}

		if at, ok := ativosMap[p.Ticker]; ok {
			at.TotalRecebido += p.ValorLiquido
			at.TotalEventos++
		} else {
			ativosMap[p.Ticker] = &AtivoPagador{
				Ticker:        p.Ticker,
				TotalRecebido: p.ValorLiquido,
				TotalEventos:  1,
			}
		}
	}

	// 1. Montar array ordenado de Anos
	totalPorAno := make([]TotalAno, 0)
	for ano, val := range totalPorAnoMap {
		totalPorAno = append(totalPorAno, TotalAno{Ano: ano, Valor: math.Round(val*100) / 100})
	}
	sort.Slice(totalPorAno, func(i, j int) bool {
		return totalPorAno[i].Ano < totalPorAno[j].Ano
	})

	// 2. Montar array ordenado de Meses (1 a 12)
	sazonalidade := make([]SazonalidadeMes, 0, 12)
	for m := 1; m <= 12; m++ {
		sazonalidade = append(sazonalidade, SazonalidadeMes{
			Mes:   m,
			Nome:  nomesMesesPT[m],
			Valor: math.Round(sazonalidadeMap[m]*100) / 100,
		})
	}

	// 3. Montar Top Ativos Pagadores
	topAtivos := make([]AtivoPagador, 0)
	for _, at := range ativosMap {
		pct := 0.0
		if totalAcumulado > 0 {
			pct = (at.TotalRecebido / totalAcumulado) * 100
		}
		topAtivos = append(topAtivos, AtivoPagador{
			Ticker:        at.Ticker,
			TotalRecebido: math.Round(at.TotalRecebido*100) / 100,
			Percentual:    math.Round(pct*10) / 10,
			TotalEventos:  at.TotalEventos,
		})
	}
	sort.Slice(topAtivos, func(i, j int) bool {
		return topAtivos[i].TotalRecebido > topAtivos[j].TotalRecebido
	})

	// 4. Últimos pagamentos (limite de 30 para visualização)
	ultimos := make([]ProventoRecebidoLinha, 0)
	if len(todos) > 0 {
		if len(todos) > 30 {
			ultimos = todos[:30]
		} else {
			ultimos = todos
		}
	}

	return &ResumoProventos{
		TotalAcumulado:     math.Round(totalAcumulado*100) / 100,
		MediaMensal12M:     math.Round((totalRecebido12M/12)*100) / 100,
		TotalRecebido12M:   math.Round(totalRecebido12M*100) / 100,
		TotalPorAno:        totalPorAno,
		SazonalidadeMensal: sazonalidade,
		TopAtivosPagadores: topAtivos,
		UltimosPagamentos:  ultimos,
	}, nil
}
