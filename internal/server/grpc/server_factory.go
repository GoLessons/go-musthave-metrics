package grpc

import (
	"github.com/GoLessons/go-musthave-metrics/internal/proto"
	"github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
)

func BuildGRPCServer(containerInstance container.Container) (*gogrpc.Server, error) {
	loggerInstance, err := container.GetService[zap.Logger](containerInstance, "logger")
	if err != nil {
		return nil, err
	}
	configInstance, err := container.GetService[config.Config](containerInstance, "config")
	if err != nil {
		return nil, err
	}
	metricServiceInstance, err := container.GetService[service.MetricService](containerInstance, "metricService")
	if err != nil {
		return nil, err
	}

	interceptorList := []gogrpc.UnaryServerInterceptor{LoggingInterceptor(loggerInstance)}
	if configInstance.TrustedSubnet != "" {
		interceptorList = append(interceptorList, TrustedSubnetInterceptor(configInstance.TrustedSubnet, loggerInstance))
	}

	var serverInstance *gogrpc.Server
	if len(interceptorList) == 1 {
		serverInstance = gogrpc.NewServer(gogrpc.UnaryInterceptor(interceptorList[0]))
	} else {
		serverInstance = gogrpc.NewServer(gogrpc.ChainUnaryInterceptor(interceptorList...))
	}

	proto.RegisterMetricsServer(serverInstance, NewMetricsGRPCService(metricServiceInstance))

	return serverInstance, nil
}
