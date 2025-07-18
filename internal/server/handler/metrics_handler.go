package handler

import (
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/goccy/go-json"
	"net/http"
	"strconv"
)

type metricsController struct {
	metricService   service.MetricService
	responseBuilder ResponseBuilder
}

type ResponseBuilder func(*http.ResponseWriter, *model.Metrics)

func NewMetricsController(metricService service.MetricService, responseBuilder ResponseBuilder) *metricsController {
	return &metricsController{
		metricService:   metricService,
		responseBuilder: responseBuilder,
	}
}

func (h *metricsController) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricData := ctx.Value(server.Metric).(model.Metrics)

	metric, err := h.metricService.Read(metricData.MType, metricData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.responseBuilder(&w, metric)
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

func JSONResposeBuilder(w *http.ResponseWriter, metric *model.Metrics) {
	responseBody, err := json.Marshal(metric)
	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}

	(*w).Header().Set("Content-Type", "application/json")
	_, err = (*w).Write(responseBody)
	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}

func PlainResposeBuilder(w *http.ResponseWriter, metric *model.Metrics) {
	var responseBody []byte
	switch metric.MType {
	case model.Counter:
		responseBody = []byte(strconv.FormatInt(*metric.Delta, 10))
	case model.Gauge:
		responseBody = []byte(strconv.FormatFloat(*metric.Value, 'g', -1, 64))
	default:
		http.Error(*w, "Unsupported metric type", http.StatusInternalServerError)
	}

	(*w).Header().Set("Content-Type", "plain/text")
	_, err := (*w).Write(responseBody)
	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
