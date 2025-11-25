package grpc

import (
	"context"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/proto"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsGRPCService_UpdateMetrics_Success(t *testing.T) {
	counterStorage := storage.NewMemStorage[serverModel.Counter]()
	gaugeStorage := storage.NewMemStorage[serverModel.Gauge]()
	metricService := service.NewMetricService(counterStorage, gaugeStorage)
	serviceInstance := NewMetricsGRPCService(metricService)

	requestInstance := &proto.UpdateMetricsRequest{
		Metrics: []*proto.Metric{
			{Id: "g1", Type: proto.Metric_GAUGE, Value: 1.23},
			{Id: "c1", Type: proto.Metric_COUNTER, Delta: 5},
		},
	}
	responseInstance, err := serviceInstance.UpdateMetrics(context.Background(), requestInstance)
	require.NoError(t, err)
	require.NotNil(t, responseInstance)

	gaugeValue, err := gaugeStorage.Get("g1")
	require.NoError(t, err)
	assert.Equal(t, 1.23, gaugeValue.Value())

	counterValue, err := counterStorage.Get("c1")
	require.NoError(t, err)
	assert.Equal(t, int64(5), counterValue.Value())
}

func TestMetricsGRPCService_UpdateMetrics_InvalidType_ReturnsInvalidArgument(t *testing.T) {
	counterStorage := storage.NewMemStorage[serverModel.Counter]()
	gaugeStorage := storage.NewMemStorage[serverModel.Gauge]()
	metricService := service.NewMetricService(counterStorage, gaugeStorage)
	serviceInstance := NewMetricsGRPCService(metricService)

	requestInstance := &proto.UpdateMetricsRequest{
		Metrics: []*proto.Metric{
			{Id: "bad", Type: proto.Metric_MType(100)},
		},
	}
	responseInstance, err := serviceInstance.UpdateMetrics(context.Background(), requestInstance)
	require.Error(t, err)
	assert.Nil(t, responseInstance)
}
