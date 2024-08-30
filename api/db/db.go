package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Db struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, uri string) (*Db, error) {
	pool, err := pgxpool.New(ctx, uri)

	if err != nil {
		return nil, err
	}

	return &Db{pool}, nil
}

func (d *Db) Ping(ctx context.Context) error {
	pingCtx, cancelPingCtx := context.WithTimeout(ctx, time.Second*5)
	defer cancelPingCtx()

	return d.Pool.Ping(pingCtx)
}

func (d *Db) Close() {
	d.Pool.Close()
}
