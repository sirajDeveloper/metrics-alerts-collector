package usecase

import (
	"testing"

	"errors"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.InitLogger(true)
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Log.Error("Error while logger.Sync", zap.Error(err))
		}
	}()
	m.Run()
}

type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) SaveAll(metrics []*model.Metrics) {
	m.Called(metrics)
}

func (m *MockFileStorage) Save(metric *model.Metrics) {
	m.Called(metric)
}

func (m *MockFileStorage) LoadAll() ([]*model.Metrics, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Metrics), args.Error(1)
}

func TestMetricsEmitterService_Send(t *testing.T) {
	tests := []struct {
		name           string
		reportInterval int
		expectSave     bool
	}{
		{
			name:           "reportInterval равен 0 - Save должен вызываться",
			reportInterval: 0,
			expectSave:     true,
		},
		{
			name:           "reportInterval не равен 0 - Save не должен вызываться",
			reportInterval: 300,
			expectSave:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileStorage := new(MockFileStorage)
			mockRepo := new(MockRepository)

			mockFileStorage.On("Save", mock.Anything).Return()

			service := NewMetricsEmitterService(mockFileStorage, mockRepo, tt.reportInterval)
			service.Send(event.MetricsEvent{Metrics: &model.Metrics{
				ID:    "test",
				MType: "gauge",
				Value: func() *float64 { v := 25.5; return &v }(),
			}})

			if tt.expectSave {
				mockFileStorage.AssertCalled(t, "Save", mock.Anything)
			} else {
				mockFileStorage.AssertNotCalled(t, "Save", mock.Anything)
			}
		})
	}
}

func TestMetricsEmitterService_EmitAll(t *testing.T) {
	metrics := []*model.Metrics{
		{
			ID:    "test",
			MType: "gauge",
			Value: func() *float64 { v := 25.5; return &v }(),
		},
		{
			ID:    "test2",
			MType: "counter",
			Delta: func() *int64 { v := int64(100); return &v }(),
		},
	}

	mockFileStorage := new(MockFileStorage)
	mockRepo := new(MockRepository)

	mockRepo.On("GetAll").Return(metrics)
	mockFileStorage.On("SaveAll", metrics).Return()

	service := NewMetricsEmitterService(mockFileStorage, mockRepo, 0)
	service.EmitAll()

	mockRepo.AssertCalled(t, "GetAll")
	mockFileStorage.AssertCalled(t, "SaveAll", metrics)
}

func TestMetricsEmitterService_RestoreAll(t *testing.T) {
	metrics := []*model.Metrics{
		{
			ID:    "test",
			MType: "gauge",
			Value: func() *float64 { v := 25.5; return &v }(),
		},
		{
			ID:    "test2",
			MType: "counter",
			Delta: func() *int64 { v := int64(100); return &v }(),
		},
	}

	mockFileStorage := new(MockFileStorage)
	mockRepo := new(MockRepository)

	mockFileStorage.On("LoadAll").Return(metrics, nil)
	for _, metric := range metrics {
		mockRepo.On("Save", metric).Return()
	}

	service := NewMetricsEmitterService(mockFileStorage, mockRepo, 0)
	err := service.RestoreAll()

	assert.NoError(t, err)
	mockFileStorage.AssertCalled(t, "LoadAll")
	for _, metric := range metrics {
		mockRepo.AssertCalled(t, "Save", metric)
	}
}

func TestMetricsEmitterService_RestoreAll_Error(t *testing.T) {
	mockFileStorage := new(MockFileStorage)
	mockRepo := new(MockRepository)

	expectedErr := errors.New("error")
	mockFileStorage.On("LoadAll").Return(nil, expectedErr)

	service := NewMetricsEmitterService(mockFileStorage, mockRepo, 0)
	err := service.RestoreAll()

	mockFileStorage.AssertCalled(t, "LoadAll")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything)
	assert.Equal(t, expectedErr, err)
}
