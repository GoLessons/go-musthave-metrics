package service

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkDBMetricDumper_Dump(b *testing.B) {
	db := openBenchDB(b)
	defer db.Close()

	if err := truncateMetrics(db); err != nil {
		b.Fatalf("failed to truncate table: %v", err)
	}

	d := NewDBMetricDumper(db, zap.NewNop())
	metrics := makeBenchMetrics(2000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := d.Dump(metrics); err != nil {
			b.Fatalf("dump error: %v", err)
		}
	}
}
