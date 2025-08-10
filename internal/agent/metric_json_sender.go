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

	return sender.send("/update", metric)
}

func (sender *jsonSender) SendBatch(metrics []model.Metrics) error {
	return sender.send("/updates", metrics)
}

func (sender *jsonSender) Close() {
	defer sender.client.Close()
}

func (sender *jsonSender) prepareBody(data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	if sender.enableGzip {
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		if _, err := gzipWriter.Write(body); err != nil {
			return nil, fmt.Errorf("failed to compress request body: %w", err)
		}
		if err := gzipWriter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}
		return buf.Bytes(), nil
	}

	return body, nil
}

func (sender *jsonSender) send(endpoint string, data interface{}) error {
	body, err := sender.prepareBody(data)
	if err != nil {
		return err
	}

	request := sender.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body)

	if sender.enableGzip {
		request.SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip")
	}

	if sender.signer != nil {
		hash, err := sender.signer.Hash(body) // подписываем финальное тело
		if err != nil {
			return fmt.Errorf("failed to calculate hash: %w", err)
		}
		request.SetHeader("HashSHA256", hash)
	}

	resp, err := request.Post(endpoint)
	if err != nil {
		return WrapSendError(0, fmt.Sprintf("can't send data to %s", endpoint), err)
	}
	if resp.IsError() {
		return NewSendError(resp.StatusCode(), "can't send data to %s, response: %s", endpoint, resp.String())
	}
	return nil
}
