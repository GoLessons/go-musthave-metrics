package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGzipMiddleware_Compress(b *testing.B) {
	payload := bytes.Repeat([]byte("x"), 16*1024)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	})
	mw := GzipMiddleware(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/value/gauge/m", nil)
		r.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}

func BenchmarkGzipMiddleware_Decompress(b *testing.B) {
	raw := bytes.Repeat([]byte(`{"id":"m","type":"gauge","value":1}`), 512)
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	_, _ = gzw.Write(raw)
	_ = gzw.Close()
	compressed := buf.Bytes()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
	})
	mw := GzipMiddleware(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(compressed))
		r.Header.Set("Content-Encoding", "gzip")
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}
