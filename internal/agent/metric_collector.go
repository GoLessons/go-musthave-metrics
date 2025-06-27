package agent

type MetricCollector struct {
}

func NewMetricCollector() *MetricCollector {
	return new(MetricCollector)
}

func (m *MetricCollector) SetGauge(name string, value CounterValue) {

}

func (m *MetricCollector) SetCounter(name string, value GaugeValue) {

}
