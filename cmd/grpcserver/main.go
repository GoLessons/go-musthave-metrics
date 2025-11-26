package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/GoLessons/go-musthave-metrics/internal/common/buildinfo"
	config "github.com/GoLessons/go-musthave-metrics/internal/server/config"
	container2 "github.com/GoLessons/go-musthave-metrics/internal/server/container"
	servergrpc "github.com/GoLessons/go-musthave-metrics/internal/server/grpc"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"go.uber.org/zap"
)

var buildVersion string
var buildDate string
var buildCommit string

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

	listener, err := net.Listen("tcp", cfg.GrpcAddress)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	server, err := servergrpc.BuildGRPCServer(c)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

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
