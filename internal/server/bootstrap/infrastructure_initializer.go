package bootstrap

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"

	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/database"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/file"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type InfrastructureInitializer struct {
	config Config
}

func NewInfrastructureInitializer(cfg Config) *InfrastructureInitializer {
	return &InfrastructureInitializer{
		config: cfg,
	}
}

type InfrastructureResult struct {
	MetricRepository repository.MetricRepository
	FileStorage      repository.MetricFileStorage
	HealthChecker    usecase.DatabaseHealthChecker
	DB               *sqlx.DB
}

func (i *InfrastructureInitializer) Initialize() (*InfrastructureResult, error) {
	fileStorage := file.NewJSONFileStorage(*i.config.GetFileStoragePath())

	if i.config.GetDatabaseDSN() != nil {
		result, err := i.initDatabaseStorage()
		if err != nil {
			logger.Log.Warn("Failed to initialize database, falling back to memory storage", zap.Error(err))
			return i.initMemoryStorage(fileStorage), nil
		}
		result.FileStorage = fileStorage
		return result, nil
	}

	return i.initMemoryStorage(fileStorage), nil
}

func (i *InfrastructureInitializer) initDatabaseStorage() (*InfrastructureResult, error) {
	if i.config.GetMigrationsPath() != nil {
		migrationRunner, err := database.NewMigrationRunner(*i.config.GetMigrationsPath(), *i.config.GetDatabaseDSN())
		if err != nil {
			return nil, err
		}
		if err := migrationRunner.Up(context.Background()); err != nil {
			return nil, err
		}
	}

	dbCtx, cancelDB := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelDB()

	db, err := sqlx.ConnectContext(dbCtx, "pgx", *i.config.GetDatabaseDSN())
	if err != nil {
		return nil, err
	}

	retryCount := 3
	if i.config.GetCountRetrySave() != nil {
		retryCount = *i.config.GetCountRetrySave()
	}

	return &InfrastructureResult{
		MetricRepository: database.NewMetricsPostgresRepository(db, retryCount),
		HealthChecker:    database.NewDBhealthCheckImpl(db),
		DB:               db,
	}, nil
}

func (i *InfrastructureInitializer) initMemoryStorage(fileStorage *file.JSONFileStorage) *InfrastructureResult {
	return &InfrastructureResult{
		MetricRepository: cache.NewMemStorage(),
		FileStorage:      fileStorage,
		HealthChecker:    nil,
		DB:               nil,
	}
}
