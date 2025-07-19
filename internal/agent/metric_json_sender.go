package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"resty.dev/v3"
)

type jsonSender struct {
	client      *resty.Client
	disableGzip bool
}

func NewJSONSender(address string, disableGzip bool) *jsonSender {
	client := resty.New()
	client.SetBaseURL("http://"+address).
		SetHeader("Content-Type", "application/json")

	return &jsonSender{
		client:      client,
		disableGzip: disableGzip,
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

	request := client.R()

	if !sender.disableGzip {
		request.SetHeader("Content-Encoding", "gzip")
		request.SetHeader("Accept-Encoding", "gzip")

		body, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("failed to marshal metric: %w", err)
		}

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)

		_, err = gzipWriter.Write(body)
		if err != nil {
			return fmt.Errorf("failed to compress request body: %w", err)
		}

		err = gzipWriter.Close()
		if err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}

		request.SetBody(buf.Bytes())
	} else {
		request.SetBody(metric)
	}

	resp, err := request.Post("/update")
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
