package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresLogger struct {
	pool *pgxpool.Pool
}

func NewPostgresLogger(ctx context.Context, dbURL string) (*PostgresLogger, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	
	// Ensure table exists
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS audit_events (
			id SERIAL PRIMARY KEY,
			request_id VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			trace JSONB NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}
	
	return &PostgresLogger{pool: pool}, nil
}

func (p *PostgresLogger) Log(ctx context.Context, event Event) error {
	raw, _ := json.Marshal(event)
	
	_, err := p.pool.Exec(ctx, `
		INSERT INTO audit_events (request_id, timestamp, trace)
		VALUES ($1, $2, $3)
	`, event.RequestID, event.Timestamp, string(raw))
	
	return err
}
