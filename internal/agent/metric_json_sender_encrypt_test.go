package agent

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
)

func TestJSONSender_EncryptsAndSigns_WithGzip(t *testing.T) {
	var capturedBody []byte
	var capturedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		capturedBody = b
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	addr := ts.Listener.Addr().String()
	signer := signature.NewSign("test-secret")
	certPath := filepath.Join("..", "..", "var", "keys", "cert.pem")
	encrypter, err := NewEncrypterFromFile(certPath)
	if err != nil {
		t.Fatalf("encrypter error: %v", err)
	}

	sender := NewJSONSender(addr, true, signer, encrypter)
	defer sender.Close()

	val := 1.23
	delta := int64(10)
	metrics := []model.Metrics{
		*model.NewGauge("g", &val),
		*model.NewCounter("c", &delta),
	}
	if err := sender.SendBatch(metrics); err != nil {
		t.Fatalf("SendBatch error: %v", err)
	}

	if capturedHeaders.Get("X-Encrypted") != "aes256gcm+rsa-oaep;v=1" {
		t.Fatalf("missing or invalid X-Encrypted header")
	}
	if capturedHeaders.Get("Content-Encoding") != "gzip" {
		t.Fatalf("missing gzip Content-Encoding")
	}

	sig := capturedHeaders.Get("HashSHA256")
	if sig == "" {
		t.Fatalf("missing HashSHA256 header")
	}
	expected, err := signer.Hash(capturedBody)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if sig != expected {
		t.Fatalf("signature mismatch")
	}

	gr, err := gzip.NewReader(bytes.NewReader(capturedBody))
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	defer gr.Close()
	raw, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("gzip read error: %v", err)
	}

	var cont encryptedContainer
	if err := json.Unmarshal(raw, &cont); err != nil {
		t.Fatalf("container unmarshal error: %v", err)
	}
	if cont.Alg != "aes256gcm+rsa-oaep" || cont.K == "" || cont.N == "" || cont.D == "" || cont.V != 1 {
		t.Fatalf("invalid encrypted container")
	}
}

func TestJSONSender_EncryptsAndSigns_NoGzip(t *testing.T) {
	var capturedBody []byte
	var capturedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		capturedBody = b
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	addr := ts.Listener.Addr().String()
	signer := signature.NewSign("test-secret")
	certPath := filepath.Join("..", "..", "var", "keys", "cert.pem")
	encrypter, err := NewEncrypterFromFile(certPath)
	if err != nil {
		t.Fatalf("encrypter error: %v", err)
	}

	sender := NewJSONSender(addr, false, signer, encrypter)
	defer sender.Close()

	val := 2.34
	metrics := []model.Metrics{*model.NewGauge("g2", &val)}
	if err := sender.SendBatch(metrics); err != nil {
		t.Fatalf("SendBatch error: %v", err)
	}

	if capturedHeaders.Get("X-Encrypted") != "aes256gcm+rsa-oaep;v=1" {
		t.Fatalf("missing or invalid X-Encrypted header")
	}
	if capturedHeaders.Get("Content-Encoding") != "" {
		t.Fatalf("unexpected Content-Encoding")
	}

	sig := capturedHeaders.Get("HashSHA256")
	if sig == "" {
		t.Fatalf("missing HashSHA256 header")
	}
	expected, err := signer.Hash(capturedBody)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if sig != expected {
		t.Fatalf("signature mismatch")
	}

	var cont encryptedContainer
	if err := json.Unmarshal(capturedBody, &cont); err != nil {
		t.Fatalf("container unmarshal error: %v", err)
	}
	if cont.Alg != "aes256gcm+rsa-oaep" || cont.K == "" || cont.N == "" || cont.D == "" || cont.V != 1 {
		t.Fatalf("invalid encrypted container")
	}
}
