package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
	"go.uber.org/zap"
)

type MetricsPostgresRepository struct {
	db *sqlx.DB
}

type MetricsDB struct {
	Name  string   `db:"name"`
	Type  string   `db:"type"`
	Delta *int64   `db:"delta"`
	Value *float64 `db:"value"`
}

func NewMetricsPostgresRepository(db *sqlx.DB) *MetricsPostgresRepository {
	return &MetricsPostgresRepository{db: db}
}

func (m *MetricsPostgresRepository) GetAll() []*model.Metrics {
	ctx := context.Background()
	dbMetrics := make([]MetricsDB, 0)
	err := m.withTx(ctx, func(tx *sqlx.Tx) error {
		return sqlx.SelectContext(ctx, tx, &dbMetrics, "SELECT name, type, delta, value FROM metrics")
	})
	if err != nil {
		logger.Log.Error("failed to query metrics", zap.Error(err))
		return []*model.Metrics{}
	}
	result := make([]*model.Metrics, 0, len(dbMetrics))
	for _, metric := range dbMetrics {
		result = append(result, metric.toDomain())
	}
	return result
}

func (m *MetricsPostgresRepository) GetMetric(mType string, name string) *model.Metrics {
	ctx := context.Background()
	dbMetric := MetricsDB{}
	err := m.withTx(ctx, func(tx *sqlx.Tx) error {
		query, args, namedErr := sqlx.Named("SELECT name, type, delta, value FROM metrics WHERE type = :type AND name = :name", map[string]any{
			"type": mType,
			"name": name,
		})
		if namedErr != nil {
			return namedErr
		}
		query = m.db.Rebind(query)
		return sqlx.GetContext(ctx, tx, &dbMetric, query, args...)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		logger.Log.Error("failed to query metric", zap.Error(err))
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

	err := m.withTx(ctx, func(tx *sqlx.Tx) error {
		rowsUpdated, updateErr := m.updateMetric(ctx, tx, dbMetric)
		if updateErr != nil {
			return updateErr
		}
		if rowsUpdated > 0 {
			return nil
		}
		return m.insertMetric(ctx, tx, dbMetric)
	})
	if err != nil {
		logger.Log.Error("failed to persist metric", zap.Error(err))
	}
}

func (m *MetricsPostgresRepository) updateMetric(ctx context.Context, exec sqlx.ExtContext, metric MetricsDB) (int64, error) {
	result, err := sqlx.NamedExecContext(ctx, exec, "UPDATE metrics SET delta = :delta, value = :value WHERE name = :name AND type = :type", metric)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func (m *MetricsPostgresRepository) insertMetric(ctx context.Context, exec sqlx.ExtContext, metric MetricsDB) error {
	_, err := sqlx.NamedExecContext(ctx, exec, "INSERT INTO metrics (name, type, delta, value) VALUES (:name, :type, :delta, :value)", metric)
	return err
}

func (m *MetricsPostgresRepository) withTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				logger.Log.Error("failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func newMetricsDBFromDomain(metric *model.Metrics) MetricsDB {
	var delta *int64
	if metric.Delta != nil {
		value := *metric.Delta
		delta = &value
	}
	var valuePtr *float64
	if metric.Value != nil {
		val := *metric.Value
		valuePtr = &val
	}
	return MetricsDB{
		Name:  metric.ID,
		Type:  metric.MType,
		Delta: delta,
		Value: valuePtr,
	}
}

func (m MetricsDB) toDomain() *model.Metrics {
	metric := model.CreateMetric(m.Name, m.Type)
	if m.Delta != nil {
		value := *m.Delta
		metric.Delta = &value
	}
	if m.Value != nil {
		val := *m.Value
		metric.Value = &val
	}
	return metric
}

var _ repository.MetricRepository = (*MetricsPostgresRepository)(nil)
