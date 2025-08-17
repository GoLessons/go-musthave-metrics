package agent

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/pkg/repeater"
	"time"
)

type Sender interface {
	Send(model.Metrics) error
	Close()
}

type BatchSender interface {
	SendBatch(metrics []model.Metrics) error
	Sender
}

type Reader interface {
	Refresh() error
	Fetch() ([]model.Metrics, error)
}

type ResetableReader interface {
	Reset()
}

type MetricCollector struct {
	storage      storage.Storage[model.Metrics]
	readers      []Reader
	simpleReader *reader.SimpleMetricsReader
	sender       Sender
	dumpInterval time.Duration
	pollDuration time.Duration
	lastLogTime  time.Time
}

func NewMetricCollector(
	storage storage.Storage[model.Metrics],
	readers []Reader,
	simpleReader *reader.SimpleMetricsReader,
	sender Sender,
	dumpInterval time.Duration,
	pollDuration time.Duration,
) *MetricCollector {
	return &MetricCollector{
		storage:      storage,
		readers:      readers,
		simpleReader: simpleReader,
		sender:       sender,
		dumpInterval: dumpInterval,
		pollDuration: pollDuration,
		lastLogTime:  time.Now(),
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
		mc.simpleReader.Reset()
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
	all, err := mc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("can't get all metrics: %w", err)
	}

	metrics := make([]model.Metrics, 0, len(all))
	for _, metric := range all {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (mc *MetricCollector) collectAllMetrics() error {
	for _, r := range mc.readers {
		err := r.Refresh()
		if err != nil {
			return fmt.Errorf("can't refresh runtime metrics: %w", err)
		}

		metrics, err := r.Fetch()
		if err != nil {
			return fmt.Errorf("can't fetch runtime metrics: %w", err)
		}

		for _, metric := range metrics {
			err = mc.storage.Set(metric.ID, metric)
			if err != nil {
				return fmt.Errorf("can't store runtime metric %s: %w", metric.ID, err)
			}
		}
	}

	err := mc.simpleReader.Refresh()
	if err != nil {
		return fmt.Errorf("can't refresh simple metrics: %w", err)
	}

	simpleMetrics, err := mc.simpleReader.Fetch()
	if err != nil {
		return fmt.Errorf("can't fetch simple metrics: %w", err)
	}

	for _, metric := range simpleMetrics {
		err = mc.storage.Set(metric.ID, metric)
		if err != nil {
			return fmt.Errorf("can't store simple metric %s: %w", metric.ID, err)
		}
	}

	return nil
}
