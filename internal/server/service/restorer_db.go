package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/Masterminds/squirrel"
	"sync"
)

type dbMetricRestorer struct {
	db    *sql.DB
	mutex sync.Mutex
}

func NewDBMetricRestorer(db *sql.DB) *dbMetricRestorer {
	return &dbMetricRestorer{
		db: db,
	}
}

func (r *dbMetricRestorer) Restore() ([]model.Metrics, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	query := squirrel.Select("name", "type", "delta", "value").
		From("metrics.metrics")

	rows, err := query.RunWith(r.db).QueryContext(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	metrics, err := r.hydrate(rows)
	if err != nil {
		return metrics, err
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over metric rows: %w", err)
	}

	return metrics, nil
}

func (r *dbMetricRestorer) hydrate(rows *sql.Rows) ([]model.Metrics, error) {
	var metrics []model.Metrics
	for rows.Next() {
		var metric model.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64

		if err := rows.Scan(&metric.ID, &metric.MType, &delta, &value); err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}

		switch metric.MType {
		case model.Counter:
			if delta.Valid {
				deltaVal := delta.Int64
				metric.Delta = &deltaVal
			}
		case model.Gauge:
			if value.Valid {
				valueVal := value.Float64
				metric.Value = &valueVal
			}
		}

		metrics = append(metrics, metric)
	}
	return metrics, nil
}
