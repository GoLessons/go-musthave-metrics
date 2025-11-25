package grpc

import (
	"context"
	"testing"

	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestLoggingInterceptor_PassesThrough(t *testing.T) {
	interceptorInstance := LoggingInterceptor(zap.NewNop())
	handlerFunction := func(contextInstance context.Context, requestInstance interface{}) (interface{}, error) {
		return "ok", nil
	}
	infoInstance := &gogrpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"}
	responseInstance, err := interceptorInstance(context.Background(), "req", infoInstance, handlerFunction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseInstance != "ok" {
		t.Fatalf("unexpected response: %v", responseInstance)
	}
}

func TestTrustedSubnetInterceptor_EmptyCIDR_Passes(t *testing.T) {
	interceptorInstance := TrustedSubnetInterceptor("", zap.NewNop())
	handlerFunction := func(contextInstance context.Context, requestInstance interface{}) (interface{}, error) {
		return "ok", nil
	}
	infoInstance := &gogrpc.UnaryServerInfo{FullMethod: "/x"}
	responseInstance, err := interceptorInstance(context.Background(), "req", infoInstance, handlerFunction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseInstance != "ok" {
		t.Fatalf("unexpected response: %v", responseInstance)
	}
}

func TestTrustedSubnetInterceptor_TrustedIP_Passes(t *testing.T) {
	interceptorInstance := TrustedSubnetInterceptor("127.0.0.0/8", zap.NewNop())
	handlerFunction := func(contextInstance context.Context, requestInstance interface{}) (interface{}, error) {
		return "ok", nil
	}
	infoInstance := &gogrpc.UnaryServerInfo{FullMethod: "/x"}
	metadataContext := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "127.0.0.1"))
	responseInstance, err := interceptorInstance(metadataContext, "req", infoInstance, handlerFunction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseInstance != "ok" {
		t.Fatalf("unexpected response: %v", responseInstance)
	}
}

func TestTrustedSubnetInterceptor_MissingIP_ReturnsInvalidArgument(t *testing.T) {
	interceptorInstance := TrustedSubnetInterceptor("127.0.0.0/8", zap.NewNop())
	handlerFunction := func(contextInstance context.Context, requestInstance interface{}) (interface{}, error) {
		return "ok", nil
	}
	infoInstance := &gogrpc.UnaryServerInfo{FullMethod: "/x"}
	responseInstance, err := interceptorInstance(context.Background(), "req", infoInstance, handlerFunction)
	if err == nil {
		t.Fatalf("expected error")
	}
	if responseInstance != nil {
		t.Fatalf("unexpected response: %v", responseInstance)
	}
}

func TestTrustedSubnetInterceptor_UntrustedIP_ReturnsPermissionDenied(t *testing.T) {
	interceptorInstance := TrustedSubnetInterceptor("127.0.0.0/8", zap.NewNop())
	handlerFunction := func(contextInstance context.Context, requestInstance interface{}) (interface{}, error) {
		return "ok", nil
	}
	infoInstance := &gogrpc.UnaryServerInfo{FullMethod: "/x"}
	metadataContext := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "10.0.0.1"))
	responseInstance, err := interceptorInstance(metadataContext, "req", infoInstance, handlerFunction)
	if err == nil {
		t.Fatalf("expected error")
	}
	if responseInstance != nil {
		t.Fatalf("unexpected response: %v", responseInstance)
	}
}
