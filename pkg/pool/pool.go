package pool

import "sync"

type Resettable interface {
	Reset()
	comparable
}

type Pool[T Resettable] struct {
	p sync.Pool
}

func New[T Resettable](factory func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

func (pl *Pool[T]) Get() T {
	v := pl.p.Get()
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

func (pl *Pool[T]) Put(x T) {
	var null T
	if x == null {
		return
	}
	x.Reset()
	pl.p.Put(x)
}
