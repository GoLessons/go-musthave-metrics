package pkg

type Storage interface {
	Set(key []byte, data []byte) error
	Get(key []byte) ([]byte, error)
	Unset(key []byte) error
}
