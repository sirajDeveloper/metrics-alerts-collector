package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

type MockRepository struct {
	mock.Mock
}

type MockSender struct {
	mock.Mock
}

func (m *MockSender) Send(e event.MetricsEvent) {
	m.Called(e)
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
		request       *dto.MetricUpdateRequest
		setupMock     func(repo *MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "создание новой gauge метрики с валидным значением",
			request: &dto.MetricUpdateRequest{
				ID:    "temperature",
				MType: "gauge",
				Value: func() *float64 { v := 25.5; return &v }(),
			},
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
			name: "создание новой counter метрики с валидным значением",
			request: &dto.MetricUpdateRequest{
				ID:    "requests",
				MType: "counter",
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
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
			name: "обновление существующей gauge метрики",
			request: &dto.MetricUpdateRequest{
				ID:    "temperature",
				MType: "gauge",
				Value: func() *float64 { v := 30.0; return &v }(),
			},
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
			name: "обновление существующей counter метрики (суммирование)",
			request: &dto.MetricUpdateRequest{
				ID:    "requests",
				MType: "counter",
				Delta: func() *int64 { v := int64(50); return &v }(),
			},
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
			name: "gauge без значения",
			request: &dto.MetricUpdateRequest{
				ID:    "temperature",
				MType: "gauge",
				Value: nil,
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "gauge", "temperature").Return(nil)
			},
			expectedError: true,
			errorContains: "gauge value is required",
		},
		{
			name: "counter без значения",
			request: &dto.MetricUpdateRequest{
				ID:    "requests",
				MType: "counter",
				Delta: nil,
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "counter", "requests").Return(nil)
			},
			expectedError: true,
			errorContains: "counter delta is required",
		},
		{
			name: "неизвестный тип метрики",
			request: &dto.MetricUpdateRequest{
				ID:    "test",
				MType: "unknown",
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetMetric", "unknown", "test").Return(nil)
			},
			expectedError: true,
			errorContains: "unknown metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockSender := new(MockSender)
			tt.setupMock(mockRepo)

			if !tt.expectedError {
				mockSender.On("Send", mock.Anything).Return()
			}

			service := NewMetricService(mockRepo, mockSender)

			err := service.MetricUpdate(tt.request)

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
			if !tt.expectedError {
				mockSender.AssertExpectations(t)
			}
		})
	}
}

func TestNewMetricService(t *testing.T) {
	mockRepo := new(MockRepository)
	mockSender := new(MockSender)
	service := NewMetricService(mockRepo, mockSender)

	assert.NotNil(t, service, "Сервис не должен быть nil")
	assert.Implements(t, (*MetricUpdater)(nil), service, "Сервис должен реализовывать интерфейс MetricUpdater")
}
