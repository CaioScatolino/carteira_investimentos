package catalog

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	BuscarPorTicker(ctx context.Context, ticker string) (*Asset, error)
	BuscarPorCNPJ(ctx context.Context, cnpj string) (*Asset, error)
	ListarTodos(ctx context.Context) ([]*Asset, error)
	SalvarEmLote(ctx context.Context, assets []*Asset) error
}

type MySQLRepository struct {
	db *sql.DB
}

func NovoMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) BuscarPorTicker(ctx context.Context, ticker string) (*Asset, error) {
	query := `
		SELECT id, ticker, cnpj, classe, COALESCE(tipo, ''), razao_social, 
		       COALESCE(codigo_cvm, ''), COALESCE(setor, ''), ativo, created_at, updated_at
		FROM assets 
		WHERE ticker = ? AND ativo = 1 LIMIT 1`

	var a Asset
	err := r.db.QueryRowContext(ctx, query, ticker).Scan(
		&a.ID, &a.Ticker, &a.CNPJ, &a.Classe, &a.Tipo,
		&a.RazaoSocial, &a.CodigoCVM, &a.Setor, &a.Ativo,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ativo %s não encontrado no catálogo", ticker)
		}
		return nil, err
	}
	return &a, nil
}

func (r *MySQLRepository) BuscarPorCNPJ(ctx context.Context, cnpj string) (*Asset, error) {
	query := `
		SELECT id, ticker, cnpj, classe, COALESCE(tipo, ''), razao_social, 
		       COALESCE(codigo_cvm, ''), COALESCE(setor, ''), ativo, created_at, updated_at
		FROM assets 
		WHERE cnpj = ? AND ativo = 1 LIMIT 1`

	var a Asset
	err := r.db.QueryRowContext(ctx, query, cnpj).Scan(
		&a.ID, &a.Ticker, &a.CNPJ, &a.Classe, &a.Tipo,
		&a.RazaoSocial, &a.CodigoCVM, &a.Setor, &a.Ativo,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("CNPJ %s não encontrado no catálogo", cnpj)
		}
		return nil, err
	}
	return &a, nil
}

func (r *MySQLRepository) ListarTodos(ctx context.Context) ([]*Asset, error) {
	query := `
		SELECT id, ticker, cnpj, classe, COALESCE(tipo, ''), razao_social, 
		       COALESCE(codigo_cvm, ''), COALESCE(setor, ''), ativo, created_at, updated_at
		FROM assets 
		WHERE ativo = 1 
		ORDER BY classe, ticker`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []*Asset
	for rows.Next() {
		var a Asset
		if err := rows.Scan(
			&a.ID, &a.Ticker, &a.CNPJ, &a.Classe, &a.Tipo,
			&a.RazaoSocial, &a.CodigoCVM, &a.Setor, &a.Ativo,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, &a)
	}

	return lista, nil
}

func (r *MySQLRepository) SalvarEmLote(ctx context.Context, assets []*Asset) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO assets (ticker, cnpj, classe, tipo, razao_social, codigo_cvm, setor, ativo)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1)
		ON DUPLICATE KEY UPDATE
			cnpj = VALUES(cnpj),
			classe = VALUES(classe),
			tipo = VALUES(tipo),
			razao_social = VALUES(razao_social),
			codigo_cvm = VALUES(codigo_cvm),
			setor = VALUES(setor),
			ativo = 1,
			updated_at = CURRENT_TIMESTAMP`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, a := range assets {
		if a.Ticker == "" || a.CNPJ == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, a.Ticker, a.CNPJ, a.Classe, a.Tipo, a.RazaoSocial, a.CodigoCVM, a.Setor); err != nil {
			return err
		}
	}

	return tx.Commit()
}
