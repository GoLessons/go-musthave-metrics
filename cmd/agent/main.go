package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strconv"
	"strings"

	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/collector"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/buildinfo"
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	fileconfig "github.com/GoLessons/go-musthave-metrics/pkg/file-config"
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
	CryptoKey      string `env:"CRYPTO_KEY" envDefault:""`
}

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	var rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "Metrics agent for collecting and sending metrics",
	}

	cfg, err := loadConfig(rootCmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return overrideWithEnv(cfg)
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
	sender, err := createSender(cfg)
	if err != nil {
		return err
	}
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

	sendChan, stopSender := collector.StartSenderPipeline(ctx, sender, rateLimit, batch, 1)

	collector.RunAgentLoop(ctx, pollTicker, dumpTicker, stg, readers, simpleReader, sendChan, stopSender)
}

func loadConfig(cmd *cobra.Command) (*Config, error) {
	envAddress := os.Getenv("ADDRESS")
	if envAddress == "" {
		envAddress = "localhost:8080"
	}

	defaults := &Config{
		Address:        envAddress,
		ReportInterval: 10,
		PollInterval:   2,
		Plain:          false,
		EnableGzip:     false,
		Batch:          false,
		SecretKey:      "",
		RateLimit:      0,
		CryptoKey:      "",
	}

	if configPath := getFileConfigPath(); configPath != "" {
		if err := fileconfig.LoadInto(configPath, defaults); err != nil {
			return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
		}
	}

	cfg := &Config{}

	cmd.Flags().StringP("config", "c", "", "Path to agent config JSON")
	cmd.Flags().StringVarP(&cfg.Address, "address", "a", defaults.Address, "HTTP server address")
	cmd.Flags().IntVarP(&cfg.ReportInterval, "report", "r", defaults.ReportInterval, "Report interval in seconds")
	cmd.Flags().IntVarP(&cfg.PollInterval, "poll", "p", defaults.PollInterval, "Poll interval in seconds")
	cmd.Flags().BoolVarP(&cfg.Plain, "plain", "", defaults.Plain, "Use plain text format instead of JSON")
	cmd.Flags().BoolVarP(&cfg.EnableGzip, "gzip", "", defaults.EnableGzip, "Disable gzip compression for JSON requests")
	cmd.Flags().BoolVarP(&cfg.Batch, "batch", "b", defaults.Batch, "Send metrics in batch mode")
	cmd.Flags().StringVarP(&cfg.SecretKey, "key", "k", defaults.SecretKey, "SecretKey for signing metrics")
	cmd.Flags().IntVarP(&cfg.RateLimit, "rate-limit", "l", defaults.RateLimit, "Rate limit for sending metrics")
	cmd.Flags().StringVarP(&cfg.CryptoKey, "crypto-key", "", defaults.CryptoKey, "Public key or certificate path for payload encryption")

	return cfg, nil
}

func getFileConfigPath() string {
	if v := os.Getenv("CONFIG"); v != "" {
		return v
	}
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-c=") {
			return strings.TrimPrefix(a, "-c=")
		}
		if strings.HasPrefix(a, "-config=") {
			return strings.TrimPrefix(a, "-config=")
		}
		if a == "-c" || a == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

func overrideWithEnv(cfg *Config) error {
	if v := os.Getenv("ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("REPORT_INTERVAL"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("ошибка парсинга REPORT_INTERVAL: %w", err)
		}
		cfg.ReportInterval = n
	}
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("ошибка парсинга POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = n
	}
	if v := os.Getenv("PLAIN"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("ошибка парсинга PLAIN: %w", err)
		}
		cfg.Plain = b
	}
	if v := os.Getenv("GZIP"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("ошибка парсинга GZIP: %w", err)
		}
		cfg.EnableGzip = b
	}
	if v := os.Getenv("BATCH"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("ошибка парсинга BATCH: %w", err)
		}
		cfg.Batch = b
	}
	if v := os.Getenv("KEY"); v != "" {
		cfg.SecretKey = v
	}
	if v := os.Getenv("RATE_LIMIT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("ошибка парсинга RATE_LIMIT: %w", err)
		}
		cfg.RateLimit = n
	}
	if v := os.Getenv("CRYPTO_KEY"); v != "" {
		cfg.CryptoKey = v
	}
	return nil
}

func createSender(cfg *Config) (agent.Sender, error) {
	if cfg.Plain {
		return agent.NewMetricURLSender(cfg.Address), nil
	}

	var signer *signature.Signer
	if cfg.SecretKey != "" {
		signer = signature.NewSign(cfg.SecretKey)
	}

	var encrypter *agent.Encrypter
	if cfg.CryptoKey != "" {
		e, err := agent.NewEncrypterFromFile(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
		encrypter = e
	}

	return agent.NewJSONSender(cfg.Address, cfg.EnableGzip, signer, encrypter), nil
}
