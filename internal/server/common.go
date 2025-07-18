package server

type contextKey string

const (
	MetricName  contextKey = "metricName"
	MetricType  contextKey = "metricType"
	MetricValue contextKey = "metricValue"
)
