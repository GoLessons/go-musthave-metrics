package agent

import (
	"context"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/pkg/repeater"
	"sync"
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
	rateLimit    int
}

func NewMetricCollector(
	storage storage.Storage[model.Metrics],
	readers []Reader,
	simpleReader *reader.SimpleMetricsReader,
	sender Sender,
	dumpInterval time.Duration,
	pollDuration time.Duration,
	rateLimit int,
) *MetricCollector {
	return &MetricCollector{
		storage:      storage,
		readers:      readers,
		simpleReader: simpleReader,
		sender:       sender,
		dumpInterval: dumpInterval,
		pollDuration: pollDuration,
		rateLimit:    rateLimit,
	}
}

func (mc *MetricCollector) Close() {
	mc.sender.Close()
}

func (mc *MetricCollector) CollectAndSendMetrics(ctx context.Context, batch bool) {
	pollTicker := time.NewTicker(mc.pollDuration)
	defer pollTicker.Stop()

	dumpTicker := time.NewTicker(mc.dumpInterval)
	defer dumpTicker.Stop()

	sendChan := make(chan []model.Metrics, 1)
	var sendGroup sync.WaitGroup

	sendGroup.Add(1)
	go mc.senderWorker(ctx, sendChan, batch, &sendGroup)

	for {
		select {
		case <-ctx.Done():
			close(sendChan)
			sendGroup.Wait()
			return
		case <-pollTicker.C:
			mc.collectAllMetrics(ctx)
		case <-dumpTicker.C:
			metrics, err := mc.fetchAllMetrics()
			if err != nil {
				fmt.Printf("can't fetch metrics: %v\n", err)
				continue
			}

			select {
			case sendChan <- metrics:
				mc.simpleReader.Reset()
			case <-ctx.Done():
				close(sendChan)
				sendGroup.Wait()
				return
			}
		}
	}
}

func (mc *MetricCollector) senderWorker(ctx context.Context, sendChan <-chan []model.Metrics, batch bool, wg *sync.WaitGroup) {
	defer wg.Done()

	var activeRequests sync.WaitGroup
	var semaphore chan struct{}
	if mc.rateLimit > 0 {
		semaphore = make(chan struct{}, mc.rateLimit)
	}

	sendMetrics := func(metricsToSend []model.Metrics) {
		defer activeRequests.Done()
		defer func() {
			if semaphore != nil {
				<-semaphore
			}
		}()

		var err error
		if batch {
			err = mc.handleBatchMode(metricsToSend)
		} else {
			err = mc.handleSingleMode(metricsToSend)
		}

		if err != nil {
			fmt.Printf("metrics sending failed: %v\n", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			activeRequests.Wait()
			return
		case metrics, ok := <-sendChan:
			if !ok {
				activeRequests.Wait()
				return
			}

			if semaphore != nil {
				select {
				case semaphore <- struct{}{}:
				case <-ctx.Done():
					activeRequests.Wait()
					return
				}
			}

			activeRequests.Add(1)
			go sendMetrics(metrics)
		}
	}
}

func (mc *MetricCollector) collectAllMetrics(ctx context.Context) {
	var wg sync.WaitGroup

	for _, r := range mc.readers {
		wg.Add(1)
		go func(reader Reader) {
			defer wg.Done()
			mc.collectFromReader(reader)
		}(r)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		mc.collectFromSimpleReader()
	}()

	// Ждем завершения чтения всех метрик
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (mc *MetricCollector) collectFromReader(reader Reader) {
	if err := reader.Refresh(); err != nil {
		fmt.Printf("can't refresh reader: %v\n", err)
		return
	}

	metrics, err := reader.Fetch()
	if err != nil {
		fmt.Printf("can't fetch metrics: %v\n", err)
		return
	}

	for _, metric := range metrics {
		if err := mc.storage.Set(metric.ID, metric); err != nil {
			fmt.Printf("can't store metric %s: %v\n", metric.ID, err)
		}
	}
}

func (mc *MetricCollector) collectFromSimpleReader() {
	if err := mc.simpleReader.Refresh(); err != nil {
		fmt.Printf("can't refresh simple reader: %v\n", err)
		return
	}

	metrics, err := mc.simpleReader.Fetch()
	if err != nil {
		fmt.Printf("can't fetch simple metrics: %v\n", err)
		return
	}

	for _, metric := range metrics {
		if err := mc.storage.Set(metric.ID, metric); err != nil {
			fmt.Printf("can't store simple metric %s: %v\n", metric.ID, err)
		}
	}
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

	return mc.sendMetricsByOne(metrics)
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
