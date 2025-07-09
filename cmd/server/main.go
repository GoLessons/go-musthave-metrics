package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/caarlos0/env"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "server",
	}

	cfg, err := loadConfig(rootCmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		return run(cfg)
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(cfg *Config) error {
	return http.ListenAndServe(cfg.Address, router.InitRouter())
}

func loadConfig(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&cfg.Address, "address", "a", cfg.Address, "HTTP server address")

	return cfg, nil
}
