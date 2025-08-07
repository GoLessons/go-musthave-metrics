package service

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	database "github.com/GoLessons/go-musthave-metrics/internal/server/db"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/pkg/repeater"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type MetricDumper interface {
	Dump([]model.Metrics) error
}

type MetricRestorer interface {
	Restore() ([]model.Metrics, error)
}

type MetricStorageService struct {
	metricService  *MetricService
	metricDumper   MetricDumper
	metricRestorer MetricRestorer
	shutdownCh     chan os.Signal
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

func NewMetricStorageService(
	metricService *MetricService,
	metricDumper MetricDumper,
	metricRestorer MetricRestorer,
) *MetricStorageService {
	return &MetricStorageService{
		metricService:  metricService,
		metricDumper:   metricDumper,
		metricRestorer: metricRestorer,
		shutdownCh:     make(chan os.Signal, 1),
		stopCh:         make(chan struct{}),
	}
}

func (s *MetricStorageService) Start(interval uint64) error {
	signal.Notify(s.shutdownCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	s.wg.Add(1)
	go s.autoStore(interval)

	s.wg.Add(1)
	go s.shutdown()

	return nil
}

func (s *MetricStorageService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *MetricStorageService) autoStore(storeInterval uint64) {
	defer s.wg.Done()

	if storeInterval == 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer ticker.Stop()

	try := repeater.NewRepeater(func(err error) {
		fmt.Printf("error dumping metrics: %v\n", err)
	})
	repeatStrategy := repeater.NewFixedDelaysStrategy(
		database.NewPostgresErrorClassifier().IsRetriable,
		time.Second*1,
		time.Second*3,
		time.Second*5,
	)

	for {
		select {
		case <-ticker.C:
			_, err := try.Repeat(
				repeatStrategy,
				func() (any, error) {
					err := StoreState(s.metricService, s.metricDumper)
					return nil, err
				},
			)
			if err != nil {
				fmt.Printf("error dumping metrics after retries: %v\n", err)
			}
		case <-s.stopCh:
			_, err := try.Repeat(
				repeatStrategy,
				func() (any, error) {
					err := StoreState(s.metricService, s.metricDumper)
					return nil, err
				},
			)
			if err != nil {
				fmt.Printf("error dumping metrics during shutdown after retries: %v\n", err)
			}
			return
		}
	}
}

func (s *MetricStorageService) shutdown() {
	defer s.wg.Done()

	select {
	case <-s.shutdownCh:
		s.Stop()
	case <-s.stopCh:
		return
	}
}

func StoreState(metricService *MetricService, metricDumper MetricDumper) error {
	counters, err := metricService.GetAllCounters()
	if err != nil {
		return fmt.Errorf("failed to get counters: %w", err)
	}

	gauges, err := metricService.GetAllGauges()
	if err != nil {
		return fmt.Errorf("failed to get gauges: %w", err)
	}

	metrics := convertMetrics(counters, gauges)
	return metricDumper.Dump(metrics)
}

func RestoreState(metricService *MetricService, metricRestorer MetricRestorer) error {
	metrics, err := metricRestorer.Restore()
	if err != nil {
		return fmt.Errorf("failed to restore metrics: %w", err)
	}

	fmt.Println("Restored metrics:", metrics)

	for _, metric := range metrics {
		if err := metricService.Save(metric); err != nil {
			return fmt.Errorf("failed to save restored metric %s: %w", metric.ID, err)
		}
	}

	return nil
}

func convertMetrics(counters map[string]serverModel.Counter, gauges map[string]serverModel.Gauge) []model.Metrics {
	metrics := []model.Metrics{}

	for _, counter := range counters {
		delta := counter.Value()
		metrics = append(metrics, model.Metrics{
			ID:    counter.Name(),
			MType: counter.Type(),
			Delta: &delta,
		})
	}

	for _, gauge := range gauges {
		value := gauge.Value()
		metrics = append(metrics,
			model.Metrics{
				ID:    gauge.Name(),
				MType: gauge.Type(),
				Value: &value,
			})
	}

	fmt.Printf("Metrics For Dump: %v\n", metrics)

	return metrics
}
