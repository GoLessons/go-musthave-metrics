package agent

type MetricCollector struct {
}

func NewMetricCollector() *MetricCollector {
	return new(MetricCollector)
}

func (m *MetricCollector) SetGauge(name string, value float64) {

}

func (m *MetricCollector) SetCounter(name string, value int64) {

}
