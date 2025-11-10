package test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/agent"
)

const (
	privateKeyPath = "/home/sufir/Ya.less/go-musthave-metrics/var/keys/private.key"
	certPath       = "/home/sufir/Ya.less/go-musthave-metrics/var/keys/cert.pem"
)

func encryptBody(t *testing.T, data []byte) ([]byte, map[string]string) {
	t.Helper()
	e, err := agent.NewEncrypterFromFile(certPath)
	if err != nil {
		t.Fatalf("encrypter init: %v", err)
	}
	body, headers, err := e.Encrypt(data)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return body, headers
}

func gzipData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(data); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func TestUpdateEncrypted_NoGzip_Success(t *testing.T) {
	opts := map[string]any{
		"Key":       "",
		"CryptoKey": privateKeyPath,
	}
	tester, err := NewTester(t, &opts)
	if err != nil {
		t.Fatalf("tester init: %v", err)
	}
	defer tester.Shutdown()

	raw := []byte(`{"id":"g1","type":"gauge","value":1.23}`)
	body, encHeaders := encryptBody(t, raw)
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Encrypted":  encHeaders["X-Encrypted"],
	}
	resp, err := tester.DoRequest(http.MethodPost, "/update", body, headers)
	if err != nil {
		t.Fatalf("post update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	getResp, err := tester.Get("/value/gauge/g1")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected get status: %d", getResp.StatusCode)
	}
	val, err := tester.ReadGzip(getResp)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(val) != "1.23" {
		t.Fatalf("unexpected value: %s", string(val))
	}
}

func TestUpdateEncrypted_Gzip_Success(t *testing.T) {
	opts := map[string]any{
		"Key":       "",
		"CryptoKey": privateKeyPath,
	}
	tester, err := NewTester(t, &opts)
	if err != nil {
		t.Fatalf("tester init: %v", err)
	}
	defer tester.Shutdown()

	raw := []byte(`{"id":"g2","type":"gauge","value":2.34}`)
	body, encHeaders := encryptBody(t, raw)
	compressed := gzipData(t, body)
	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-Encrypted":      encHeaders["X-Encrypted"],
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}
	resp, err := tester.DoRequest(http.MethodPost, "/update", compressed, headers)
	if err != nil {
		t.Fatalf("post update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	getResp, err := tester.Get("/value/gauge/g2")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected get status: %d", getResp.StatusCode)
	}
	val, err := tester.ReadGzip(getResp)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(val) != "2.34" {
		t.Fatalf("unexpected value: %s", string(val))
	}
}

func TestUpdateEncrypted_InvalidContainer_400(t *testing.T) {
	opts := map[string]any{
		"Key":       "",
		"CryptoKey": privateKeyPath,
	}
	tester, err := NewTester(t, &opts)
	if err != nil {
		t.Fatalf("tester init: %v", err)
	}
	defer tester.Shutdown()

	invalid := []byte(`{"alg":"wrong","k":"","n":"","d":"","v":1}`)
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Encrypted":  "aes256gcm+rsa-oaep;v=1",
	}
	resp, err := tester.DoRequest(http.MethodPost, "/update", invalid, headers)
	if err != nil {
		t.Fatalf("post update: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got: %d", resp.StatusCode)
	}
}

func TestUpdateEncrypted_MissingKey_400(t *testing.T) {
	opts := map[string]any{
		"Key":       "",
		"CryptoKey": "",
	}
	tester, err := NewTester(t, &opts)
	if err != nil {
		t.Fatalf("tester init: %v", err)
	}
	defer tester.Shutdown()

	raw := []byte(`{"id":"g3","type":"gauge","value":3.45}`)
	body, encHeaders := encryptBody(t, raw)
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Encrypted":  encHeaders["X-Encrypted"],
	}
	resp, err := tester.DoRequest(http.MethodPost, "/update", body, headers)
	if err != nil {
		t.Fatalf("post update: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got: %d", resp.StatusCode)
	}
}

func TestUpdatePlain_NoEncryption_Success(t *testing.T) {
	opts := map[string]any{
		"Key":       "",
		"CryptoKey": privateKeyPath,
	}
	tester, err := NewTester(t, &opts)
	if err != nil {
		t.Fatalf("tester init: %v", err)
	}
	defer tester.Shutdown()

	body := []byte(`{"id":"g0","type":"gauge","value":2.5}`)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := tester.DoRequest(http.MethodPost, "/update", body, headers)
	if err != nil {
		t.Fatalf("post update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	getResp, err := tester.Get("/value/gauge/g0")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected get status: %d", getResp.StatusCode)
	}
	val, err := tester.ReadGzip(getResp)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(val) != "2.5" {
		t.Fatalf("unexpected value: %s", string(val))
	}
}

func TestUpdatesEncrypted_NoGzip_Success(t *testing.T) {
	opts := map[string]any{
		"Key":       "",
		"CryptoKey": privateKeyPath,
	}
	tester, err := NewTester(t, &opts)
	if err != nil {
		t.Fatalf("tester init: %v", err)
	}
	defer tester.Shutdown()

	raw := []byte(`[{"id":"bg","type":"gauge","value":1.1},{"id":"bc","type":"counter","delta":3}]`)
	body, encHeaders := encryptBody(t, raw)
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Encrypted":  encHeaders["X-Encrypted"],
	}
	resp, err := tester.DoRequest(http.MethodPost, "/updates", body, headers)
	if err != nil {
		t.Fatalf("post updates: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	getResp, err := tester.Get("/value/gauge/bg")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected get status: %d", getResp.StatusCode)
	}
	val, err := tester.ReadGzip(getResp)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(val) != "1.1" {
		t.Fatalf("unexpected value: %s", string(val))
	}

	getResp2, err := tester.Get("/value/counter/bc")
	if err != nil {
		t.Fatalf("get value counter: %v", err)
	}
	if getResp2.StatusCode != http.StatusOK {
		t.Fatalf("unexpected get status: %d", getResp2.StatusCode)
	}
	val2, err := tester.ReadGzip(getResp2)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(val2) != "3" {
		t.Fatalf("unexpected counter value: %s", string(val2))
	}
}
