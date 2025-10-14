package agent

import "github.com/GoLessons/go-musthave-metrics/internal/model"

type MetricReader[T any] interface {
	Get(name string) (T, bool)
}

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
