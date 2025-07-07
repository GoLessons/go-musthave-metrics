package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type Config struct {
	Address        string
	ReportInterval int
	PollInterval   int
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "Metrics agent for collecting and sending metrics",
	}

	cfg := loadConfig(rootCmd)
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
	metricCollector := MetricCollectorFactory(cfg.Address, cfg.ReportInterval, cfg.PollInterval)
	defer metricCollector.Close()
	metricCollector.CollectAndSendMetrics()
}

func loadConfig(cmd *cobra.Command) *Config {
	cfg := &Config{}

	cmd.Flags().StringVarP(&cfg.Address, "address", "a", "localhost:8080", "HTTP server address")
	cmd.Flags().IntVarP(&cfg.ReportInterval, "report", "r", 10, "Report interval in seconds")
	cmd.Flags().IntVarP(&cfg.PollInterval, "poll", "p", 2, "Poll interval in seconds")

	return cfg
}

func MetricCollectorFactory(address string, reportInterval, pollInterval int) *agent.MetricCollector {
	return agent.NewMetricCollector(
		storage.NewMemStorage[agent.GaugeValue](),
		storage.NewMemStorage[agent.CounterValue](),
		agent.NewMemStatsReader(),
		agent.NewMetricSender(address),
		time.Duration(reportInterval)*time.Second,
		time.Duration(pollInterval)*time.Second,
	)
}
