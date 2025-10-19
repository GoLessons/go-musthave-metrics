package agent

import (
	"sync"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
)

type simpleSenderMock struct {
	mu     sync.Mutex
	sent   []model.Metrics
	err    error
	closed bool
}

func (m *simpleSenderMock) Send(metric model.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, metric)
	return m.err
}
func (m *simpleSenderMock) Close() { m.closed = true }

type batchSenderMock struct {
	simpleSenderMock
	batchCalls int
	batchErr   error
}

func (m *batchSenderMock) SendBatch(metrics []model.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batchCalls++
	m.sent = append(m.sent, metrics...)
	return m.batchErr
}

func TestHandleSingleMode_Success(t *testing.T) {
	val := 42.0
	metrics := []model.Metrics{
		*model.NewGauge("g1", &val),
		*model.NewGauge("g2", &val),
	}

	sender := &simpleSenderMock{}
	if err := handleSingleMode(sender, metrics); err != nil {
		t.Fatalf("handleSingleMode returned error: %v", err)
	}

	if len(sender.sent) != len(metrics) {
		t.Fatalf("expected %d metrics sent, got %d", len(metrics), len(sender.sent))
	}
}

func TestHandleBatchMode_UsesBatchWhenAvailable(t *testing.T) {
	val := 7.0
	metrics := []model.Metrics{
		*model.NewGauge("g1", &val),
		*model.NewGauge("g2", &val),
		*model.NewGauge("g3", &val),
	}

	sender := &batchSenderMock{}
	if err := handleBatchMode(sender, metrics); err != nil {
		t.Fatalf("handleBatchMode returned error: %v", err)
	}

	if sender.batchCalls != 1 {
		t.Fatalf("expected SendBatch to be called once, got %d", sender.batchCalls)
	}
	if len(sender.sent) != len(metrics) {
		t.Fatalf("expected %d metrics captured in mock after batch send, got %d", len(metrics), len(sender.sent))
	}
}

func TestHandleBatchMode_FallbackToSingle(t *testing.T) {
	val := 1.23
	metrics := []model.Metrics{
		*model.NewGauge("g1", &val),
		*model.NewGauge("g2", &val),
	}

	sender := &simpleSenderMock{}
	if err := handleBatchMode(sender, metrics); err != nil {
		t.Fatalf("handleBatchMode returned error: %v", err)
	}

	if len(sender.sent) != len(metrics) {
		t.Fatalf("expected %d metrics sent by fallback, got %d", len(metrics), len(sender.sent))
	}
}
