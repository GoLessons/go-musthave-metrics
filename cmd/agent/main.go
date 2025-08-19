package main

import (
	"context"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/caarlos0/env"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

	metricsStorage := storage.NewMemStorage[model.Metrics]()
	readers := []agent.Reader{reader.NewRuntimeMetricsReader(), reader.NewSystemMetricsReader()}
	simpleReader := reader.NewSimpleMetricsReader()
	sender := createSender(cfg)
	defer sender.Close()

	pollDuration := time.Duration(cfg.PollInterval) * time.Second
	dumpInterval := time.Duration(cfg.ReportInterval) * time.Second

	collectAndSendMetrics(ctx, metricsStorage, readers, simpleReader, sender, pollDuration, dumpInterval, cfg.RateLimit, cfg.Batch)

	return nil
}

func collectAndSendMetrics(
	ctx context.Context,
	stg storage.Storage[model.Metrics],
	readers []agent.Reader,
	simpleReader *reader.SimpleMetricsReader,
	sender agent.Sender,
	pollDuration, dumpInterval time.Duration,
	rateLimit int,
	batch bool,
) {
	pollTicker := time.NewTicker(pollDuration)
	defer pollTicker.Stop()

	dumpTicker := time.NewTicker(dumpInterval)
	defer dumpTicker.Stop()

	sendChan := make(chan []model.Metrics, 1)
	var sendGroup sync.WaitGroup

	sendGroup.Add(1)
	go agent.SenderWorker(ctx, sendChan, sender, rateLimit, batch, &sendGroup)

	for {
		select {
		case <-ctx.Done():
			close(sendChan)
			sendGroup.Wait()
			return
		case <-pollTicker.C:
			collectAllMetrics(ctx, stg, readers, simpleReader)
		case <-dumpTicker.C:
			metrics, err := fetchAllMetrics(stg)
			if err != nil {
				fmt.Printf("can't fetch metrics: %v\n", err)
				continue
			}

			select {
			case sendChan <- metrics:
				simpleReader.Reset()
			case <-ctx.Done():
				close(sendChan)
				sendGroup.Wait()
				return
			}
		}
	}
}

func collectAllMetrics(
	ctx context.Context,
	stg storage.Storage[model.Metrics],
	readers []agent.Reader,
	simpleReader *reader.SimpleMetricsReader,
) {
	var wg sync.WaitGroup

	for _, r := range readers {
		wg.Add(1)
		go func(rd agent.Reader) {
			defer wg.Done()
			collectFromReader(stg, rd)
		}(r)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		collectFromSimpleReader(stg, simpleReader)
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

func collectFromReader(stg storage.Storage[model.Metrics], rd agent.Reader) {
	if err := rd.Refresh(); err != nil {
		fmt.Printf("can't refresh reader: %v\n", err)
		return
	}

	metrics, err := rd.Fetch()
	if err != nil {
		fmt.Printf("can't fetch metrics: %v\n", err)
		return
	}

	for _, metric := range metrics {
		if err := stg.Set(metric.ID, metric); err != nil {
			fmt.Printf("can't store metric %s: %v\n", metric.ID, err)
		}
	}
}

func collectFromSimpleReader(stg storage.Storage[model.Metrics], simpleReader *reader.SimpleMetricsReader) {
	if err := simpleReader.Refresh(); err != nil {
		fmt.Printf("can't refresh simple reader: %v\n", err)
		return
	}

	metrics, err := simpleReader.Fetch()
	if err != nil {
		fmt.Printf("can't fetch simple metrics: %v\n", err)
		return
	}

	for _, metric := range metrics {
		if err := stg.Set(metric.ID, metric); err != nil {
			fmt.Printf("can't store simple metric %s: %v\n", metric.ID, err)
		}
	}
}

func fetchAllMetrics(stg storage.Storage[model.Metrics]) ([]model.Metrics, error) {
	all, err := stg.GetAll()
	if err != nil {
		return nil, fmt.Errorf("can't get all metrics: %w", err)
	}

	metrics := make([]model.Metrics, 0, len(all))
	for _, metric := range all {
		metrics = append(metrics, metric)
	}

	return metrics, nil
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

func createSender(cfg *Config) agent.Sender {
	if cfg.Plain {
		return agent.NewMetricURLSender(cfg.Address)
	}

	var signer *signature.Signer
	if cfg.SecretKey != "" {
		signer = signature.NewSign(cfg.SecretKey)
	}
	return agent.NewJSONSender(cfg.Address, cfg.EnableGzip, signer)
}
