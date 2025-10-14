package handler

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"net/http"
)

type ListController struct {
	counterStorage storage.Storage[model.Counter]
	gaugeStorage   storage.Storage[model.Gauge]
}

func NewListController(counterStorage storage.Storage[model.Counter], gaugeStorage storage.Storage[model.Gauge]) *ListController {
	return &ListController{
		counterStorage: counterStorage,
		gaugeStorage:   gaugeStorage,
	}
}

func (controller *ListController) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte("<table><tr><th>Метрика</th><th>Значение</th><tr>\n"))
	if err != nil {
		http.Error(w, "Can't render page", http.StatusInternalServerError)
		return
	}

	gaugeMetrics, err := controller.gaugeStorage.GetAll()
	if err != nil {
		http.Error(w, "Can't read gaugeMetrics", http.StatusInternalServerError)
		return
	}

	for _, metric := range gaugeMetrics {
		err := controller.renderMetric(metric.Name(), metric.Value(), w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Can't render gaugeMetrics: %v", err), http.StatusInternalServerError)
			return
		}
	}

	counterMetrics, err := controller.counterStorage.GetAll()
	if err != nil {
		http.Error(w, "Can't read counterMetrics", http.StatusInternalServerError)
		return
	}

	for _, metric := range counterMetrics {
		err := controller.renderMetric(metric.Name(), metric.Value(), w)
		if err != nil {
			http.Error(w, "Can't render counterMetrics", http.StatusInternalServerError)
			return
		}
	}

	_, err = w.Write([]byte("</table>\n"))
	if err != nil {
		http.Error(w, "Can't render page", http.StatusInternalServerError)
		return
	}
}

func (controller *ListController) renderMetric(metricName string, metricValue interface{}, w http.ResponseWriter) error {
	var strVal string
	switch v := metricValue.(type) {
	case int, int8, int16, int32, int64:
		strVal = fmt.Sprintf("%d", v)
	case float32, float64:
		strVal = fmt.Sprintf("%f", v)
	default:
		return fmt.Errorf("unsupported metric: value := %s of %v", metricValue, v)
	}

	_, err := w.Write([]byte(fmt.Sprintf("<tr><td>%s</td><td>%s</td><tr>\n", metricName, strVal)))
	if err != nil {
		return err
	}

	return nil
}
