package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/Masterminds/squirrel"
	"go.uber.org/zap"
	"log"
	"sync"
)

type dbMetricDumper struct {
	db     *sql.DB
	logger *zap.Logger
	mutex  sync.Mutex
}

func NewDBMetricDumper(db *sql.DB, logger *zap.Logger) *dbMetricDumper {
	return &dbMetricDumper{
		db:     db,
		logger: logger,
	}
}

func (d *dbMetricDumper) Dump(metrics []model.Metrics) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if len(metrics) == 0 {
		return nil
	}

	insert := squirrel.Insert("metrics.metrics").
		Columns("name", "type", "delta", "value").
		PlaceholderFormat(squirrel.Dollar).
		Suffix("ON CONFLICT (name,type) DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value")

	for _, metric := range metrics {
		insert = insert.Values(metric.ID, metric.MType, metric.Delta, metric.Value)
	}

	queryString, args, err := insert.ToSql()
	if err != nil {
		return err
	}

	d.logger.Info(queryString, zap.Any("args", args))

	_, err = insert.RunWith(d.db).ExecContext(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to execute insert query: %w", err)
	}

	return nil
}

// Новая структура для реализации интерфейса MetricRestorer
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

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	log.Printf("%s %v", queryString, args)

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
