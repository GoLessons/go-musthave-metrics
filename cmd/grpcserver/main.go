package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GoLessons/go-musthave-metrics/internal/common/buildinfo"
	"github.com/GoLessons/go-musthave-metrics/internal/common/netaddr"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/proto"
	config "github.com/GoLessons/go-musthave-metrics/internal/server/config"
	container2 "github.com/GoLessons/go-musthave-metrics/internal/server/container"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var buildVersion string
var buildDate string
var buildCommit string

type metricsServer struct {
	proto.UnimplementedMetricsServer
	ms *service.MetricService
}

func newMetricsServer(ms *service.MetricService) *metricsServer {
	return &metricsServer{ms: ms}
}

func (s *metricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	for _, pm := range req.Metrics {
		m, err := protoToModel(pm)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		if err := s.ms.Save(m); err != nil {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
	}
	return &proto.UpdateMetricsResponse{}, nil
}

func protoToModel(pm *proto.Metric) (model.Metrics, error) {
	var m model.Metrics
	m.ID = pm.Id
	switch pm.Type {
	case proto.Metric_GAUGE:
		m.MType = model.Gauge
		v := pm.Value
		m.Value = &v
	case proto.Metric_COUNTER:
		m.MType = model.Counter
		d := pm.Delta
		m.Delta = &d
	default:
		return m, fmt.Errorf("unknown metric type")
	}
	return m, nil
}

func trustedSubnetInterceptor(cidr string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	var prefix netip.Prefix
	var enabled bool
	if cidr != "" {
		p, err := netip.ParsePrefix(cidr)
		if err == nil {
			prefix = p
			enabled = true
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !enabled {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		vals := md.Get("x-real-ip")
		ipStr := ""
		if len(vals) > 0 {
			ipStr = strings.TrimSpace(vals[0])
		}
		if ipStr == "" {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		addr, err := netaddr.ParseAddr(ipStr)
		if err != nil {
			if logger != nil {
				logger.Warn("invalid X-Real-IP", zap.String("ip", ipStr))
			}
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		if !prefix.Contains(addr) {
			if logger != nil {
				logger.Warn("ip not in trusted subnet", zap.String("ip", ipStr), zap.String("cidr", prefix.String()))
			}
			return nil, status.Error(codes.PermissionDenied, "Forbidden")
		}
		return handler(ctx, req)
	}
}

func main() {
	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	c, err := container2.InitContainer()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	serverLogger, err := container.GetService[zap.Logger](c, "logger")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	cfg, err := container.GetService[config.Config](c, "config")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	metricService, err := container.GetService[service.MetricService](c, "metricService")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", cfg.GrpcAddress)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	server := grpc.NewServer(grpc.UnaryInterceptor(trustedSubnetInterceptor(cfg.TrustedSubnet, serverLogger)))
	proto.RegisterMetricsServer(server, newMetricsServer(metricService))

	serverLogger.Info("grpc server listening", zap.String("address", listener.Addr().String()))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := server.Serve(listener); err != nil {
			serverLogger.Error("grpc server error", zap.Error(err))
		}
	}()

	<-quit
	server.GracefulStop()
	serverLogger.Info("grpc server stopped")
}
