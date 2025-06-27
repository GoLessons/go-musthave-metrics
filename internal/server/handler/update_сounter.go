package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"net/http"
	"strconv"
)

type UpdateCounter struct {
	storage storage.Storage[model.Counter]
}

func NewUpdateCounter(storage storage.Storage[model.Counter]) *UpdateCounter {
	return &UpdateCounter{
		storage: storage,
	}
}

func (h UpdateCounter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricName, ok := ctx.Value(server.MetricName).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var metric model.Counter
	metric, err := h.storage.Get(metricName)
	if err != nil {
		metric = *model.NewCounter(metricName)
	}

	metricValueRaw, ok := ctx.Value(server.MetricValue).(string)
	if !ok {
		http.Error(w, "Metric not defined", http.StatusBadRequest)
		return
	}
	metricValue, err := strconv.ParseInt(metricValueRaw, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Metric value incorrect (%s)", err.Error()), http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("Update counter: %s = %d + %d\n", metricName, metric.Value(), metricValue)))

	metric.Inc(metricValue)

	err = h.storage.Set(metricName, metric)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Write([]byte(fmt.Sprintf("Counter new value: %s = %d\n", metricName, metric.Value())))
}
