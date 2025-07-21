package service

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"os"
	"sync"
)

type fileMetricDumper struct {
	filePath string
	mutex    sync.Mutex
}

func NewFileMetricDumper(filePath string) *fileMetricDumper {
	return &fileMetricDumper{
		filePath: filePath,
	}
}

func (d *fileMetricDumper) Dump(metrics []model.Metrics) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	tmpFilePath := d.filePath + ".tmp"
	file, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write metrics to file: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	if err := os.Rename(tmpFilePath, d.filePath); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func (d *fileMetricDumper) Restore() (metrics []model.Metrics, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, err := os.Stat(d.filePath); os.IsNotExist(err) {
		return metrics, err
	}

	file, err := os.OpenFile(d.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return metrics, nil
}
