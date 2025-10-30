package file

import (
	"encoding/json"
	"os"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type JSONFileStorage struct {
	filePath string
}

var _ repository.MetricFileStorage = (*JSONFileStorage)(nil)

func NewJSONFileStorage(filePath string) *JSONFileStorage {
	return &JSONFileStorage{filePath: filePath}
}

func (s *JSONFileStorage) SaveAll(metrics []*model.Metrics) {
	logger.Log.Info("file SaveAll start", zap.String("path", s.filePath), zap.Int("count", len(metrics)))
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("file SaveAll marshal error", zap.Error(err))
		return
	}
	if err := os.WriteFile(s.filePath, jsonData, 0644); err != nil {
		logger.Log.Error("file SaveAll write error", zap.Error(err))
		return
	}
	logger.Log.Info("file SaveAll done", zap.String("path", s.filePath), zap.Int("bytes", len(jsonData)))
}

func (s *JSONFileStorage) Save(metric *model.Metrics) {
	logger.Log.Info("file Save start", zap.String("path", s.filePath), zap.String("id", metric.ID), zap.String("type", metric.MType))
	jsonData, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error("file Save marshal error", zap.Error(err))
		return
	}
	f, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Log.Error("file Save open error", zap.Error(err))
		return
	}
	defer f.Close()
	if _, err := f.Write(append(jsonData, '\n')); err != nil {
		logger.Log.Error("file Save write error", zap.Error(err))
		return
	}
	logger.Log.Info("file Save done", zap.String("path", s.filePath), zap.Int("bytes", len(jsonData)))
}
