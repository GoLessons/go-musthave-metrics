package agent

type CounterValue int64
type GaugeValue float64

type MetricReader[T any] interface {
	Get(name string) (T, bool)
}
