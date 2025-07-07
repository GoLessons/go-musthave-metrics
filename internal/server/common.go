package server

type contextKey string

const (
	MetricName  contextKey = "metricName"
	MetricValue contextKey = "metricValue"
)
