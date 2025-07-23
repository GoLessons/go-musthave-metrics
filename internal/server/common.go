package server

type contextKey string

const (
	Metric      contextKey = "metric"
	MetricName  contextKey = "metricName"
	MetricType  contextKey = "metricType"
	MetricValue contextKey = "metricValue"
)
