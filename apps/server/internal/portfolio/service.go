package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"carteira_investimentos/server/internal/storage"
)

var (
	ErrDadosInvalidos = errors.New("dados de entrada inválidos para a posição")
)

type Service struct {
	repo  Repository
	store storage.SnapshotStore
}

func NovoService(repo Repository, store storage.SnapshotStore) *Service {
	return &Service{
		repo:  repo,
		store: store,
	}
}

// AdicionarOuAtualizarPosicao valida as regras financeiras, persiste a posição e registra no livro-razão.
func (s *Service) AdicionarOuAtualizarPosicao(ctx context.Context, usuarioID int64, req CriarPosicaoRequest) (*Posicao, error) {
	ticker := strings.ToUpper(strings.TrimSpace(req.Ticker))
	if ticker == "" {
		return nil, fmt.Errorf("%w: o ticker não pode ser vazio", ErrDadosInvalidos)
	}

	if req.Quantidade <= 0 {
		return nil, fmt.Errorf("%w: a quantidade deve ser maior que zero", ErrDadosInvalidos)
	}

	if req.PrecoMedio <= 0 {
		return nil, fmt.Errorf("%w: o preço médio deve ser maior que zero", ErrDadosInvalidos)
	}

	// Validação das 5 Macro Classes
	switch req.Classe {
	case ClasseRendaFixa, ClasseFII, ClasseAcaoBR, ClasseInternacional, ClasseCaixa:
		// Classe válida
	default:
		return nil, fmt.Errorf("%w: classe de ativo '%s' inválida", ErrDadosInvalidos, req.Classe)
	}

	corretora := strings.TrimSpace(req.Corretora)
	if corretora == "" {
		corretora = "Principal"
	}

	categoriaMacro := strings.TrimSpace(req.CategoriaMacro)
	if categoriaMacro == "" {
		categoriaMacro = "GERAL"
	}

	valorAplicado := req.Quantidade * req.PrecoMedio

	var dataVencimento *time.Time
	if req.Vencimento != nil && *req.Vencimento != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.Vencimento)
		if err == nil {
			dataVencimento = &parsedDate
		}
	}

	// 1. Verifica se já existe uma posição prévia para este usuário, ticker e corretora
	posExistente, err := s.repo.ObterPosicaoPorTicker(ctx, usuarioID, ticker, corretora)
	if err == nil && posExistente != nil {
		tipoOperacao := strings.ToUpper(strings.TrimSpace(req.TipoOperacao))

		// Modo Substituição Manual Explícita (correção de lançamento)
		if tipoOperacao == "SUBSTITUIR" || tipoOperacao == "EDICAO" {
			posExistente.Quantidade = req.Quantidade
			posExistente.PrecoMedio = req.PrecoMedio
			posExistente.ValorAplicado = req.Quantidade * req.PrecoMedio
			if req.Classe != "" {
				posExistente.Classe = req.Classe
			}
			if categoriaMacro != "GERAL" {
				posExistente.CategoriaMacro = categoriaMacro
			}
			posExistente.TaxaContratada = req.TaxaContratada
			posExistente.Indexador = req.Indexador
			posExistente.Vencimento = dataVencimento
			if req.Observacoes != nil {
				posExistente.Observacoes = req.Observacoes
			}

			if err := s.repo.SalvarPosicao(ctx, posExistente); err != nil {
				return nil, err
			}

			nota := fmt.Sprintf("Ajuste manual de custódia: Qtd: %.4f, PM: R$ %.2f", req.Quantidade, req.PrecoMedio)
			transacao := &Transacao{
				UsuarioID:     usuarioID,
				PosicaoID:     &posExistente.ID,
				Ticker:        ticker,
				Tipo:          TransacaoAporte,
				Quantidade:    req.Quantidade,
				PrecoUnitario: req.PrecoMedio,
				CustosTaxas:   0.0,
				ValorTotal:    posExistente.ValorAplicado,
				DataOperacao:  time.Now(),
				Corretora:     corretora,
				Notas:         &nota,
			}
			_ = s.repo.RegistrarTransacao(ctx, transacao)

			return posExistente, nil
		}

		// Modo Padrão: Novo Aporte com Média Ponderada Contábil
		// Fórmula: PM_novo = ((Q_antiga * PM_antigo) + (Q_nova * P_compra)) / (Q_antiga + Q_nova)
		qtdAntiga := posExistente.Quantidade
		pmAntigo := posExistente.PrecoMedio
		valAntigo := posExistente.ValorAplicado
		if valAntigo <= 0 {
			valAntigo = qtdAntiga * pmAntigo
		}

		qtdNova := req.Quantidade
		precoCompra := req.PrecoMedio
		valNovo := qtdNova * precoCompra

		qtdTotal := qtdAntiga + qtdNova
		if qtdTotal <= 0 {
			return nil, fmt.Errorf("%w: a quantidade total consolidada deve ser maior que zero", ErrDadosInvalidos)
		}

		novoPrecoMedio := (valAntigo + valNovo) / qtdTotal
		novoValorAplicado := valAntigo + valNovo

		posExistente.Quantidade = qtdTotal
		posExistente.PrecoMedio = novoPrecoMedio
		posExistente.ValorAplicado = novoValorAplicado
		if req.Classe != "" {
			posExistente.Classe = req.Classe
		}
		if categoriaMacro != "GERAL" {
			posExistente.CategoriaMacro = categoriaMacro
		}
		if req.TaxaContratada != nil {
			posExistente.TaxaContratada = req.TaxaContratada
		}
		if req.Indexador != nil {
			posExistente.Indexador = req.Indexador
		}
		if dataVencimento != nil {
			posExistente.Vencimento = dataVencimento
		}
		if req.Observacoes != nil {
			posExistente.Observacoes = req.Observacoes
		}

		if err := s.repo.SalvarPosicao(ctx, posExistente); err != nil {
			return nil, err
		}

		// Registra o aporte no livro-razão mantendo rastreabilidade contábil e fiscal
		nota := fmt.Sprintf("Aporte: +%.4f cotas a R$ %.2f (PM: R$ %.2f -> R$ %.2f)", qtdNova, precoCompra, pmAntigo, novoPrecoMedio)
		transacao := &Transacao{
			UsuarioID:     usuarioID,
			PosicaoID:     &posExistente.ID,
			Ticker:        ticker,
			Tipo:          TransacaoAporte,
			Quantidade:    qtdNova,
			PrecoUnitario: precoCompra,
			CustosTaxas:   0.0,
			ValorTotal:    valNovo,
			DataOperacao:  time.Now(),
			Corretora:     corretora,
			Notas:         &nota,
		}
		_ = s.repo.RegistrarTransacao(ctx, transacao)

		return posExistente, nil
	}

	// 2. Não possui posição prévia: Cadastro Inicial
	pos := &Posicao{
		UsuarioID:      usuarioID,
		Ticker:         ticker,
		Classe:         req.Classe,
		CategoriaMacro: categoriaMacro,
		Quantidade:     req.Quantidade,
		PrecoMedio:     req.PrecoMedio,
		ValorAplicado:  valorAplicado,
		TaxaContratada: req.TaxaContratada,
		Indexador:      req.Indexador,
		Vencimento:     dataVencimento,
		Corretora:      corretora,
		Observacoes:    req.Observacoes,
	}

	// Salva no banco de dados
	if err := s.repo.SalvarPosicao(ctx, pos); err != nil {
		return nil, err
	}

	// Registra a transação inicial de compra no livro-razão contábil
	nota := fmt.Sprintf("Posição inicial: %.4f cotas a R$ %.2f", req.Quantidade, req.PrecoMedio)
	transacao := &Transacao{
		UsuarioID:     usuarioID,
		PosicaoID:     &pos.ID,
		Ticker:        ticker,
		Tipo:          TransacaoCompra,
		Quantidade:    req.Quantidade,
		PrecoUnitario: req.PrecoMedio,
		CustosTaxas:   0.0,
		ValorTotal:    valorAplicado,
		DataOperacao:  time.Now(),
		Corretora:     corretora,
		Notas:         &nota,
	}
	_ = s.repo.RegistrarTransacao(ctx, transacao)

	return pos, nil
}

