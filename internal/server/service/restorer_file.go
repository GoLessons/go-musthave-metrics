package service

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"os"
	"sync"
)

type fileMetricRestorer struct {
	filePath string
	mutex    sync.Mutex
}

func NewFileMetricRestorer(filePath string) *fileMetricRestorer {
	return &fileMetricRestorer{
		filePath: filePath,
	}
}

func (r *fileMetricRestorer) Restore() (metrics []model.Metrics, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return metrics, nil
	}

	file, err := os.OpenFile(r.filePath, os.O_RDONLY, 0644)
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
