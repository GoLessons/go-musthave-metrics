package agent

import (
	"fmt"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
)

var (
	PollCount   string = "PollCount"
	RandomValue string = "RandomValue"
)

type MetricCollector struct {
	gaugeStorage   storage.Storage[GaugeValue]
	counterStorage storage.Storage[CounterValue]
	memStatReader  *memStatsReader[float64]
	pollCounter    *pollCounter[CounterValue]
	randomizer     *Randomizer
	sender         *sender
	dumpInterval   time.Duration
	pollDuration   time.Duration
	lastLogTime    time.Time
}

func NewMetricCollector(
	gaugeStorage storage.Storage[GaugeValue],
	counterStorage storage.Storage[CounterValue],
	memStatReader *memStatsReader[float64],
	sender *sender,
	dumpInterval time.Duration,
	pollDuration time.Duration,
) *MetricCollector {
	return &MetricCollector{
		gaugeStorage:   gaugeStorage,
		counterStorage: counterStorage,
		memStatReader:  memStatReader,
		pollCounter:    NewPollCounter(0),
		randomizer:     NewRandomizer(),
		sender:         sender,
		dumpInterval:   dumpInterval,
		pollDuration:   pollDuration,
		lastLogTime:    time.Now(),
	}
}

func (mc *MetricCollector) Close() {
	mc.sender.Close()
}

func (mc *MetricCollector) CollectAndSendMetrics() {
	for {
		isNeedSend := time.Since(mc.lastLogTime) >= mc.dumpInterval

		mc.handleMemStats(isNeedSend)
		mc.pollCounter.Increment()

		err := mc.counterStorage.Set(PollCount, mc.pollCounter.Count())
		if err != nil {
			fmt.Println(err)
		}

		randomValue := mc.randomizer.Randomize()
		err = mc.gaugeStorage.Set(RandomValue, randomValue)
		if err != nil {
			fmt.Println(err)
		}

		if isNeedSend {
			err := mc.sender.Send(RandomValue, randomValue)
			if err != nil {
				fmt.Printf("Cannot send metric: %s\n%v", RandomValue, err)
			}

			err = mc.sender.Send(PollCount, mc.pollCounter.Count())
			if err != nil {
				fmt.Printf("Cannot send metric: %s\n%v", PollCount, err)
			}

			mc.lastLogTime = time.Now()
			mc.pollCounter.Reset()
		}

		time.Sleep(mc.pollDuration)
	}
}

func (mc *MetricCollector) handleMemStats(isNeedSend bool) {
	mc.memStatReader.Refresh()

	for _, metricName := range mc.memStatReader.SupportedMetrics() {
		metricVal, ok := mc.memStatReader.Get(metricName)
		if !ok {
			fmt.Println("Cannot read metric: " + metricName)
		}

		err := mc.gaugeStorage.Set(metricName, GaugeValue(metricVal))
		if err != nil {
			fmt.Println(err)
		}

		if isNeedSend {
			err := mc.sender.Send(metricName, metricVal)
			if err != nil {
				fmt.Printf("Cannot send metric: %s\n%v", metricName, err)
			}
		}
	}
}
