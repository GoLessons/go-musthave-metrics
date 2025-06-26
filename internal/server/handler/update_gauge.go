package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/storage"
	"net/http"
	"strconv"
)

type UpdateGauge struct {
	storage storage.Storage[model.Gauge]
}

func NewUpdateGauge(storage storage.Storage[model.Gauge]) *UpdateGauge {
	return &UpdateGauge{
		storage: storage,
	}
}

func (h UpdateGauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	w.Write([]byte(fmt.Sprintf("Update old gauge: %s = %f\n", metricName, metric.Value())))

	metric.Set(metricValue)

	err = h.storage.Set(metricName, metric)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Write([]byte(fmt.Sprintf("Gauge new value: %s = %f\n", metricName, metric.Value())))
}
