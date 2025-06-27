package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"net/http"
	"strconv"
)

type GaugeController struct {
	storage storage.Storage[model.Gauge]
}

func NewGaugeController(storage storage.Storage[model.Gauge]) *GaugeController {
	return &GaugeController{
		storage: storage,
	}
}

func (h GaugeController) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricName, ok := ctx.Value(server.MetricName).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var metric model.Gauge
	metric, err := h.storage.Get(metricName)
	if err != nil {
		metric = *model.NewGauge(metricName)
	}

	metricValueRaw, ok := ctx.Value(server.MetricValue).(string)
	if !ok {
		http.Error(w, "Metric not defined", http.StatusBadRequest)
		return
	}
	metricValue, err := strconv.ParseFloat(metricValueRaw, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Metric value incorrect (%s)", err.Error()), http.StatusBadRequest)
		return
	}

	metric.Set(metricValue)

	err = h.storage.Set(metricName, metric)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h GaugeController) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricName, ok := ctx.Value(server.MetricName).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	metric, err := h.storage.Get(metricName)
	if err != nil {
		http.Error(w, "Metric not found: "+metricName, http.StatusNotFound)
		return
	}

	_, err = w.Write([]byte(strconv.FormatFloat(metric.Value(), 'f', -1, 64)))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
