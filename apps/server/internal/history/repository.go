package history

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository define as operações de persistência do histórico diário
type Repository interface {
	SalvarSnapshotLote(ctx context.Context, registros []*RegistroHistorico) error
	BuscarHistoricoTicker(ctx context.Context, ticker string, limiteDias int) ([]*RegistroHistorico, error)
}

type MySQLRepository struct {
	db *sql.DB
}

// NovoMySQLRepository instancia o repositório
func NovoMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// SalvarSnapshotLote executa a gravação atômica em lote com idempotência (ON DUPLICATE KEY UPDATE)
func (r *MySQLRepository) SalvarSnapshotLote(ctx context.Context, registros []*RegistroHistorico) error {
	if len(registros) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO historico_rankings_diarios (
			data_pregao, ticker, classe, posicao_ranking, score, preco_atual,
			preco_teto_consolidado, premio_desconto_percentual, dy, pl, pvp,
			status, is_provento_atipico, modelos_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			posicao_ranking = VALUES(posicao_ranking),
			score = VALUES(score),
			preco_atual = VALUES(preco_atual),
			preco_teto_consolidado = VALUES(preco_teto_consolidado),
			premio_desconto_percentual = VALUES(premio_desconto_percentual),
			dy = VALUES(dy),
			pl = VALUES(pl),
			pvp = VALUES(pvp),
			status = VALUES(status),
			is_provento_atipico = VALUES(is_provento_atipico),
			modelos_json = VALUES(modelos_json);
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	for _, reg := range registros {
		dataStr := reg.DataPregao.Format("2006-01-02")
		isAtipicoInt := 0
		if reg.IsProventoAtipico {
			isAtipicoInt = 1
		}

		_, err := stmt.ExecContext(ctx,
			dataStr,
			reg.Ticker,
			reg.Classe,
			reg.PosicaoRanking,
			reg.Score,
			reg.PrecoAtual,
			reg.PrecoTetoConsolidado,
			reg.PremioDescontoPercentual,
			reg.DY,
			reg.PL,
			reg.PVP,
			reg.Status,
			isAtipicoInt,
			reg.ModelosJSON,
		)
		if err != nil {
			return fmt.Errorf("erro ao gravar snapshot do ativo %s: %w", reg.Ticker, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("falha ao comitar transação de histórico: %w", err)
	}

	return nil
}

// BuscarHistoricoTicker retorna os últimos N pregões de um ativo para plotar no gráfico do modal
func (r *MySQLRepository) BuscarHistoricoTicker(ctx context.Context, ticker string, limiteDias int) ([]*RegistroHistorico, error) {
	query := `
		SELECT id, data_pregao, ticker, classe, posicao_ranking, score, preco_atual,
		       preco_teto_consolidado, premio_desconto_percentual, dy, 
		       COALESCE(pl, 0.0), COALESCE(pvp, 0.0), status, is_provento_atipico, 
		       COALESCE(modelos_json, '{}'), created_at
		FROM historico_rankings_diarios
		WHERE ticker = ?
		ORDER BY data_pregao ASC
		LIMIT ?;
	`

	rows, err := r.db.QueryContext(ctx, query, ticker, limiteDias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []*RegistroHistorico
	for rows.Next() {
		var reg RegistroHistorico
		var isAtipicoInt int
		if err := rows.Scan(
			&reg.ID, &reg.DataPregao, &reg.Ticker, &reg.Classe, &reg.PosicaoRanking,
			&reg.Score, &reg.PrecoAtual, &reg.PrecoTetoConsolidado,
			&reg.PremioDescontoPercentual, &reg.DY, &reg.PL, &reg.PVP,
			&reg.Status, &isAtipicoInt, &reg.ModelosJSON, &reg.CreatedAt,
		); err != nil {
			return nil, err
		}
		reg.IsProventoAtipico = (isAtipicoInt == 1)
		lista = append(lista, &reg)
	}

	return lista, nil
}
