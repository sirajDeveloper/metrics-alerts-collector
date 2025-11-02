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
	metrics, err := s.LoadAll()
	if err != nil {
		logger.Log.Error("file Save load error", zap.Error(err))
		return
	}
	found := false
	for i, m := range metrics {
		if m.ID == metric.ID && m.MType == metric.MType {
			metrics[i] = metric
			found = true
			break
		}
	}
	if !found {
		metrics = append(metrics, metric)
	}
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("file Save marshal error", zap.Error(err))
		return
	}
	if err := os.WriteFile(s.filePath, jsonData, 0644); err != nil {
		logger.Log.Error("file Save write error", zap.Error(err))
		return
	}
	logger.Log.Info("file Save done", zap.String("path", s.filePath), zap.Int("bytes", len(jsonData)))
}

func (s *JSONFileStorage) LoadAll() ([]*model.Metrics, error) {
	logger.Log.Info("file LoadAll start", zap.String("path", s.filePath))
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log.Info("file LoadAll file not found", zap.String("path", s.filePath))
			return []*model.Metrics{}, nil
		}
		logger.Log.Error("file LoadAll read error", zap.Error(err))
		return nil, err
	}
	if len(data) == 0 {
		logger.Log.Info("file LoadAll file empty", zap.String("path", s.filePath))
		return []*model.Metrics{}, nil
	}
	var metrics []*model.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		logger.Log.Error("file LoadAll unmarshal error", zap.Error(err))
		return nil, err
	}
	logger.Log.Info("file LoadAll done", zap.String("path", s.filePath), zap.Int("count", len(metrics)))
	return metrics, nil
}
