package reader

type MetricReader[T any] interface {
	Get(name string) (T, bool)
}
