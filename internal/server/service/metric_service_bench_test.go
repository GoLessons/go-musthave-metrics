package service

import (
	"fmt"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
)

func BenchmarkMetricService_SaveCounter(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := NewMetricService(sCounter, sGauge)

	ids := make([]string, 128)
	for i := range ids {
		ids[i] = fmt.Sprintf("counter_%d", i)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := ids[i%len(ids)]
		delta := int64(i % 100)
		_ = ms.Save(model.Metrics{ID: id, MType: model.Counter, Delta: &delta})
	}
}

func BenchmarkMetricService_SaveGauge(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := NewMetricService(sCounter, sGauge)

	ids := make([]string, 128)
	for i := range ids {
		ids[i] = fmt.Sprintf("gauge_%d", i)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := ids[i%len(ids)]
		val := float64(i % 100)
		_ = ms.Save(model.Metrics{ID: id, MType: model.Gauge, Value: &val})
	}
}

func BenchmarkMetricService_ReadCounter(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := NewMetricService(sCounter, sGauge)

	for i := 0; i < 1000; i++ {
		c := serverModel.NewCounter(fmt.Sprintf("c_%d", i))
		c.Inc(int64(i))
		_ = sCounter.Set(c.Name(), *c)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("c_%d", i%1000)
		_, _ = ms.Read(model.Counter, name)
	}
}

func BenchmarkMetricService_ReadGauge(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := NewMetricService(sCounter, sGauge)

	for i := 0; i < 1000; i++ {
		g := serverModel.NewGauge(fmt.Sprintf("g_%d", i))
		g.Set(float64(i))
		_ = sGauge.Set(g.Name(), *g)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("g_%d", i%1000)
		_, _ = ms.Read(model.Gauge, name)
	}
}

func BenchmarkMetricService_GetAll(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := NewMetricService(sCounter, sGauge)

	for i := 0; i < 2000; i++ {
		c := serverModel.NewCounter(fmt.Sprintf("c_%d", i))
		c.Inc(int64(i))
		_ = sCounter.Set(c.Name(), *c)

		g := serverModel.NewGauge(fmt.Sprintf("g_%d", i))
		g.Set(float64(i))
		_ = sGauge.Set(g.Name(), *g)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.Run("counters", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ms.GetAllCounters()
		}
	})
	b.Run("gauges", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ms.GetAllGauges()
		}
	})
}
