package test

import (
	"bytes"
	"encoding/json"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"net/http"
	"net/http/httptest"
)

type tester struct {
	testServer *httptest.Server
	httpClient *http.Client
}

func NewTester() *tester {
	var testStorageCounter = storage.NewMemStorage[model.Counter]()
	var testStorageGauge = storage.NewMemStorage[model.Gauge]()
	return &tester{
		testServer: httptest.NewServer(router.InitRouter(testStorageCounter, testStorageGauge)),
		httpClient: http.DefaultClient,
	}
}

func (tester *tester) Post(path string, body interface{}) (*http.Response, error) {
	return tester.DoRequest(http.MethodPost, path, body)
}

func (tester *tester) Get(path string) (*http.Response, error) {

	return tester.DoRequest(http.MethodGet, path, nil)
}

func (tester *tester) DoRequest(method string, endpoint string, body interface{}) (resp *http.Response, err error) {
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

	req, _ := http.NewRequest(method, tester.testServer.URL+endpoint, &data)
	req.Header.Set("Content-Type", "text/plain")

	httpClient := tester.httpClient
	resp, err = httpClient.Do(req)

	return resp, err
}

func (tester *tester) Shutdown() {
	defer tester.testServer.Close()
}
