package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GrpcMetricSender struct {
	client         proto.MetricsClient
	retryCount     int
	requestTimeout time.Duration
	agentIP        string
}

func NewMetricsClient(client proto.MetricsClient, retryCount int, requestTimeout time.Duration, agentIP string) *GrpcMetricSender {
	return &GrpcMetricSender{
		client:         client,
		retryCount:     retryCount,
		requestTimeout: requestTimeout,
		agentIP:        agentIP,
	}
}

func toProto(metric domain.Metric) (proto.Metric_MType, error) {
	switch metric.Type {
	case domain.Counter:
		return proto.Metric_COUNTER, nil
	case domain.Gauge:
		return proto.Metric_GAUGE, nil
	}
	return 0, fmt.Errorf("unknown metric type: %s", metric.Type)
}

func (s *GrpcMetricSender) sendUpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) error {
	return ExecuteWithRetry(s.retryCount, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, s.requestTimeout)
		defer cancel()

		_, err := s.client.UpdateMetrics(requestCtx, req)
		if err == nil {
			return nil
		}

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
			logger.Log.Warn("transient gRPC error, will retry", zap.String("code", st.Code().String()), zap.Error(err))
			return err
		default:
			logger.Log.Warn("non-retriable gRPC error", zap.String("code", st.Code().String()), zap.Error(err))
			return nil
		}
	})
}

var _ usecase.MetricBatchSender = (*GrpcMetricSender)(nil)

func (s *GrpcMetricSender) SendBatch(metrics []domain.Metric) {
	if len(metrics) == 0 {
		return
	}
	pbMetrics := make([]*proto.Metric, 0, len(metrics))
	for i := range metrics {
		m := metrics[i]
		pbType, err := toProto(m)
		if err != nil {
			logger.Log.Warn("invalid metric type", zap.String("metric", m.Name), zap.Error(err))
			continue
		}

		pbMetric := &proto.Metric_builder{
			Id:   m.Name,
			Type: pbType,
		}

		switch m.Type {
		case domain.Counter:
			val, ok := m.Value.(int64)
			if !ok {
				logger.Log.Error("invalid counter value type", zap.String("metric", m.Name))
				continue
			}
			pbMetric.Delta = val
		case domain.Gauge:
			val, ok := m.Value.(float64)
			if !ok {
				logger.Log.Error("invalid gauge value type", zap.String("metric", m.Name))
				continue
			}
			pbMetric.Value = val
		default:
			logger.Log.Error("unsupported metric type", zap.String("metric", m.Name))
			continue
		}

		pbMetrics = append(pbMetrics, pbMetric.Build())
	}

	if len(pbMetrics) == 0 {
		return
	}

	req := &proto.UpdateMetricsRequest_builder{
		Metrics: pbMetrics,
	}

	md := metadata.New(map[string]string{
		"x-real-ip": s.agentIP,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	if err := s.sendUpdateMetrics(ctx, req.Build()); err != nil {
		logger.Log.Warn("failed to send metrics batch via gRPC", zap.Error(err))
	}
}
