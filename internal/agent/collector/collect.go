package collector

import (
	"context"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"sync"
)

func CollectAllMetrics(
	ctx context.Context,
	stg storage.Storage[model.Metrics],
	readers []agent.Reader,
	simpleReader *reader.SimpleMetricsReader,
) {
	var wg sync.WaitGroup

	for _, r := range readers {
		wg.Add(1)
		go func(rd agent.Reader) {
			defer wg.Done()
			collectFromReader(stg, rd)
		}(r)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		CollectFromSimpleReader(stg, simpleReader)
	}()

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

func collectFromReader(stg storage.Storage[model.Metrics], rd agent.Reader) {
	if err := rd.Refresh(); err != nil {
		fmt.Printf("can't refresh reader: %v\n", err)
		return
	}

	metrics, err := rd.Fetch()
	if err != nil {
		fmt.Printf("can't fetch metrics: %v\n", err)
		return
	}

	for _, metric := range metrics {
		if err := stg.Set(metric.ID, metric); err != nil {
			fmt.Printf("can't store metric %s: %v\n", metric.ID, err)
		}
	}
}

func CollectFromSimpleReader(stg storage.Storage[model.Metrics], simpleReader *reader.SimpleMetricsReader) {
	if err := simpleReader.Refresh(); err != nil {
		fmt.Printf("can't refresh simple reader: %v\n", err)
		return
	}

	metrics, err := simpleReader.Fetch()
	if err != nil {
		fmt.Printf("can't fetch simple metrics: %v\n", err)
		return
	}

	for _, metric := range metrics {
		if err := stg.Set(metric.ID, metric); err != nil {
			fmt.Printf("can't store simple metric %s: %v\n", metric.ID, err)
		}
	}
}

func FetchAllMetrics(stg storage.Storage[model.Metrics]) ([]model.Metrics, error) {
	all, err := stg.GetAll()
	if err != nil {
		return nil, fmt.Errorf("can't get all metrics: %w", err)
	}

	metrics := make([]model.Metrics, 0, len(all))
	for _, metric := range all {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
