package storage

import (
	"errors"
)

type MemStorage[T any] struct {
	container map[string]T
}

func NewMemStorage[T any]() *MemStorage[T] {
	return &MemStorage[T]{container: make(map[string]T)}
}

func (s *MemStorage[T]) Set(key string, value T) error {
	s.container[key] = value
	return nil
}

func (s *MemStorage[T]) Get(key string) (T, error) {
	metric, exists := s.container[key]
	if !exists {
		return metric, errors.New("metric not found")
	}

	return metric, nil
}

func (s *MemStorage[T]) Unset(key string) error {
	delete(s.container, key)
	return nil
}
