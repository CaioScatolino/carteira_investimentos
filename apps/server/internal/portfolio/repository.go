package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrPosicaoNaoEncontrada = errors.New("posição não encontrada na carteira")
)

// Repository define o contrato para persistência segura da carteira.
type Repository interface {
	ListarPosicoes(ctx context.Context, usuarioID int64) ([]Posicao, error)
	ObterPosicaoPorID(ctx context.Context, usuarioID, posicaoID int64) (*Posicao, error)
	SalvarPosicao(ctx context.Context, pos *Posicao) error
	ExcluirPosicao(ctx context.Context, usuarioID, posicaoID int64) error
	RegistrarTransacao(ctx context.Context, t *Transacao) error
	ListarTransacoes(ctx context.Context, usuarioID int64) ([]Transacao, error)
}

type MySQLRepository struct {
	db *sql.DB
}

func NovoMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// ListarPosicoes recupera todas as posições ativas do usuário logado (Isolamento Multi-Tenant).
func (r *MySQLRepository) ListarPosicoes(ctx context.Context, usuarioID int64) ([]Posicao, error) {
	query := `
		SELECT id, usuario_id, ticker, classe, categoria_macro, quantidade, preco_medio,
		       valor_aplicado, taxa_contratada, indexador, vencimento, corretora, observacoes,
		       created_at, updated_at
		FROM carteira_posicoes
		WHERE usuario_id = ?
		ORDER BY classe ASC, valor_aplicado DESC`

	rows, err := r.db.QueryContext(ctx, query, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar posições da carteira: %w", err)
	}
	defer rows.Close()

	var posicoes []Posicao
	for rows.Next() {
		var p Posicao
		var vencimento sql.NullTime

		err := rows.Scan(
			&p.ID, &p.UsuarioID, &p.Ticker, &p.Classe, &p.CategoriaMacro,
			&p.Quantidade, &p.PrecoMedio, &p.ValorAplicado,
			&p.TaxaContratada, &p.Indexador, &vencimento,
			&p.Corretora, &p.Observacoes, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao ler linha de posição: %w", err)
		}

		if vencimento.Valid {
			p.Vencimento = &vencimento.Time
		}

		posicoes = append(posicoes, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posicoes, nil
}

// ObterPosicaoPorID busca uma posição específica garantindo proteção Anti-IDOR via usuario_id.
func (r *MySQLRepository) ObterPosicaoPorID(ctx context.Context, usuarioID, posicaoID int64) (*Posicao, error) {
	query := `
		SELECT id, usuario_id, ticker, classe, categoria_macro, quantidade, preco_medio,
		       valor_aplicado, taxa_contratada, indexador, vencimento, corretora, observacoes,
		       created_at, updated_at
		FROM carteira_posicoes
		WHERE id = ? AND usuario_id = ?`

	var p Posicao
	var vencimento sql.NullTime

	err := r.db.QueryRowContext(ctx, query, posicaoID, usuarioID).Scan(
		&p.ID, &p.UsuarioID, &p.Ticker, &p.Classe, &p.CategoriaMacro,
		&p.Quantidade, &p.PrecoMedio, &p.ValorAplicado,
		&p.TaxaContratada, &p.Indexador, &vencimento,
		&p.Corretora, &p.Observacoes, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPosicaoNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar posição por ID: %w", err)
	}

	if vencimento.Valid {
		p.Vencimento = &vencimento.Time
	}

	return &p, nil
}

// SalvarPosicao insere ou atualiza uma posição via Upsert inteligente no MySQL.
func (r *MySQLRepository) SalvarPosicao(ctx context.Context, pos *Posicao) error {
	query := `
		INSERT INTO carteira_posicoes (
			usuario_id, ticker, classe, categoria_macro, quantidade, preco_medio,
			valor_aplicado, taxa_contratada, indexador, vencimento, corretora, observacoes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			quantidade = VALUES(quantidade),
			preco_medio = VALUES(preco_medio),
			valor_aplicado = VALUES(valor_aplicado),
			taxa_contratada = VALUES(taxa_contratada),
			indexador = VALUES(indexador),
			vencimento = VALUES(vencimento),
			categoria_macro = VALUES(categoria_macro),
			observacoes = VALUES(observacoes),
			updated_at = CURRENT_TIMESTAMP()`

	var vencimentoVal any = nil
	if pos.Vencimento != nil {
		vencimentoVal = pos.Vencimento.Format("2006-01-02")
	}

	res, err := r.db.ExecContext(ctx, query,
		pos.UsuarioID, pos.Ticker, pos.Classe, pos.CategoriaMacro,
		pos.Quantidade, pos.PrecoMedio, pos.ValorAplicado,
		pos.TaxaContratada, pos.Indexador, vencimentoVal,
		pos.Corretora, pos.Observacoes,
	)
	if err != nil {
		return fmt.Errorf("falha ao salvar/atualizar posição: %w", err)
	}

	if pos.ID == 0 {
		insertedID, err := res.LastInsertId()
		if err == nil && insertedID > 0 {
			pos.ID = insertedID
		}
	}

	return nil
}

// ExcluirPosicao remove com proteção Anti-IDOR (apenas se pertencer ao usuário logado).
func (r *MySQLRepository) ExcluirPosicao(ctx context.Context, usuarioID, posicaoID int64) error {
	query := `DELETE FROM carteira_posicoes WHERE id = ? AND usuario_id = ?`

	res, err := r.db.ExecContext(ctx, query, posicaoID, usuarioID)
	if err != nil {
		return fmt.Errorf("falha ao excluir posição: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPosicaoNaoEncontrada
	}

	return nil
}

// RegistrarTransacao insere um novo evento de compra/venda no histórico contábil.
func (r *MySQLRepository) RegistrarTransacao(ctx context.Context, t *Transacao) error {
	query := `
		INSERT INTO carteira_transacoes (
			usuario_id, posicao_id, ticker, tipo, quantidade, preco_unitario,
			custos_taxas, valor_total, data_operacao, corretora, notas
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, query,
		t.UsuarioID, t.PosicaoID, t.Ticker, t.Tipo, t.Quantidade,
		t.PrecoUnitario, t.CustosTaxas, t.ValorTotal,
		t.DataOperacao.Format("2006-01-02 15:04:05"),
		t.Corretora, t.Notas,
	)
	if err != nil {
		return fmt.Errorf("falha ao registrar transação no livro-razão: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		t.ID = id
	}

	return nil
}

// ListarTransacoes recupera o extrato de movimentações contábeis do usuário logado.
func (r *MySQLRepository) ListarTransacoes(ctx context.Context, usuarioID int64) ([]Transacao, error) {
	query := `
		SELECT id, usuario_id, posicao_id, ticker, tipo, quantidade, preco_unitario,
		       custos_taxas, valor_total, data_operacao, corretora, notas, created_at
		FROM carteira_transacoes
		WHERE usuario_id = ?
		ORDER BY data_operacao DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar transações: %w", err)
	}
	defer rows.Close()

	var transacoes []Transacao
	for rows.Next() {
		var t Transacao
		var posicaoID sql.NullInt64

		err := rows.Scan(
			&t.ID, &t.UsuarioID, &posicaoID, &t.Ticker, &t.Tipo,
			&t.Quantidade, &t.PrecoUnitario, &t.CustosTaxas, &t.ValorTotal,
			&t.DataOperacao, &t.Corretora, &t.Notas, &t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao ler linha de transação: %w", err)
		}

		if posicaoID.Valid {
			t.PosicaoID = &posicaoID.Int64
		}

		transacoes = append(transacoes, t)
	}

	return transacoes, rows.Err()
}
