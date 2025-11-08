package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
	"go.uber.org/zap"
)

type MetricsPostgresRepository struct {
	pool *pgxpool.Pool
}

type MetricsDB struct {
	ID    int64
	Name  string
	Type  string
	Delta sql.NullInt64
	Value sql.NullFloat64
}

func NewMetricsPostgresRepository(pool *pgxpool.Pool) *MetricsPostgresRepository {
	return &MetricsPostgresRepository{pool: pool}
}

func (m *MetricsPostgresRepository) GetAll() []*model.Metrics {
	rows, err := m.pool.Query(context.Background(), "SELECT name, type, delta, value FROM metrics")
	if err != nil {
		logger.Log.Error("failed to query metrics", zap.Error(err))
		return []*model.Metrics{}
	}
	defer rows.Close()

	result := make([]*model.Metrics, 0)
	for rows.Next() {
		dbMetric := MetricsDB{}
		if err := rows.Scan(&dbMetric.Name, &dbMetric.Type, &dbMetric.Delta, &dbMetric.Value); err != nil {
			logger.Log.Error("failed to scan metric row", zap.Error(err))
			continue
		}
		result = append(result, dbMetric.toDomain())
	}

	if err := rows.Err(); err != nil {
		logger.Log.Error("metrics rows iteration error", zap.Error(err))
	}

	return result
}

func (m *MetricsPostgresRepository) GetMetric(mType string, name string) *model.Metrics {
	row := m.pool.QueryRow(context.Background(), "SELECT name, type, delta, value FROM metrics WHERE type = $1 AND name = $2", mType, name)
	dbMetric := MetricsDB{}
	if err := row.Scan(&dbMetric.Name, &dbMetric.Type, &dbMetric.Delta, &dbMetric.Value); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		logger.Log.Error("failed to scan metric", zap.Error(err))
		return nil
	}

	return dbMetric.toDomain()
}

func (m *MetricsPostgresRepository) Save(metric *model.Metrics) {
	if metric == nil {
		return
	}

	dbMetric := newMetricsDBFromDomain(metric)
	ctx := context.Background()

	tag, err := m.updateMetric(ctx, dbMetric)
	if err != nil {
		logger.Log.Error("failed to update metric", zap.Error(err))
		return
	}

	if tag.RowsAffected() > 0 {
		return
	}

	err = m.insertMetric(ctx, dbMetric)
	if err != nil {
		logger.Log.Error("failed to insert metric", zap.Error(err))
	}
}

func (m *MetricsPostgresRepository) updateMetric(ctx context.Context, metric MetricsDB) (pgconn.CommandTag, error) {
	return m.pool.Exec(ctx, "UPDATE metrics SET delta = $3, value = $4 WHERE name = $1 AND type = $2", metric.Name, metric.Type, metric.Delta, metric.Value)
}

func (m *MetricsPostgresRepository) insertMetric(ctx context.Context, metric MetricsDB) error {
	_, err := m.pool.Exec(ctx, "INSERT INTO metrics (name, type, delta, value) VALUES ($1, $2, $3, $4)", metric.Name, metric.Type, metric.Delta, metric.Value)
	return err
}

func newMetricsDBFromDomain(metric *model.Metrics) MetricsDB {
	delta := sql.NullInt64{}
	if metric.Delta != nil {
		delta.Int64 = *metric.Delta
		delta.Valid = true
	}

	value := sql.NullFloat64{}
	if metric.Value != nil {
		value.Float64 = *metric.Value
		value.Valid = true
	}

	return MetricsDB{
		Name:  metric.ID,
		Type:  metric.MType,
		Delta: delta,
		Value: value,
	}
}

func (m MetricsDB) toDomain() *model.Metrics {
	metric := model.CreateMetric(m.Name, m.Type)
	if m.Delta.Valid {
		value := m.Delta.Int64
		metric.Delta = &value
	}
	if m.Value.Valid {
		val := m.Value.Float64
		metric.Value = &val
	}
	return metric
}

var _ repository.MetricRepository = (*MetricsPostgresRepository)(nil)
