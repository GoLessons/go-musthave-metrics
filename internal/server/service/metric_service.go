package service

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
)

type MetricService struct {
	counterStorage storage.Storage[serverModel.Counter]
	gaugeStorage   storage.Storage[serverModel.Gauge]
}

func NewMetricService(
	counterStorage storage.Storage[serverModel.Counter],
	gaugeStorage storage.Storage[serverModel.Gauge],
) *MetricService {
	return &MetricService{
		counterStorage: counterStorage,
		gaugeStorage:   gaugeStorage,
	}
}

func (ms *MetricService) Save(metric model.Metrics) error {
	err := ms.validate(metric)
	if err != nil {
		return err
	}

	switch metric.MType {
	case model.Counter:
		var counter serverModel.Counter
		counter, err = ms.counterStorage.Get(metric.ID)
		if err != nil {
			counter = *serverModel.NewCounter(metric.ID)
		}

		counter.Inc(*metric.Delta)
		err = ms.counterStorage.Set(metric.ID, counter)
		if err != nil {
			return fmt.Errorf("failed to update counter: %s", err.Error())
		}

	case model.Gauge:
		var gauge serverModel.Gauge
		gauge, err = ms.gaugeStorage.Get(metric.ID)
		if err != nil {
			gauge = *serverModel.NewGauge(metric.ID)
		}

		gauge.Set(*metric.Value)
		err = ms.gaugeStorage.Set(metric.ID, gauge)
		if err != nil {
			return fmt.Errorf("failed to update gauge: %s", err.Error())
		}

	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	return nil
}

func (ms *MetricService) Read(metricType string, metricName string) (*model.Metrics, error) {
	switch metricType {
	case model.Counter:
		metric, err := ms.counterStorage.Get(metricName)
		if err != nil {
			return nil, fmt.Errorf("metric not found: %s", metricName)
		}

		val := metric.Value()
		return &model.Metrics{
			ID:    metric.Name(),
			MType: metric.Type(),
			Delta: &val,
		}, nil
	case model.Gauge:
		metric, err := ms.gaugeStorage.Get(metricName)
		if err != nil {
			return nil, fmt.Errorf("metric not found: %s", metricName)
		}

		val := metric.Value()
		return &model.Metrics{
			ID:    metric.Name(),
			MType: metric.Type(),
			Value: &val,
		}, nil
	}

	return nil, fmt.Errorf("unknown metric type: %s", metricType)
}

func (ms *MetricService) validate(metric model.Metrics) error {
	if metric.ID == "" || metric.MType == "" {
		return fmt.Errorf("missing required fields: id or type")
	}

	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("missing required field: delta")
		}

	case model.Gauge:
		if metric.Value == nil {

			return fmt.Errorf("missing required field: value")
		}

	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	return nil
}
