package main

import (
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/spf13/cobra"
	"net/http"
)

var address string

var rootCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&address, "address", "a", "localhost:8080", "Metric server address")
	rootCmd.FParseErrWhitelist.UnknownFlags = false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe(address, router.InitRouter())
}
