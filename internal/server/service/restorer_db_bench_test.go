package service

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkDBMetricRestorer_Restore(b *testing.B) {
	db := openBenchDB(b)
	defer db.Close()

	if err := truncateMetrics(db); err != nil {
		b.Fatalf("failed to truncate table: %v", err)
	}

	seed := makeBenchMetrics(2000)
	if err := NewDBMetricDumper(db, zap.NewNop()).Dump(seed); err != nil {
		b.Fatalf("failed to seed metrics: %v", err)
	}

	r := NewDBMetricRestorer(db)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := r.Restore(); err != nil {
			b.Fatalf("restore error: %v", err)
		}
	}
}
