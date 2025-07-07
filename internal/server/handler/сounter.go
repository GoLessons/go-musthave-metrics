package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"net/http"
	"strconv"
)

type CounterController struct {
	storage storage.Storage[model.Counter]
}

func NewCounterController(storage storage.Storage[model.Counter]) *CounterController {
	return &CounterController{
		storage: storage,
	}
}

func (h CounterController) Update(w http.ResponseWriter, r *http.Request) {
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

	metric.Inc(metricValue)

	err = h.storage.Set(metricName, metric)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h CounterController) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricName, ok := ctx.Value(server.MetricName).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	metric, err := h.storage.Get(metricName)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	_, err = w.Write([]byte(fmt.Sprintf("%d\n", metric.Value())))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
