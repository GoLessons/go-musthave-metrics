package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/caarlos0/env"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type Config struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	Plain          bool   `env:"PLAIN" envDefault:"false"`
	DisableGzip    bool   `env:"NO_GZIP" envDefault:"true"`
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

		run(cfg)
		return nil
	}

	rootCmd.FParseErrWhitelist.UnknownFlags = false

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *Config) {
	metricCollector := MetricCollectorFactory(cfg)
	defer metricCollector.Close()
	metricCollector.CollectAndSendMetrics()
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
	cmd.Flags().BoolVarP(&cfg.DisableGzip, "no-gzip", "", cfg.DisableGzip, "Disable gzip compression for JSON requests")

	return cfg, nil
}

func MetricCollectorFactory(cfg *Config) *agent.MetricCollector {
	var sender agent.Sender
	if cfg.Plain {
		sender = agent.NewMetricURLSender(cfg.Address)
	} else {
		sender = agent.NewJSONSender(cfg.Address, cfg.DisableGzip)
	}

	return agent.NewMetricCollector(
		storage.NewMemStorage[agent.GaugeValue](),
		storage.NewMemStorage[agent.CounterValue](),
		agent.NewMemStatsReader(),
		sender,
		time.Duration(cfg.ReportInterval)*time.Second,
		time.Duration(cfg.PollInterval)*time.Second,
	)
}
