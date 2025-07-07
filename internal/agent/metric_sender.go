package agent

import (
	"fmt"
	"resty.dev/v3"
)

type sender struct {
	client *resty.Client
}

func NewMetricSender(address string) *sender {
	client := resty.New()
	client.SetBaseURL("http://"+address).
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

	resp, err := client.R().
		SetPathParam("metricName", metric.name).
		SetPathParam("metricType", metric.metricType).
		SetPathParam("metricVal", metric.value).
		Post("/update/{metricType}/{metricName}/{metricVal}")
	if err != nil {
		return fmt.Errorf("can't send metric: %s = %s\nprevious: %w", metric.name, metric.value, err)
	}

	if resp.IsError() {
		return fmt.Errorf("can't send metric: %s = %s\nresponse: %s", metric.name, metric.value, resp.String())
	}

	fmt.Printf("metric was sent succsesfully: %s = %s\n", metric.name, metric.value)
	return nil
}

func (sender *sender) convertMetricData(metricName string, metricValue interface{}) (*metricData, error) {
	var metricType, strVal string
	switch v := metricValue.(type) {
	case int, int8, int16, int32, int64, CounterValue:
		strVal = fmt.Sprintf("%d", v)
		metricType = "counter"
	case float32, float64, GaugeValue:
		strVal = fmt.Sprintf("%f", v)
		metricType = "gauge"
	default:
		return nil, fmt.Errorf("unsupported metric: value := %s of %v", metricValue, v)
	}

	return &metricData{
			name:       metricName,
			metricType: metricType,
			value:      strVal,
		},
		nil
}
