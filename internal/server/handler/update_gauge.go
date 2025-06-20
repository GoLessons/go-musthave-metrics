package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/pkg"
	"net/http"
	"strconv"
)

type UpdateGauge struct {
	storage pkg.Storage[model.Gauge]
}

func NewUpdateGauge(storage pkg.Storage[model.Gauge]) *UpdateGauge {
	return &UpdateGauge{
		storage: storage,
	}
}

func (h UpdateGauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricName, ok := ctx.Value("metricName").(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var metric model.Gauge

	metric, err := h.storage.Get(metricName)
	if err != nil {
		metric = *model.NewGauge(metricName)
	}

	metricValueRaw, ok := ctx.Value("metricValue").(string)
	if !ok {
		http.Error(w, "Metric not defined", http.StatusBadRequest)
		return
	}
	metricValue, err := strconv.ParseFloat(metricValueRaw, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Metric value incorrect (%s)", err.Error()), http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("Gauge old value: %s = %d\n", metricName, metric.Value())))

	metric.Set(metricValue)

	w.Write([]byte(fmt.Sprintf("Update gauge: %s = %d\n", metricName, metric.Value(), metricValue)))
}
