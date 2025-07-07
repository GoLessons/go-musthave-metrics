package storage

type Storage[T any] interface {
	Set(key string, value T) error
	Get(key string) (T, error)
	GetAll() (map[string]T, error)
	Unset(key string) error
}
