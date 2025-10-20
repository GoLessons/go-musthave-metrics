package collector

import (
	"context"
	"sync"

	"github.com/GoLessons/go-musthave-metrics/internal/agent"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
)

func StartSenderPipeline(
	ctx context.Context,
	sender agent.Sender,
	rateLimit int,
	batch bool,
	buffer int,
) (chan<- []model.Metrics, func()) {
	sendChan := make(chan []model.Metrics, buffer)

	var wg sync.WaitGroup
	wg.Add(1)
	go agent.SenderWorker(ctx, sendChan, sender, rateLimit, batch, &wg)

	stop := func() {
		close(sendChan)
		wg.Wait()
	}

	return sendChan, stop
}
