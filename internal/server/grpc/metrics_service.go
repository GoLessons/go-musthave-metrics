package grpc

import (
	"context"

	"github.com/GoLessons/go-musthave-metrics/internal/proto"
	"github.com/GoLessons/go-musthave-metrics/internal/proto/convert"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsGRPCService struct {
	proto.UnimplementedMetricsServer
	metricService *service.MetricService
}

func NewMetricsGRPCService(metricService *service.MetricService) *MetricsGRPCService {
	return &MetricsGRPCService{metricService: metricService}
}

func (serviceInstance *MetricsGRPCService) UpdateMetrics(contextInstance context.Context, requestInstance *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	for _, protoMetric := range requestInstance.Metrics {
		metric, err := convert.ProtoToModel(protoMetric)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		if err := serviceInstance.metricService.Save(metric); err != nil {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
	}
	return &proto.UpdateMetricsResponse{}, nil
}
