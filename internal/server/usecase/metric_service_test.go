package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetMetric(metricType, metricName string) *model.Metrics {
	args := m.Called(metricType, metricName)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*model.Metrics)
}

func (m *MockRepository) Save(metrics *model.Metrics) {
	m.Called(metrics)
}

func (m *MockRepository) GetAll() []*model.Metrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*model.Metrics)
}

func TestMetricService_MetricUpdate(t *testing.T) {
	tests := []struct {
		name          string
		metricType    string
		metricName    string
		metricValue   string
		setupMock     func(repo *MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:        "создание новой gauge метрики с валидным значением",
			metricType:  "gauge",
			metricName:  "temperature",
			metricValue: "25.5",
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "gauge", "temperature").Return(nil)

				repo.On("Save", mock.MatchedBy(func(m *model.Metrics) bool {
					return m.ID == "temperature" &&
						m.MType == "gauge" &&
						m.Value != nil &&
						*m.Value == 25.5
				})).Return()
			},
			expectedError: false,
		},
		{
			name:        "создание новой counter метрики с валидным значением",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "100",
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "counter", "requests").Return(nil)

				repo.On("Save", mock.MatchedBy(func(m *model.Metrics) bool {
					return m.ID == "requests" &&
						m.MType == "counter" &&
						m.Delta != nil &&
						*m.Delta == int64(100)
				})).Return()
			},
			expectedError: false,
		},
		{
			name:        "обновление существующей gauge метрики",
			metricType:  "gauge",
			metricName:  "temperature",
			metricValue: "30.0",
			setupMock: func(repo *MockRepository) {
				existingMetric := &model.Metrics{
					ID:    "temperature",
					MType: "gauge",
				}
				oldValue := 25.5
				existingMetric.Value = &oldValue

				repo.On("GetMetric", "gauge", "temperature").Return(existingMetric)

				repo.On("Save", mock.MatchedBy(func(m *model.Metrics) bool {
					return m.ID == "temperature" &&
						m.Value != nil &&
						*m.Value == 30.0
				})).Return()
			},
			expectedError: false,
		},
		{
			name:        "обновление существующей counter метрики (суммирование)",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "50",
			setupMock: func(repo *MockRepository) {
				existingMetric := &model.Metrics{
					ID:    "requests",
					MType: "counter",
				}
				oldDelta := int64(100)
				existingMetric.Delta = &oldDelta

				repo.On("GetMetric", "counter", "requests").Return(existingMetric)

				repo.On("Save", mock.MatchedBy(func(m *model.Metrics) bool {
					return m.ID == "requests" &&
						m.Delta != nil &&
						*m.Delta == int64(150)
				})).Return()
			},
			expectedError: false,
		},
		{
			name:        "невалидное значение для gauge (не число)",
			metricType:  "gauge",
			metricName:  "temperature",
			metricValue: "invalid",
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "gauge", "temperature").Return(nil)
			},
			expectedError: true,
			errorContains: "invalid float64 value",
		},
		{
			name:        "невалидное значение для counter (не целое число)",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "12.34",
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "counter", "requests").Return(nil)
			},
			expectedError: true,
			errorContains: "invalid int64 value",
		},
		{
			name:        "неизвестный тип метрики",
			metricType:  "unknown",
			metricName:  "test",
			metricValue: "123",
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "unknown", "test").Return(nil)
			},
			expectedError: true,
			errorContains: "invalid metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			service := NewMetricService(mockRepo)

			err := service.MetricUpdate(tt.metricType, tt.metricName, tt.metricValue)

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка, но её не было")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains,
						"Ошибка должна содержать ожидаемый текст")
				}
			} else {
				assert.NoError(t, err, "Не ожидалась ошибка, но она произошла")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNewMetricService(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewMetricService(mockRepo)

	assert.NotNil(t, service, "Сервис не должен быть nil")
	assert.Implements(t, (*MetricUpdater)(nil), service, "Сервис должен реализовывать интерфейс MetricUpdater")
}
