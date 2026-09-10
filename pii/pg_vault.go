package pii

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresVault struct {
	pool *pgxpool.Pool
}

func NewPostgresVault(ctx context.Context, dbURL string) (*PostgresVault, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS token_vault (
			id SERIAL PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			raw_value TEXT NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "42P07") {
			return nil, err
		}
	}
	
	return &PostgresVault{pool: pool}, nil
}

func (v *PostgresVault) Store(ctx context.Context, entityType string, raw string) (string, error) {
	var id int
	err := v.pool.QueryRow(ctx, `
		INSERT INTO token_vault (entity_type, raw_value, token)
		VALUES ($1, $2, 'pending_' || gen_random_uuid())
		RETURNING id
	`, entityType, raw).Scan(&id)
	
	if err != nil {
		return "", err
	}
	
	token := fmt.Sprintf("[%s_%03d]", entityType, id)
	_, err = v.pool.Exec(ctx, `UPDATE token_vault SET token = $1 WHERE id = $2`, token, id)
	
	return token, err
}

func (v *PostgresVault) Retrieve(ctx context.Context, token string) (string, error) {
	var raw string
	err := v.pool.QueryRow(ctx, `SELECT raw_value FROM token_vault WHERE token = $1`, token).Scan(&raw)
	return raw, err
}
