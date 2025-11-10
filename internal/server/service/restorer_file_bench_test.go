package service

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
)

func BenchmarkFileMetricRestorer_Restore(b *testing.B) {
	tmp := filepath.Join(b.TempDir(), "metrics_restore.json")
	d := NewFileMetricDumper(tmp)

	metrics := make([]model.Metrics, 2000)
	for i := range metrics {
		if i%2 == 0 {
			delta := int64(i)
			metrics[i] = *model.NewCounter("c_bench_"+strconv.Itoa(i), &delta)
		} else {
			value := float64(i) * 0.5
			metrics[i] = *model.NewGauge("g_bench_"+strconv.Itoa(i), &value)
		}
	}
	_ = d.Dump(metrics)

	r := NewFileMetricRestorer(tmp)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = r.Restore()
	}
}
