package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

const defaultPingTimeout = 5 * time.Second

type DBhealthCheckImpl struct {
	db *sqlx.DB
}

func NewDBhealthCheckImpl(db *sqlx.DB) *DBhealthCheckImpl {
	return &DBhealthCheckImpl{db: db}
}

var _ usecase.DatabaseHealthChecker = (*DBhealthCheckImpl)(nil)

func (p *DBhealthCheckImpl) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()
	return p.db.PingContext(pingCtx)
}

func (p *DBhealthCheckImpl) Close() {
	p.db.Close()
}
