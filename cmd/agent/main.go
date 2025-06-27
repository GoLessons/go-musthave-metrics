package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"time"
)

func main() {
	gaugeStorage := storage.NewMemStorage[agent.GaugeValue]()
	counterStorage := storage.NewMemStorage[agent.CounterValue]()
	memStatReader := agent.NewMemStatsReader()
	poolCounter := agent.NewPollCounter(0)
	randomizer := agent.NewRandomizer()
	dumpInterval := 10 * time.Second
	lastLogTime := time.Now()
	sender := agent.NewMetricSender()
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
			metricVal, ok := memStatReader.Get("Alloc")
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

		//fmt.Printf("%s: %f\n", "RandomValue", randomValue)
		//fmt.Printf("%s: %d\n", "PollCount", poolCounter.Count())

		if isNeedSend {
			lastLogTime = time.Now()
		}

		time.Sleep(time.Second * 2)
	}
}
