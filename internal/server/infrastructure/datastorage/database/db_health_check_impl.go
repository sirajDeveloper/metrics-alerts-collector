package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

const defaultPingTimeout = 5 * time.Second

type DBhealthCheckImpl struct {
	pool *pgxpool.Pool
}

func NewDBhealthCheckImpl(pool *pgxpool.Pool) *DBhealthCheckImpl {
	return &DBhealthCheckImpl{pool: pool}
}

var _ usecase.DatabaseHealthChecker = (*DBhealthCheckImpl)(nil)

func (p *DBhealthCheckImpl) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()
	return p.pool.Ping(pingCtx)
}

func (p *DBhealthCheckImpl) Close() {
	p.pool.Close()
}

func (p *DBhealthCheckImpl) Pool() *pgxpool.Pool {
	return p.pool
}
