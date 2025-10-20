package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/pkg/repeater"
)

func SenderWorker(
	ctx context.Context,
	sendChan <-chan []model.Metrics,
	sender Sender,
	rateLimit int,
	batch bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var activeRequests sync.WaitGroup
	var semaphore chan struct{}
	if rateLimit > 0 {
		semaphore = make(chan struct{}, rateLimit)
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
			err = handleBatchMode(sender, metricsToSend)
		} else {
			err = handleSingleMode(sender, metricsToSend)
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

func handleBatchMode(sender Sender, metrics []model.Metrics) error {
	try := repeater.NewRepeater(func(err error) {
		fmt.Printf("Ошибка отправки пакета метрик: %v\n", err)
	})
	repeatStrategy := createRetryStrategy()
	_, err := try.Repeat(
		repeatStrategy,
		func() (any, error) {
			return nil, sendMetricsBatch(sender, metrics)
		},
	)
	if err != nil {
		return fmt.Errorf("can't send metrics batch after retries: %w", err)
	}

	return nil
}

func handleSingleMode(sender Sender, metrics []model.Metrics) error {
	try := repeater.NewRepeater(func(err error) {
		fmt.Printf("Ошибка отправки пакета метрик: %v\n", err)
	})
	repeatStrategy := createRetryStrategy()

	_, err := try.Repeat(
		repeatStrategy,
		func() (any, error) {
			return nil, sendMetricsByOne(sender, metrics)
		},
	)
	if err != nil {
		return fmt.Errorf("can't send metrics batch after retries: %w", err)
	}

	return nil
}

func sendMetricsBatch(sender Sender, metrics []model.Metrics) error {
	if batchSender, ok := sender.(BatchSender); ok {
		return batchSender.SendBatch(metrics)
	}
	return sendMetricsByOne(sender, metrics)
}

func sendMetricsByOne(sender Sender, metrics []model.Metrics) error {
	for _, metric := range metrics {
		if err := sender.Send(metric); err != nil {
			return fmt.Errorf("can't send metric: %s\n%w", metric.ID, err)
		}
	}
	return nil
}

func createRetryStrategy() repeater.Strategy {
	return repeater.NewFixedDelaysStrategy(
		NewAgentErrorClassifier().IsRetriable,
		time.Second*1,
		time.Second*3,
		time.Second*5,
	)
}