// ExcluirPosicao remove um ativo da carteira garantindo o isolamento multi-tenant.
func (s *Service) ExcluirPosicao(ctx context.Context, usuarioID, posicaoID int64) error {
	return s.repo.ExcluirPosicao(ctx, usuarioID, posicaoID)
}

// ObterResumoCarteira consolida todas as posições do usuário, calculando rentabilidade real e alocações.
func (s *Service) ObterResumoCarteira(ctx context.Context, usuarioID int64) (*ResumoCarteira, error) {
	posicoes, err := s.repo.ListarPosicoes(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	var totalInvestido float64
	var patrimonioTotal float64
	posicoesConsolidadas := make([]PosicaoConsolidada, 0, len(posicoes))
	alocacaoValorPorClasse := make(map[ClasseAtivo]float64)

	for _, p := range posicoes {
		totalInvestido += p.ValorAplicado

		cotacaoAtual := p.PrecoMedio
		var precoTeto float64
		var scoreFund int
		sinal := "AGUARDAR"

		// Se o ativo for de renda variável e estiver no nosso Snapshot, pegamos a cotação e valuation reais
		if s.store != nil && (p.Classe == ClasseAcaoBR || p.Classe == ClasseFII || p.Classe == ClasseInternacional) {
			if ativo, err := s.store.Obter(p.Ticker); err == nil && ativo != nil {
				if ativo.PrecoAtual > 0 {
					cotacaoAtual = ativo.PrecoAtual
				}
				scoreFund = int(ativo.Score)
				if ativo.PrecoTetoConsolidado > 0 {
					precoTeto = ativo.PrecoTetoConsolidado
				} else if ativo.PrecoTetoBazin > 0 {
					precoTeto = ativo.PrecoTetoBazin
				} else if ativo.ValorGraham > 0 {
					precoTeto = ativo.ValorGraham
				}
			}
		}

		valorMercado := p.Quantidade * cotacaoAtual
		lucroValor := valorMercado - p.ValorAplicado

		var rentabilidadePct float64
		if p.ValorAplicado > 0 {
			rentabilidadePct = (lucroValor / p.ValorAplicado) * 100
		}

		var margemSegurancaPct float64
		if precoTeto > 0 && cotacaoAtual > 0 {
			margemSegurancaPct = ((precoTeto - cotacaoAtual) / cotacaoAtual) * 100
			if margemSegurancaPct > 5.0 && scoreFund >= 70 {
				sinal = "COMPRAR_APORTE"
			} else if margemSegurancaPct < -15.0 {
				sinal = "REAVALIAR"
			}
		}

		patrimonioTotal += valorMercado
		alocacaoValorPorClasse[p.Classe] += valorMercado

		posicoesConsolidadas = append(posicoesConsolidadas, PosicaoConsolidada{
			Posicao:              p,
			CotacaoAtual:         cotacaoAtual,
			ValorAtualMercado:    valorMercado,
			LucroPrejuizoValor:   lucroValor,
			RentabilidadePct:     rentabilidadePct,
			PrecoTetoEstimado:    precoTeto,
			MargemSegurancaPct:   margemSegurancaPct,
			ScoreFundamentalista: scoreFund,
			SinalRebalanceamento: sinal,
		})
	}

	lucroTotal := patrimonioTotal - totalInvestido
	var rentabilidadeTotalPct float64
	if totalInvestido > 0 {
		rentabilidadeTotalPct = (lucroTotal / totalInvestido) * 100
	}

	// Calcula os percentuais atuais de alocação macro
	alocacaoAtualPct := make(map[ClasseAtivo]float64)
	for classe, valor := range alocacaoValorPorClasse {
		if patrimonioTotal > 0 {
			alocacaoAtualPct[classe] = (valor / patrimonioTotal) * 100
		} else {
			alocacaoAtualPct[classe] = 0.0
		}
	}

	// Metas oficiais de macro alocação patrimonial
	alocacaoAlvoPct := map[ClasseAtivo]float64{
		ClasseRendaFixa:     40.0,
		ClasseFII:           22.5,
		ClasseAcaoBR:        21.0,
		ClasseInternacional: 16.0,
		ClasseCaixa:         0.5,
	}

	return &ResumoCarteira{
		PatrimonioTotal:  patrimonioTotal,
		TotalInvestido:   totalInvestido,
		LucroTotalValor:  lucroTotal,
		RentabilidadePct: rentabilidadeTotalPct,
		Posicoes:         posicoesConsolidadas,
		AlocacaoAtual:    alocacaoAtualPct,
		AlocacaoAlvo:     alocacaoAlvoPct,
	}, nil
}

// ListarTransacoes retorna o extrato de compras e vendas do usuário.
func (s *Service) ListarTransacoes(ctx context.Context, usuarioID int64) ([]Transacao, error) {
	return s.repo.ListarTransacoes(ctx, usuarioID)
}

// ListarTransacoesPorTicker retorna o extrato filtrado por um ticker específico.
func (s *Service) ListarTransacoesPorTicker(ctx context.Context, usuarioID int64, ticker string) ([]Transacao, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	return s.repo.ListarTransacoesPorTicker(ctx, usuarioID, ticker)
}

// RebalancearCarteira consolida a carteira atual e simula a distribuição ótima do aporte.
func (s *Service) RebalancearCarteira(ctx context.Context, usuarioID int64, valorAporte float64) (*ResultadoRebalanceamento, error) {
	resumo, err := s.ObterResumoCarteira(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	return SimularRebalanceamento(resumo, valorAporte), nil
}
