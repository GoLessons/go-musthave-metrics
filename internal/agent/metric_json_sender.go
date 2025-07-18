package agent

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"resty.dev/v3"
)

type jsonSender struct {
	client *resty.Client
}

func NewJSONSender(address string) *jsonSender {
	client := resty.New()
	client.SetBaseURL("http://"+address).
		SetHeader("Content-Type", "application/json")

	return &jsonSender{
		client: client,
	}
}

func (sender *jsonSender) Send(metric model.Metrics) error {
	client := sender.client

	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric %s has nil Delta", metric.ID)
		}
	case model.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric %s has nil Value", metric.ID)
		}
	default:
		return fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	resp, err := client.R().
		SetBody(metric).
		Post("/update")
	if err != nil {
		return fmt.Errorf("can't send metric: %s (type: %s): %w", metric.ID, metric.MType, err)
	}

	if resp.IsError() {
		return fmt.Errorf("can't send metric: %s (type: %s), response: %s", metric.ID, metric.MType, resp.String())
	}

	return nil
}

func (sender *jsonSender) Close() {
	defer sender.client.Close()
}
