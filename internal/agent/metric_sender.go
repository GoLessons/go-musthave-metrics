package agent

import (
	"fmt"
	"resty.dev/v3"
)

type sender struct {
	client *resty.Client
}

func NewMetricSender() *sender {
	client := resty.New()
	client.SetBaseURL("http://localhost:8080").
		SetHeader("Content-Type", "text/plain")

	return &sender{
		client: client,
	}
}

type metricData struct {
	name       string
	metricType string
	value      string
}

func (sender *sender) Close() {
	defer sender.client.Close()
}

func (sender *sender) Send(metricName string, value interface{}) (err error) {
	client := sender.client

	metric, err := sender.convertMetricData(metricName, value)
	if err != nil {
		return err
	}

	_, err = client.R().
		SetPathParam("metricName", metric.name).
		SetPathParam("metricType", metric.metricType).
		SetPathParam("metricVal", metric.value).
		Post("/update/{metricType}/{metricName}/{metricVal}")
	if err != nil {
		return fmt.Errorf("can't send metric: %s = %s\nprevious: %w", metric.name, metric.value, err)
	}

	fmt.Printf("metric was sent succsesfully: %s = %s\n", metric.name, metric.value)
	return nil
}

func (sender *sender) convertMetricData(metricName string, value interface{}) (*metricData, error) {
	var metricType, strVal string
	if _, ok := value.(int64); ok {
		metricType = "counter"
		strVal = fmt.Sprintf("%d", value)
	} else if _, ok := value.(float64); ok {
		metricType = "gauge"
		strVal = fmt.Sprintf("%f", value)
	} else {
		return nil, fmt.Errorf("unsupported metric: type := %s, value := %s", metricType, value)
	}

	return &metricData{
			name:       metricName,
			metricType: metricType,
			value:      strVal,
		},
		nil
}
