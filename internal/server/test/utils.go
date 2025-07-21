package test

import (
	"bytes"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/goccy/go-json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type tester struct {
	t                  *testing.T
	testServer         *httptest.Server
	httpClient         *http.Client
	testStorageCounter *storage.MemStorage[model.Counter]
	testStorageGauge   *storage.MemStorage[model.Gauge]
}

func NewTester(t *testing.T) *tester {
	var testStorageCounter = storage.NewMemStorage[model.Counter]()
	var testStorageGauge = storage.NewMemStorage[model.Gauge]()
	metricService := service.NewMetricService(testStorageCounter, testStorageGauge)
	return &tester{
		testServer:         httptest.NewServer(router.InitRouter(metricService, testStorageCounter, testStorageGauge)),
		httpClient:         http.DefaultClient,
		testStorageCounter: testStorageCounter,
		testStorageGauge:   testStorageGauge,
		t:                  t,
	}
}

func (tester *tester) HaveCouner(metric model.Counter) error {
	return tester.testStorageCounter.Set(metric.Name(), metric)
}

func (tester *tester) HaveGauge(metric model.Gauge) error {
	return tester.testStorageGauge.Set(metric.Name(), metric)
}

func (tester *tester) Post(path string, body interface{}) (*http.Response, error) {
	return tester.DoRequest(http.MethodPost, path, body, "text/plain")
}

func (tester *tester) Get(path string) (*http.Response, error) {
	return tester.DoRequest(http.MethodGet, path, nil, "text/plain")
}

func (tester *tester) DoRequest(method string, endpoint string, body interface{}, contentType string) (resp *http.Response, err error) {
	var data bytes.Buffer
	if s, ok := body.(string); ok {
		data = *bytes.NewBuffer([]byte(s))
	}
	if m, ok := body.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		data = *bytes.NewBuffer(jsonData)
	}
	if _, ok := body.(string); !ok && contentType == "application/json" {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		if testing.Verbose() {
			tester.t.Logf("JSON data: %s", string(jsonData))
		}

		data = *bytes.NewBuffer(jsonData)
	}

	req, _ := http.NewRequest(method, tester.testServer.URL+endpoint, &data)
	req.Header.Set("Content-Type", contentType)
	req.Header.Del("Accept-Encoding")
	req.Header.Del("Content-Encoding")

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	resp, err = httpClient.Do(req)

	/*if testing.Verbose() && resp != nil {
		tester.t.Logf("--- Request ---")
		tester.t.Logf("URL: %s", req.URL)
		tester.t.Logf("Header: %s", req.Header)
		reqBody, _ := io.ReadAll(req.Body)
		tester.t.Logf("Body: %s", reqBody)
		tester.t.Logf("--- Response ---")
		tester.t.Logf("Status: %d", resp.StatusCode)
		tester.t.Logf("Headers: %+v", resp.Header)
		respBody, _ := io.ReadAll(resp.Body)
		tester.t.Logf("Body: %s", respBody)
	}*/

	return resp, err
}

func (tester *tester) Shutdown() {
	defer tester.testServer.Close()
}
