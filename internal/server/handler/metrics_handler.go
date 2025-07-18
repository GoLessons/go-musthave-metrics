package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/goccy/go-json"
	"io"
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
	metricData, err := h.receiveMetric(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusBadRequest)
		return
	}

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
	metricData, err := h.receiveMetric(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.metricService.Save(*metricData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *metricsController) receiveMetric(r *http.Request) (*model.Metrics, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %s", err.Error())
	}
	defer r.Body.Close()

	var metrics model.Metrics
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %s", err.Error())
	}

	return &metrics, err
}
