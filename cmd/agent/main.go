package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/caarlos0/env"
	"github.com/spf13/cobra"
)

type Config struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	Plain          bool   `env:"PLAIN" envDefault:"false"`
	EnableGzip     bool   `env:"GZIP" envDefault:"false"`
	Batch          bool   `env:"BATCH" envDefault:"false"`
	SecretKey      string `env:"KEY" envDefault:""`
	RateLimit      int    `env:"RATE_LIMIT" envDefault:"0"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "Metrics agent for collecting and sending metrics",
	}

	cfg, err := loadConfig(rootCmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cfg.ReportInterval <= 0 {
			return fmt.Errorf("report interval must be positive, got %d", cfg.ReportInterval)
		}
		if cfg.PollInterval <= 0 {
			return fmt.Errorf("poll interval must be positive, got %d", cfg.PollInterval)
		}

		return run(cfg)
	}

	rootCmd.FParseErrWhitelist.UnknownFlags = false

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Получен сигнал завершения, завершаем работу...")
		cancel()
	}()

	metricCollector := MetricCollectorFactory(cfg)
	defer metricCollector.Close()

	metricCollector.CollectAndSendMetrics(ctx, cfg.Batch)

	return nil
}

func loadConfig(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&cfg.Address, "address", "a", cfg.Address, "HTTP server address")
	cmd.Flags().IntVarP(&cfg.ReportInterval, "report", "r", cfg.ReportInterval, "Report interval in seconds")
	cmd.Flags().IntVarP(&cfg.PollInterval, "poll", "p", cfg.PollInterval, "Poll interval in seconds")
	cmd.Flags().BoolVarP(&cfg.Plain, "plain", "", cfg.Plain, "Use plain text format instead of JSON")
	cmd.Flags().BoolVarP(&cfg.EnableGzip, "gzip", "", cfg.EnableGzip, "Disable gzip compression for JSON requests")
	cmd.Flags().BoolVarP(&cfg.Batch, "batch", "b", cfg.Batch, "Send metrics in batch mode")
	cmd.Flags().StringVarP(&cfg.SecretKey, "key", "k", cfg.SecretKey, "SecretKey for signing metrics")
	cmd.Flags().IntVarP(&cfg.RateLimit, "rate-limit", "l", cfg.RateLimit, "Rate limit for sending metrics")

	return cfg, nil
}

func MetricCollectorFactory(cfg *Config) *agent.MetricCollector {
	var sender agent.Sender

	if cfg.Plain {
		sender = agent.NewMetricURLSender(cfg.Address)
	} else {
		var signer *signature.Signer
		if cfg.SecretKey != "" {
			signer = signature.NewSign(cfg.SecretKey)
		}
		sender = agent.NewJSONSender(cfg.Address, cfg.EnableGzip, signer)
	}

	return agent.NewMetricCollector(
		storage.NewMemStorage[model.Metrics](),
		[]agent.Reader{reader.NewRuntimeMetricsReader(), reader.NewSystemMetricsReader()},
		reader.NewSimpleMetricsReader(),
		sender,
		time.Duration(cfg.ReportInterval)*time.Second,
		time.Duration(cfg.PollInterval)*time.Second,
		cfg.RateLimit,
	)
}
