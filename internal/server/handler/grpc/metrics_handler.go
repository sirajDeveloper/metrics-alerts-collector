package grpchandler

import (
	"context"
	"fmt"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/proto"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricsHandler struct {
	proto.UnimplementedMetricsServer
	metricUpdater usecase.MetricUpdater
}

func NewMetricsHandler(metricUpdater usecase.MetricUpdater) *MetricsHandler {
	return &MetricsHandler{
		metricUpdater: metricUpdater,
	}
}

func (h *MetricsHandler) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	metrics := req.GetMetrics()
	if len(metrics) == 0 {
		return &proto.UpdateMetricsResponse{}, nil
	}

	md, _ := metadata.FromIncomingContext(ctx)
	ipAddress := ""
	if md != nil {
		values := md.Get("x-real-ip")
		if len(values) > 0 {
			ipAddress = values[0]
		}
	}

	for _, m := range metrics {
		if m == nil {
			continue
		}

		metricType, err := mapProtoTypeToString(m.GetType())
		if err != nil {
			logger.Log.Error("invalid metric type in gRPC request", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		updateReq := dto.MetricUpdateRequest{
			ID:        m.GetId(),
			MType:     metricType,
			Delta:     nil,
			Value:     nil,
			IPAddress: ipAddress,
		}

		switch metricType {
		case "counter":
			value := m.GetDelta()
			updateReq.Delta = &value
		case "gauge":
			value := m.GetValue()
			updateReq.Value = &value
		}

		if err := h.metricUpdater.MetricUpdate(&updateReq); err != nil {
			logger.Log.Error("failed to update metric from gRPC request", zap.Any("request", updateReq), zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func mapProtoTypeToString(t proto.Metric_MType) (string, error) {
	switch t {
	case proto.Metric_COUNTER:
		return "counter", nil
	case proto.Metric_GAUGE:
		return "gauge", nil
	default:
		return "", fmt.Errorf("unknown metric type: %v", t)
	}
}
