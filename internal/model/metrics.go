package model

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func NewGauge(ID string, Value *float64) *Metrics {
	return &Metrics{
		ID:    ID,
		MType: Gauge,
		
	}
}

func NewCounter(ID string, Delta *int64) *Metrics {
	return &Metrics{
		ID:    ID,
		MType: Counter,
pe
	}
}
