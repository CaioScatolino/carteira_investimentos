package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Repository define as operações de persistência de usuários
type Repository interface {
	Criar(ctx context.Context, u *Usuario) error
	BuscarPorEmail(ctx context.Context, email string) (*Usuario, error)
	BuscarPorUUID(ctx context.Context, uuid string) (*Usuario, error)
	AtualizarTentativasFalhas(ctx context.Context, id int64, tentativas int, bloqueadoAte *time.Time) error
	RegistrarLoginSucesso(ctx context.Context, id int64) error
}

type MySQLRepository struct {
	db *sql.DB
}

// NovoMySQLRepository instancia o repositório
func NovoMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Criar persiste um novo usuário com prepared statement
func (r *MySQLRepository) Criar(ctx context.Context, u *Usuario) error {
	query := `
		INSERT INTO usuarios (uuid, nome, email, senha_hash, role, status)
		VALUES (?, ?, ?, ?, ?, ?);
	`
	res, err := r.db.ExecContext(ctx, query, u.UUID, u.Nome, u.Email, u.SenhaHash, u.Role, u.Status)
	if err != nil {
		return fmt.Errorf("erro ao inserir usuário: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		u.ID = id
	}
	return nil
}

// BuscarPorEmail localiza um usuário pelo e-mail
func (r *MySQLRepository) BuscarPorEmail(ctx context.Context, email string) (*Usuario, error) {
	query := `
		SELECT id, uuid, nome, email, senha_hash, role, status,
		       tentativas_falhas, bloqueado_ate, ultimo_login_at, created_at, updated_at
		FROM usuarios
		WHERE email = ?
		LIMIT 1;
	`
	var u Usuario
	var bloqueadoAte, ultimoLogin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.UUID, &u.Nome, &u.Email, &u.SenhaHash, &u.Role, &u.Status,
		&u.TentativasFalhas, &bloqueadoAte, &ultimoLogin, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Retorna nil explicitamente quando não encontrado
		}
		return nil, fmt.Errorf("erro ao consultar usuário por email: %w", err)
	}

	if bloqueadoAte.Valid {
		u.BloqueadoAte = &bloqueadoAte.Time
	}
	if ultimoLogin.Valid {
		u.UltimoLoginAt = &ultimoLogin.Time
	}

	return &u, nil
}

// BuscarPorUUID localiza o usuário pelo identificador público seguro
func (r *MySQLRepository) BuscarPorUUID(ctx context.Context, uuid string) (*Usuario, error) {
	query := `
		SELECT id, uuid, nome, email, senha_hash, role, status,
		       tentativas_falhas, bloqueado_ate, ultimo_login_at, created_at, updated_at
		FROM usuarios
		WHERE uuid = ?
		LIMIT 1;
	`
	var u Usuario
	var bloqueadoAte, ultimoLogin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&u.ID, &u.UUID, &u.Nome, &u.Email, &u.SenhaHash, &u.Role, &u.Status,
		&u.TentativasFalhas, &bloqueadoAte, &ultimoLogin, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao consultar usuário por uuid: %w", err)
	}

	if bloqueadoAte.Valid {
		u.BloqueadoAte = &bloqueadoAte.Time
	}
	if ultimoLogin.Valid {
		u.UltimoLoginAt = &ultimoLogin.Time
	}

	return &u, nil
}

// AtualizarTentativasFalhas atualiza a contagem de falhas e data de bloqueio se aplicável
func (r *MySQLRepository) AtualizarTentativasFalhas(ctx context.Context, id int64, tentativas int, bloqueadoAte *time.Time) error {
	query := `
		UPDATE usuarios
		SET tentativas_falhas = ?, bloqueado_ate = ?
		WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query, tentativas, bloqueadoAte, id)
	return err
}

// RegistrarLoginSucesso reseta contadores de falhas e atualiza timestamp de login
func (r *MySQLRepository) RegistrarLoginSucesso(ctx context.Context, id int64) error {
	query := `
		UPDATE usuarios
		SET tentativas_falhas = 0, bloqueado_ate = NULL, ultimo_login_at = ?
		WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
