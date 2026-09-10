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
			timestamp TIMESTAMP NOT NULL,
			actor VARCHAR(255) NOT NULL,
			action VARCHAR(255) NOT NULL,
			amount DECIMAL(15,2),
			decision VARCHAR(50) NOT NULL,
			reason TEXT,
			policy_version VARCHAR(50) NOT NULL,
			raw_payload JSONB
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
		INSERT INTO audit_events (timestamp, actor, action, amount, decision, reason, policy_version, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.Timestamp, event.Actor, event.Action, event.Amount, event.Decision, event.Reason, event.PolicyVer, string(raw))
	
	return err
}
