package collector

import (
	"context"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"time"
)

func RunAgentLoop(
	ctx context.Context,
	pollTicker, dumpTicker *time.Ticker,
	stg storage.Storage[model.Metrics],
	readers []agent.Reader,
	simpleReader *reader.SimpleMetricsReader,
	out chan<- []model.Metrics,
	stopSender func(),
) {
	defer stopSender()

	for {
		select {
		case <-ctx.Done():
			return

		case <-pollTicker.C:
			handlePollTick(ctx, stg, readers, simpleReader)

		case <-dumpTicker.C:
			if err := HandleDumpTick(ctx, stg, simpleReader, out); err != nil {
				// Если контекст отменён — выходим, иначе логируем и продолжаем
				if ctx.Err() != nil {
					return
				}
				fmt.Printf("can't fetch metrics: %v\n", err)
			}
		}
	}
}

func handlePollTick(
	ctx context.Context,
	stg storage.Storage[model.Metrics],
	readers []agent.Reader,
	simpleReader *reader.SimpleMetricsReader,
) {
	CollectAllMetrics(ctx, stg, readers, simpleReader)
}

func HandleDumpTick(
	ctx context.Context,
	stg storage.Storage[model.Metrics],
	simpleReader *reader.SimpleMetricsReader,
	out chan<- []model.Metrics,
) error {
	metrics, err := FetchAllMetrics(stg)
	if err != nil {
		return err
	}

	select {
	case out <- metrics:
		simpleReader.Reset()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
