package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Service struct {
	DB *sql.DB
}

func New(ctx context.Context, dsn string, maxAttempts int, retryInterval time.Duration) (*Service, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxAttempts; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			if err = db.PingContext(ctx); err == nil {
				break
			}
		}
		if i < maxAttempts-1 {
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Service{DB: db}, nil
}

func (s *Service) Close() error {
	return s.DB.Close()
}