package reader

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"sync"
)

type SystemMetricsReader struct {
	mu              sync.RWMutex
	totalMemory     uint64
	freeMemory      uint64
	cpuUtilizations []float64
}

func NewSystemMetricsReader() *SystemMetricsReader {
	return &SystemMetricsReader{}
}

func (r *SystemMetricsReader) Refresh() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory info: %w", err)
	}

	r.totalMemory = memInfo.Total
	r.freeMemory = memInfo.Free

	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Errorf("failed to get CPU info: %w", err)
	}
	r.cpuUtilizations = cpuPercents

	return nil
}

func (r *SystemMetricsReader) Fetch() ([]model.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalMem := float64(r.totalMemory)
	freeMem := float64(r.freeMemory)

	metrics := make([]model.Metrics, 0, 2+len(r.cpuUtilizations))
	metrics = append(metrics,
		*model.NewGauge("TotalMemory", &totalMem),
		*model.NewGauge("FreeMemory", &freeMem),
	)

	for i, util := range r.cpuUtilizations {
		u := util
		metrics = append(metrics, *model.NewGauge(fmt.Sprintf("CPUutilization%d", i+1), &u))
	}

	return metrics, nil
}
