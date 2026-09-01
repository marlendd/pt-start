package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const databasePingTimeout = 5 * time.Second

func openDatabase(
	ctx context.Context,
	databaseURL string,
) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"create database pool: %w",
			err,
		)
	}

	pingCtx, cancel := context.WithTimeout(
		ctx,
		databasePingTimeout,
	)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"ping database: %w",
			err,
		)
	}

	return pool, nil
}
