package main

import (
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/spf13/cobra"
	"net/http"
)

type Config struct {
	Address string
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "server",
	}

	cfg := loadConfig(rootCmd)
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

func loadConfig(cmd *cobra.Command) *Config {
	cfg := &Config{}

	cmd.Flags().StringVarP(&cfg.Address, "address", "a", "localhost:8080", "HTTP server address")

	return cfg
}
