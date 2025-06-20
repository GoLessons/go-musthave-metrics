package model

import "github.com/GoLessons/go-musthave-metrics/internal/model"

type Metric struct {
	name       string
	metricType string
}

func (m *Metric) Name() string {
	return m.name
}

func (m *Metric) Type() string {
	return m.metricType
}

type Counter struct {
	Metric
	value int64
}

func NewCounter(name string) *Counter {
	return &Counter{
		Metric: Metric{
			name:       name,
			metricType: model.Counter,
		},
		value: 0,
	}
}

func (m *Counter) Inc(val int64) {
	m.value += val
}

func (m *Counter) Value() int64 {
	return m.value
}

type Gauge struct {
	Metric
	value float64
}

func NewGauge(name string) *Gauge {
	return &Gauge{
		Metric: Metric{
			name:       name,
			metricType: model.Gauge,
		},
		value: 0.0,
	}
}

func (m *Gauge) Set(value float64) {
	m.value = value
}

func (m *Gauge) Value() float64 {
	return m.value
}
