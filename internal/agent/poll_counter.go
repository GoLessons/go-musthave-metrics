package agent

type pollCounter[T CounterValue] struct {
	count uint64
}

func NewPollCounter[T CounterValue](startValue uint64) *pollCounter[T] {
	return &pollCounter[T]{
		count: startValue,
	}
}

func (p *pollCounter[T]) Increment() {
	p.count++
}

func (p *pollCounter[T]) Reset() {
	p.count = 0
}

func (p *pollCounter[T]) Count() CounterValue {
	return CounterValue(p.count)
}
