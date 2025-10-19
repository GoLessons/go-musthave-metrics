package agent

import (
	"fmt"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"resty.dev/v3"
)

type urPathSender struct {
	client *resty.Client
}

func NewMetricURLSender(address string) *urPathSender {
	client := resty.New()
	client.SetBaseURL("http://"+address).
		SetHeader("Content-Type", "text/plain")

	return &urPathSender{
		client: client,
	}
}

type metricData struct {
	name       string
	metricType string
	value      string
}

func (sender *urPathSender) Close() {
	defer sender.client.Close()
}

func (sender *urPathSender) Send(metric model.Metrics) (err error) {
	client := sender.client

	metricData, err := sender.convertMetricData(metric)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetPathParam("metricName", metricData.name).
		SetPathParam("metricType", metricData.metricType).
		SetPathParam("metricVal", metricData.value).
		Post("/update/{metricType}/{metricName}/{metricVal}")
	if err != nil {
		return WrapSendError(0, fmt.Sprintf("can't send metric: %s = %s", metricData.name, metricData.value), err)
	}

	if resp.IsError() {
		return NewSendError(resp.StatusCode(), "can't send metric: %s = %s\nresponse: %s", metricData.name, metricData.value, resp.String())
	}

	return nil
}

func (sender *urPathSender) convertMetricData(metric model.Metrics) (*metricData, error) {
	var strVal string

	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return nil, fmt.Errorf("counter metric %s has nil Delta", metric.ID)
		}
		strVal = fmt.Sprintf("%d", *metric.Delta)
	case model.Gauge:
		if metric.Value == nil {
			return nil, fmt.Errorf("gauge metric %s has nil Value", metric.ID)
		}
		strVal = fmt.Sprintf("%f", *metric.Value)
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	return &metricData{
			name:       metric.ID,
			metricType: metric.MType,
			value:      strVal,
		},
		nil
}
