package storage

import (
	"errors"
	"sync"
)

type MemStorage[T any] struct {
	mutex     sync.RWMutex
	container map[string]T
}

func NewMemStorage[T any]() *MemStorage[T] {
	return &MemStorage[T]{container: make(map[string]T)}
}

func (s *MemStorage[T]) Set(key string, value T) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.container[key] = value
	return nil
}

func (s *MemStorage[T]) Get(key string) (T, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	metric, exists := s.container[key]
	if !exists {
		return metric, errors.New("metric not found")
	}

	return metric, nil
}

func (s *MemStorage[T]) Unset(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.container, key)
	return nil
}

func (s *MemStorage[T]) GetAll() (map[string]T, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Возвраращаем мапы, чтобы избежать гонок
	result := make(map[string]T, len(s.container))
	for k, v := range s.container {
		result[k] = v
	}

	return result, nil
}
