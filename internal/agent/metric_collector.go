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
		err := mc.handle()
		if err != nil {
			fmt.Printf("metrics handling failed: %v\n", err)
		}

		time.Sleep(mc.pollDuration)
	}
}

func (mc *MetricCollector) handle() error {
	isNeedSend := time.Since(mc.lastLogTime) >= mc.dumpInterval

	err := mc.handleMemStats(isNeedSend)
	if err != nil {
		return fmt.Errorf("can't handle metrics: %w", err)
	}

	mc.pollCounter.Increment()

	err = mc.counterStorage.Set(PollCount, mc.pollCounter.Count())
	if err != nil {
		return fmt.Errorf("storage error for: %s\n%w", PollCount, err)
	}

	randomValue := mc.randomizer.Randomize()
	err = mc.gaugeStorage.Set(RandomValue, randomValue)
	if err != nil {
		return fmt.Errorf("storage error for: %s\n%w", RandomValue, err)
	}

	if isNeedSend {
		err := mc.sender.Send(RandomValue, randomValue)
		if err != nil {
			return fmt.Errorf("Cannot send metric: %s\n%w", RandomValue, err)
		}

		err = mc.sender.Send(PollCount, mc.pollCounter.Count())
		if err != nil {
			return fmt.Errorf("Cannot send metric: %s\n%w", PollCount, err)
		}

		mc.lastLogTime = time.Now()

		// если все метрики успешно отправлены серверу, сбрасываем счётчик
		mc.pollCounter.Reset()
	}

	return nil
}

func (mc *MetricCollector) handleMemStats(isNeedSend bool) error {
	mc.memStatReader.Refresh()

	for _, metricName := range mc.memStatReader.SupportedMetrics() {
		metricVal, ok := mc.memStatReader.Get(metricName)
		if !ok {
			return fmt.Errorf("Cannot read metric: " + metricName)
		}

		err := mc.gaugeStorage.Set(metricName, GaugeValue(metricVal))
		if err != nil {
			return err
		}

		if isNeedSend {
			err := mc.sender.Send(metricName, metricVal)
			if err != nil {
				return fmt.Errorf("Cannot send metric: %s\n%v", metricName, err)
			}
		}
	}

	return nil
}
