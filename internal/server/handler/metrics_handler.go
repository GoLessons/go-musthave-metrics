package handler

import (
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/goccy/go-json"
	"net/http"
)

type metricsController struct {
	metricService service.MetricService
}

func NewMetricsController(metricService service.MetricService) *metricsController {
	return &metricsController{
		metricService: metricService,
	}
}

func (h *metricsController) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricData := ctx.Value(server.Metric).(model.Metrics)

	metric, err := h.metricService.Read(metricData.MType, metricData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *metricsController) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricData := ctx.Value(server.Metric).(model.Metrics)

	err := h.metricService.Save(metricData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
