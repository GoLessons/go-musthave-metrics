package storage

import (
	"fmt"
	"testing"

	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
)

func BenchmarkMemStorage_Set(b *testing.B) {
	s := NewMemStorage[serverModel.Counter]()
	keys := make([]string, 1024)
	for i := range keys {
		keys[i] = fmt.Sprintf("k_%d", i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := serverModel.NewCounter(keys[i%len(keys)])
		c.Inc(int64(i))
		_ = s.Set(c.Name(), *c)
	}
}

func BenchmarkMemStorage_Get(b *testing.B) {
	s := NewMemStorage[serverModel.Gauge]()
	for i := 0; i < 4096; i++ {
		g := serverModel.NewGauge(fmt.Sprintf("g_%d", i))
		g.Set(float64(i))
		_ = s.Set(g.Name(), *g)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get(fmt.Sprintf("g_%d", i%4096))
	}
}

func BenchmarkMemStorage_GetAll(b *testing.B) {
	s := NewMemStorage[serverModel.Counter]()
	for i := 0; i < 4096; i++ {
		c := serverModel.NewCounter(fmt.Sprintf("c_%d", i))
		c.Inc(int64(i))
		_ = s.Set(c.Name(), *c)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.GetAll()
	}
}
