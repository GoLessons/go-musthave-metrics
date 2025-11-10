package service

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openBenchDB(b *testing.B) *sql.DB {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		b.Skip("DATABASE_DSN is not set; skipping DB benchmarks")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		b.Skipf("failed to open DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		b.Skipf("failed to ping DB: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		b.Skipf("failed to ensure schema: %v", err)
	}

	return db
}

func ensureSchema(db *sql.DB) error {
	const createSchema = `CREATE SCHEMA IF NOT EXISTS "metrics";`
	const createTable = `
CREATE TABLE IF NOT EXISTS "metrics"."metrics" (
    "name"  text NOT NULL,
    "type"  text NOT NULL,
    "delta" bigint DEFAULT NULL,
    "value" double precision,
    PRIMARY KEY ("name","type")
);`
	if _, err := db.Exec(createSchema); err != nil {
		return err
	}
	if _, err := db.Exec(createTable); err != nil {
		return err
	}
	return nil
}

func truncateMetrics(db *sql.DB) error {
	_, err := db.Exec(`TRUNCATE TABLE "metrics"."metrics"`)
	return err
}

func makeBenchMetrics(n int) []model.Metrics {
	metrics := make([]model.Metrics, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			delta := int64(i)
			metrics[i] = *model.NewCounter("c_bench_"+strconv.Itoa(i), &delta)
		} else {
			value := float64(i) * 0.5
			metrics[i] = *model.NewGauge("g_bench_"+strconv.Itoa(i), &value)
		}
	}
	return metrics
}
