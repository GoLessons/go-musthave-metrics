package test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/config"
	serverConfig "github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type tester struct {
	t                  *testing.T
	testServer         *httptest.Server
	httpClient         *http.Client
	testStorageCounter *storage.MemStorage[model.Counter]
	testStorageGauge   *storage.MemStorage[model.Gauge]
}

func NewTester(t *testing.T, options *map[string]any) *tester {
	cfg, err := serverConfig.LoadConfig(options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var testStorageCounter = storage.NewMemStorage[model.Counter]()
	var testStorageGauge = storage.NewMemStorage[model.Gauge]()
	metricService := service.NewMetricService(testStorageCounter, testStorageGauge)
	serverLogger, _ := logger.NewLogger(zap.NewDevelopmentConfig())

	c := container.NewSimpleContainer(map[string]any{
		"logger":         serverLogger,
		"config":         cfg,
		"counterStorage": testStorageCounter,
		"gaugeStorage":   testStorageGauge,
		"metricService":  metricService,
	})
	container.SimpleRegisterFactory(&c, "db", config.DBFactory())
	container.SimpleRegisterFactory(&c, "router", router.RouterFactory())

	r, err := container.GetService[chi.Mux](c, "router")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	return &tester{
		t:                  t,
		testServer:         httptest.NewServer(r),
		httpClient:         &http.Client{},
		testStorageCounter: testStorageCounter,
		testStorageGauge:   testStorageGauge,
	}
}

func (tester *tester) HaveCouner(metric model.Counter) error {
	return tester.testStorageCounter.Set(metric.Name(), metric)
}

func (tester *tester) HaveGauge(metric model.Gauge) error {
	return tester.testStorageGauge.Set(metric.Name(), metric)
}

func (tester *tester) Post(path string, body interface{}) (*http.Response, error) {
	return tester.DoRequest(http.MethodPost, path, body, map[string]string{"Content-Type": "text/plain"})
}

func (tester *tester) Get(path string) (*http.Response, error) {
	return tester.DoRequest(http.MethodGet, path, nil, map[string]string{"Content-Type": "text/plain"})
}

/*func (tester *tester) DoRequest(method string, endpoint string, body interface{}, headers map[string]string) (resp *http.Response, err error) {
	bodyReader, err := tester.buildBody(body, headers)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(method, tester.testServer.URL+endpoint, bodyReader)
	req.Header.Del("Accept-Encoding")
	req.Header.Del("Content-Encoding")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	resp, err = httpClient.Do(req)

	if testing.Verbose() {
		tester.t.Logf("--- Request ---")
		tester.t.Logf("URL: %s", req.URL)
		tester.t.Logf("Header: %s", req.Header)
		if req.Body != nil {
			reqBodyBytes, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			tester.t.Logf("Body: %s", string(reqBodyBytes))
		} else {
			tester.t.Logf("Body: <nil>")
		}
	}
	if testing.Verbose() && resp != nil {
		tester.t.Logf("--- Response ---")
		tester.t.Logf("Status: %d", resp.StatusCode)
		tester.t.Logf("Headers: %+v", resp.Header)
		if resp.Body != nil {
			respBodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))
			tester.t.Logf("Body: %s", string(respBodyBytes))
		}
	}

	return resp, err
}*/

func (tester *tester) DoRequest(method string, endpoint string, body interface{}, headers map[string]string) (resp *http.Response, err error) {
	bodyBytes, err := tester.buildBody(body, headers)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequest(method, tester.testServer.URL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	if bodyBytes != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	req.Header.Del("Accept-Encoding")
	req.Header.Del("Content-Encoding")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if testing.Verbose() {
		tester.t.Logf("--- Request ---")
		tester.t.Logf("%s %s", req.Method, req.URL)
		tester.t.Logf("Header: %s", req.Header)

		if req.Body != nil {
			rb, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(rb))
			tester.t.Logf("Body: %s", string(rb))
		} else {
			tester.t.Logf("Body: <nil>")
		}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	resp, err = httpClient.Do(req)

	if testing.Verbose() && resp != nil {
		tester.t.Logf("--- Response ---")
		tester.t.Logf("Status: %d", resp.StatusCode)
		tester.t.Logf("Headers: %+v", resp.Header)

		if resp.Body != nil {
			rb, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(rb))
			const maxLog = 4096
			if len(rb) > maxLog {
				tester.t.Logf("Body (truncated %d bytes): %s...", len(rb), string(rb[:maxLog]))
			} else {
				tester.t.Logf("Body: %s", string(rb))
			}
		} else {
			tester.t.Logf("Body: <nil>")
		}
	}

	return resp, err
}

func (tester *tester) buildBody(body interface{}, headers map[string]string) (bodyBytes []byte, err error) {
	tester.t.Logf("Try convert [%T]: %v", body, body)
	switch v := body.(type) {
	case []byte:
		bodyBytes = v
	case string:
		bodyBytes = []byte(v)
	case map[string]interface{}:
		bodyBytes, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if headers != nil {
			if _, ok := headers["Content-Type"]; !ok {
				headers["Content-Type"] = "application/json"
			}
		}
	case io.ReadCloser:
		bodyBytes, err = io.ReadAll(v)
		v.Close()
		if err != nil {
			return nil, err
		}
	case io.Reader:
		bodyBytes, err = io.ReadAll(v)
		if err != nil {
			return nil, err
		}
	default:
		if headers != nil && strings.EqualFold(headers["Content-Type"], "application/json") {
			bodyBytes, err = json.Marshal(v)
			if err != nil {
				return nil, err
			}
		} else {
			bodyBytes = []byte(fmt.Sprint(v))
		}
	}
	return bodyBytes, err
}

func (tester *tester) Shutdown() {
	defer tester.testServer.Close()
}

func (tester *tester) ReadGzip(resp *http.Response) ([]byte, error) {
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return io.ReadAll(gr)
	}

	return io.ReadAll(resp.Body)
}
