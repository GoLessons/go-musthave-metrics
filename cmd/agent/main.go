package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/spf13/cobra"
	"time"
)

var (
	address        string
	reportInterval int
	pollInterval   int
)

var rootCmd = &cobra.Command{
	Use: "metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportInterval <= 0 {
			return fmt.Errorf("report interval must be positive, got %d", reportInterval)
		}
		if pollInterval <= 0 {
			return fmt.Errorf("poll interval must be positive, got %d", pollInterval)
		}

		metricCommand()
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&address, "address", "a", "localhost:8080", "Metric server address")
	rootCmd.Flags().IntVarP(&reportInterval, "report", "r", 10, "Report interval in seconds")
	rootCmd.Flags().IntVarP(&pollInterval, "poll", "p", 2, "Poll interval in seconds")

	rootCmd.FParseErrWhitelist.UnknownFlags = false
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("agent error: %v\n", err)
	}
}

func metricCommand() {
	gaugeStorage := storage.NewMemStorage[agent.GaugeValue]()
	counterStorage := storage.NewMemStorage[agent.CounterValue]()
	memStatReader := agent.NewMemStatsReader()
	poolCounter := agent.NewPollCounter(0)
	randomizer := agent.NewRandomizer()
	dumpInterval := time.Duration(reportInterval) * time.Second
	pollDuration := time.Duration(pollInterval) * time.Second
	lastLogTime := time.Now()
	sender := agent.NewMetricSender(address)
	defer sender.Close()

	for {
		memStatReader.Refresh()
		poolCounter.Increment()
		isNeedSend := time.Since(lastLogTime) >= dumpInterval
		err := counterStorage.Set("PollCount", poolCounter.Count())
		if err != nil {
			fmt.Println(err)
		}

		for _, metricName := range memStatReader.SupportedMetrics() {
			metricVal, ok := memStatReader.Get(metricName)
			if !ok {
				fmt.Println("Cannot read metric: " + metricName)
			}

			err := gaugeStorage.Set(metricName, agent.GaugeValue(metricVal))
			if err != nil {
				fmt.Println(err)
			}

			if isNeedSend {
				err := sender.Send(metricName, metricVal)
				if err != nil {
					fmt.Printf("Cannot send metric: %s\n%v", metricName, err)
				}
			}
		}

		randomValue := randomizer.Randomize()

		err = gaugeStorage.Set("RandomValue", randomValue)
		if err != nil {
			fmt.Println(err)
		}

		if isNeedSend {
			err := sender.Send("RandomValue", randomValue)
			if err != nil {
				fmt.Printf("Cannot send metric: %s\n%v", "RandomValue", err)
			}

			err = sender.Send("PollCount", poolCounter.Count())
			if err != nil {
				fmt.Printf("Cannot send metric: %s\n%v", "PollCount", err)
			}

			lastLogTime = time.Now()
		}

		time.Sleep(pollDuration)
	}
}
