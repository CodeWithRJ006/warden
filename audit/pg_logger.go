package audit

import (
	"context"
	"encoding/json"
	"strings"

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
		if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "42P07") {
			return nil, err
		}
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

func (p *PostgresLogger) GetLogs(ctx context.Context, limit int) ([]Event, error) {
	rows, err := p.pool.Query(ctx, "SELECT trace FROM audit_events ORDER BY id DESC LIMIT $1", limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var traceStr string
		if err := rows.Scan(&traceStr); err != nil { continue }
		var event Event
		json.Unmarshal([]byte(traceStr), &event)
		events = append(events, event)
	}
	return events, nil
}

