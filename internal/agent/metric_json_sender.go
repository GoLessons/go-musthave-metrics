package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"net/http"
	"resty.dev/v3"
)

type jsonSender struct {
	client     *resty.Client
	enableGzip bool
	signer     *signature.Signer
}

func NewJSONSender(address string, enableGzip bool, signer *signature.Signer) *jsonSender {
	client := resty.New().SetTransport(&http.Transport{
		DisableCompression: true,
	})

	client.SetBaseURL("http://"+address).
		SetHeader("Content-Type", "application/json")

	return &jsonSender{
		client:     client,
		enableGzip: enableGzip,
		signer:     signer,
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

	var body []byte
	var err error

	if sender.enableGzip {
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
		body, err = json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("failed to marshal metric: %w", err)
		}
		request.SetBody(metric)
	}

	// Добавляем подпись, если signer не nil
	if sender.signer != nil {
		hash, err := sender.signer.Hash(body)
		if err != nil {
			return fmt.Errorf("failed to calculate hash: %w", err)
		}
		request.SetHeader("HashSHA256", hash)
	}

	resp, err := request.Post("/update")
	if err != nil {
		return WrapSendError(0, fmt.Sprintf("can't send metric: %s (type: %s)", metric.ID, metric.MType), err)
	}

	if resp.IsError() {
		return NewSendError(resp.StatusCode(), "can't send metric: %s (type: %s), response: %s", metric.ID, metric.MType, resp.String())
	}

	return nil
}

func (sender *jsonSender) SendBatch(metrics []model.Metrics) error {
	client := sender.client
	request := client.R()

	if sender.enableGzip {
		request.SetHeader("Content-Encoding", "gzip")
		request.SetHeader("Accept-Encoding", "gzip")

		body, err := json.Marshal(metrics)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics: %w", err)
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
		request.SetBody(metrics)
	}

	resp, err := request.Post("/updates")
	if err != nil {
		return WrapSendError(0, "can't send metrics batch", err)
	}

	if resp.IsError() {
		return NewSendError(resp.StatusCode(), "can't send metrics batch, response: %s", resp.String())
	}

	return nil
}

func (sender *jsonSender) Close() {
	defer sender.client.Close()
}
