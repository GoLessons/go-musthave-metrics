package reader

import "github.com/GoLessons/go-musthave-metrics/internal/model"

type Reader interface {
	Refresh() error
	Fetch() ([]model.Metrics, error)
}
