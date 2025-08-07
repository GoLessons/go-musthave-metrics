package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/Masterminds/squirrel"
	"go.uber.org/zap"
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
