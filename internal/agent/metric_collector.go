package agent

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/pkg/repeater"
	"time"
)

var (
	PollCount   = "PollCount"
	RandomValue = "RandomValue"
)

type Sender interface {
	Send(model.Metrics) error
	Close()
}

type BatchSender interface {
	SendBatch(metrics []model.Metrics) error
	Sender
}

type MetricCollector struct {
	gaugeStorage   storage.Storage[GaugeValue]
	counterStorage storage.Storage[CounterValue]
	memStatReader  *memStatsReader[float64]
	pollCounter    *pollCounter[CounterValue]
	randomizer     *Randomizer
	sender         Sender
	dumpInterval   time.Duration
	pollDuration   time.Duration
	lastLogTime    time.Time
}

func NewMetricCollector(
	gaugeStorage storage.Storage[GaugeValue],
	counterStorage storage.Storage[CounterValue],
	memStatReader *memStatsReader[float64],
	sender Sender,
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

func (mc *MetricCollector) CollectAndSendMetrics(batch bool) {
	for {
		err := mc.handle(batch)
		if err != nil {
			fmt.Printf("metrics handling failed: %v\n", err)
		}

		time.Sleep(mc.pollDuration)
	}
}

func (mc *MetricCollector) handle(batch bool) error {
	err := mc.collectAllMetrics()
	if err != nil {
		return fmt.Errorf("can't collect metrics: %w", err)
	}

	isNeedSend := time.Since(mc.lastLogTime) >= mc.dumpInterval
	if isNeedSend {
		metrics, err := mc.fetchAllMetrics()
		if err != nil {
			return fmt.Errorf("can't fetch metrics: %w", err)
		}

		if batch {
			err = mc.handleBatchMode(metrics)
		} else {
			err = mc.handleSingleMode(metrics)
		}

		if err != nil {
			return err
		}

		mc.lastLogTime = time.Now()
		mc.pollCounter.Reset()
	}

	return nil
}

func (mc *MetricCollector) createRetryStrategy() repeater.Strategy {
	return repeater.NewFixedDelaysStrategy(
		NewAgentErrorClassifier().IsRetriable,
		time.Second*1,
		time.Second*3,
		time.Second*5,
	)
}

func (mc *MetricCollector) handleBatchMode(metrics []model.Metrics) error {
	try := repeater.NewRepeater(func(err error) {
		fmt.Printf("Ошибка отправки пакета метрик: %v\n", err)
	})
	repeatStrategy := mc.createRetryStrategy()
	_, err := try.Repeat(
		repeatStrategy,
		func() (any, error) {
			return nil, mc.sendMetricsBatch(metrics)
		},
	)

	if err != nil {
		return fmt.Errorf("can't send metrics batch after retries: %w", err)
	}

	return nil
}

func (mc *MetricCollector) handleSingleMode(metrics []model.Metrics) error {
	try := repeater.NewRepeater(func(err error) {
		fmt.Printf("Ошибка отправки пакета метрик: %v\n", err)
	})
	repeatStrategy := mc.createRetryStrategy()

	_, err := try.Repeat(
		repeatStrategy,
		func() (any, error) {
			return nil, mc.sendMetricsByOne(metrics)
		},
	)
	if err != nil {
		return fmt.Errorf("can't send metrics batch after retries: %w", err)
	}

	return nil
}

func (mc *MetricCollector) sendMetricsBatch(metrics []model.Metrics) error {
	if batchSender, ok := mc.sender.(BatchSender); ok {
		return batchSender.SendBatch(metrics)
	}

	err := mc.sendMetricsByOne(metrics)
	if err != nil {
		return err
	}

	return nil
}

func (mc *MetricCollector) sendMetricsByOne(metrics []model.Metrics) error {
	for _, metric := range metrics {
		err := mc.sender.Send(metric)
		if err != nil {
			return fmt.Errorf("can't send metric: %s\n%w", metric.ID, err)
		}
	}

	return nil
}

func (mc *MetricCollector) fetchAllMetrics() ([]model.Metrics, error) {
	metrics := []model.Metrics{}

	randomValue, err := mc.gaugeStorage.Get(RandomValue)
	if err != nil {
		return nil, fmt.Errorf("can't get random value: %w", err)
	}
	metrics = append(metrics, *model.NewGauge(
		RandomValue,
		(*float64)(&randomValue),
	))

	poolCount, err := mc.counterStorage.Get(PollCount)
	if err != nil {
		return nil, fmt.Errorf("can't get poll count: %w", err)
	}
	metrics = append(metrics, *model.NewCounter(
		PollCount,
		(*int64)(&poolCount),
	))

	for _, metricName := range mc.memStatReader.SupportedMetrics() {
		metricValue, err := mc.gaugeStorage.Get(metricName)
		if err != nil {
			return nil, fmt.Errorf("can't get random value: %w", err)
		}
		metrics = append(metrics, *model.NewGauge(
			metricName,
			(*float64)(&metricValue),
		))
	}

	return metrics, nil
}

func (mc *MetricCollector) collectAllMetrics() error {
	mc.memStatReader.Refresh()

	for _, metricName := range mc.memStatReader.SupportedMetrics() {
		metricVal, ok := mc.memStatReader.Get(metricName)
		if !ok {
			return fmt.Errorf("can't read metric: %s", metricName)
		}

		err := mc.gaugeStorage.Set(metricName, GaugeValue(metricVal))
		if err != nil {
			return err
		}
	}

	randomValue := mc.randomizer.Randomize()
	err := mc.gaugeStorage.Set(RandomValue, randomValue)
	if err != nil {
		return fmt.Errorf("storage error for: %s\n%w", RandomValue, err)
	}

	mc.pollCounter.Increment()
	err = mc.counterStorage.Set(PollCount, mc.pollCounter.Count())
	if err != nil {
		return fmt.Errorf("storage error for: %s\n%w", PollCount, err)
	}

	return nil
}
