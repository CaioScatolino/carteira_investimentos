package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver MySQL oficial
)

// NovoMySQLConnection inicializa o pool de conexões otimizado
func NovoMySQLConnection(usuario, senha, host, porta, banco string) (*sql.DB, error) {
	var dsn string
	if senha == "" {
		dsn = fmt.Sprintf("%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			usuario, host, porta, banco)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			usuario, senha, host, porta, banco)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar driver MySQL: %w", err)
	}

	// Connection Pool de Alta Performance
	db.SetMaxOpenConns(25)                 // Máximo de conexões abertas
	db.SetMaxIdleConns(5)                  // Conexões prontas em repouso
	db.SetConnMaxLifetime(5 * time.Minute) // Reciclagem periódica

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("falha ao conectar ao MySQL (%s:%s): %w", host, porta, err)
	}

	return db, nil
}
