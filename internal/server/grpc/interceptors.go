package grpc

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/server/security"
	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(logger *zap.Logger) gogrpc.UnaryServerInterceptor {
	return func(contextInstance context.Context, requestInstance interface{}, infoInstance *gogrpc.UnaryServerInfo, handlerFunction gogrpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		responseInstance, err := handlerFunction(contextInstance, requestInstance)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Convert(err).Code()
		}
		if logger != nil {
			logger.Info(
				"grpc request",
				zap.String("method", infoInstance.FullMethod),
				zap.Duration("duration", time.Since(startTime)),
				zap.String("code", statusCode.String()),
			)
		}
		return responseInstance, err
	}
}

func TrustedSubnetInterceptor(trustedCIDR string, logger *zap.Logger) gogrpc.UnaryServerInterceptor {
	var cidrPrefix netip.Prefix
	var checkEnabled bool
	if trustedCIDR != "" {
		prefix, err := security.ParseTrustedCIDR(trustedCIDR)
		if err == nil {
			cidrPrefix = prefix
			checkEnabled = true
		}
	}
	return func(contextInstance context.Context, requestInstance interface{}, infoInstance *gogrpc.UnaryServerInfo, handlerFunction gogrpc.UnaryHandler) (interface{}, error) {
		if !checkEnabled {
			return handlerFunction(contextInstance, requestInstance)
		}
		metadataInstance, ok := metadata.FromIncomingContext(contextInstance)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		values := metadataInstance.Get("x-real-ip")
		ipString := ""
		if len(values) > 0 {
			ipString = strings.TrimSpace(values[0])
		}
		if ipString == "" {
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		trusted, err := security.IsIPTrusted(cidrPrefix, ipString)
		if err != nil {
			if logger != nil {
				logger.Warn("invalid X-Real-IP", zap.String("ip", ipString))
			}
			return nil, status.Error(codes.InvalidArgument, "Bad Request")
		}
		if !trusted {
			if logger != nil {
				logger.Warn("ip not in trusted subnet", zap.String("ip", ipString), zap.String("cidr", cidrPrefix.String()))
			}
			return nil, status.Error(codes.PermissionDenied, "Forbidden")
		}
		return handlerFunction(contextInstance, requestInstance)
	}
}
