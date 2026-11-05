package b3import

import (
	"context"
	"database/sql"
	"fmt"
	"io"
)

type ResultadoImportacao struct {
	PosicoesSalvas   int     `json:"posicoes_salvas"`
	OperacoesSalvas  int     `json:"operacoes_salvas"`
	ProventosSalvos  int     `json:"proventos_salvos"`
	TotalInvestido   float64 `json:"total_investido"`
	TotalProventos   float64 `json:"total_proventos"`
	DadosSubstituidos bool    `json:"dados_substituidos"`
}

type Service struct {
	db     *sql.DB
	parser *ParserB3
}

func NovoService(db *sql.DB) *Service {
	return &Service{
		db:     db,
		parser: NovoParserB3(),
	}
}

// AnalisarPlanilha executa a leitura prévia do arquivo B3 sem persistir no banco de dados
func (s *Service) AnalisarPlanilha(r io.Reader) (*RelatorioAnaliseB3, error) {
	return s.parser.AnalisarPlanilha(r)
}

// SalvarImportacao persiste atomicamente as posições, histórico contábil e proventos no MySQL
func (s *Service) SalvarImportacao(ctx context.Context, usuarioID int64, req ConfirmarImportacaoRequest) (*ResultadoImportacao, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conexão com banco de dados indisponível")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("falha ao iniciar transação de importação: %w", err)
	}
	defer tx.Rollback()

	// 1. Desativar / Limpar registros anteriores se solicitado (Anti-Duplicação)
	if req.LimparDadosAnteriores {
		if _, err := tx.ExecContext(ctx, "DELETE FROM carteira_transacoes WHERE usuario_id = ?", usuarioID); err != nil {
			return nil, fmt.Errorf("falha ao limpar transações anteriores: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM carteira_posicoes WHERE usuario_id = ?", usuarioID); err != nil {
			return nil, fmt.Errorf("falha ao limpar posições anteriores: %w", err)
		}
		if req.ImportarProventos {
			if _, err := tx.ExecContext(ctx, "DELETE FROM proventos_recebidos WHERE usuario_id = ?", usuarioID); err != nil {
				return nil, fmt.Errorf("falha ao limpar proventos anteriores: %w", err)
			}
		}
	}

	// 2. Persistir Posições Ativas Consolidadas
	posStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO carteira_posicoes 
			(usuario_id, ticker, classe, categoria_macro, quantidade, preco_medio, valor_aplicado, corretora)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			quantidade = VALUES(quantidade),
			preco_medio = VALUES(preco_medio),
			valor_aplicado = VALUES(valor_aplicado),
			classe = VALUES(classe),
			corretora = VALUES(corretora)
	`)
	if err != nil {
		return nil, fmt.Errorf("falha ao preparar statement de posições: %w", err)
	}
	defer posStmt.Close()

	totalInvestido := 0.0
	for _, p := range req.Posicoes {
		corretora := p.Corretora
		if corretora == "" {
			corretora = "Principal"
		}
		categoria := "GERAL"
		ticker := p.Ticker
		if len(ticker) > 60 {
			ticker = ticker[:60]
		}

		_, err := posStmt.ExecContext(ctx,
			usuarioID, ticker, p.Classe, categoria,
			p.Quantidade, p.PrecoMedio, p.ValorAplicado, corretora,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao persistir posição de %s: %w", p.Ticker, err)
		}
		totalInvestido += p.ValorAplicado
	}

	// 3. Persistir Histórico de Operações (Livro-Razão)
	opStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO carteira_transacoes
			(usuario_id, ticker, tipo, quantidade, preco_unitario, valor_total, data_operacao, corretora, notas)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("falha ao preparar statement de transações: %w", err)
	}
	defer opStmt.Close()

	for _, op := range req.Operacoes {
		corretora := op.Corretora
		if corretora == "" {
			corretora = "Principal"
		}
		notas := op.Notas
		if notas == "" {
			notas = "Importação B3"
		}
		ticker := op.Ticker
		if len(ticker) > 60 {
			ticker = ticker[:60]
		}

		_, err := opStmt.ExecContext(ctx,
			usuarioID, ticker, op.Tipo, op.Quantidade,
			op.PrecoUnitario, op.ValorTotal, op.Data, corretora, notas,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao persistir transação de %s: %w", op.Ticker, err)
		}
	}

	// 4. Persistir Proventos Históricos se solicitado
	totalProventosSalvos := 0
	totalProventosValor := 0.0

	if req.ImportarProventos && len(req.Proventos) > 0 {
		provStmt, err := tx.PrepareContext(ctx, `
			INSERT IGNORE INTO proventos_recebidos
				(usuario_id, ticker, tipo, data_pagamento, quantidade, valor_unitario, valor_liquido, corretora)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return nil, fmt.Errorf("falha ao preparar statement de proventos: %w", err)
		}
		defer provStmt.Close()

		for _, pv := range req.Proventos {
			corretora := pv.Corretora
			if corretora == "" {
				corretora = "Principal"
			}
			ticker := pv.Ticker
			if len(ticker) > 60 {
				ticker = ticker[:60]
			}

			res, err := provStmt.ExecContext(ctx,
				usuarioID, ticker, pv.Tipo, pv.DataPagamento,
				pv.Quantidade, pv.ValorUnitario, pv.ValorTotal, corretora,
			)
			if err == nil {
				if rows, _ := res.RowsAffected(); rows > 0 {
					totalProventosSalvos++
					totalProventosValor += pv.ValorTotal
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("falha ao confirmar transação no banco: %w", err)
	}

	return &ResultadoImportacao{
		PosicoesSalvas:    len(req.Posicoes),
		OperacoesSalvas:   len(req.Operacoes),
		ProventosSalvos:   totalProventosSalvos,
		TotalInvestido:    totalInvestido,
		TotalProventos:    totalProventosValor,
		DadosSubstituidos: req.LimparDadosAnteriores,
	}, nil
}

