package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	address        string
	reportInterval int
	pollInterval   int
)

func init() {
	rootCmd.Flags().StringVarP(&address, "address", "a", "localhost:8080", "HTTP server address")
	rootCmd.Flags().IntVarP(&reportInterval, "report", "r", 10, "Report interval in seconds")
	rootCmd.Flags().IntVarP(&pollInterval, "poll", "p", 2, "Poll interval in seconds")

	// Запрещаем использование неизвестных флагов
	rootCmd.FParseErrWhitelist.UnknownFlags = false
}

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Metrics agent for collecting and sending metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportInterval <= 0 {
			return fmt.Errorf("report interval must be positive, got %d", reportInterval)
		}
		if pollInterval <= 0 {
			return fmt.Errorf("poll interval must be positive, got %d", pollInterval)
		}

		run()
		return nil
	},
}

func run() {
	metricCollector := MetricCollectorFactory(address, reportInterval, pollInterval)
	defer metricCollector.Close()
	metricCollector.CollectAndSendMetrics()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
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
