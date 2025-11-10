package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"go.uber.org/zap"
)

func BenchmarkSignatureMiddleware_Verify(b *testing.B) {
	signer := signature.NewSign("bench_key")
	m := NewSignatureMiddleware(signer, zap.NewNop())
	body := bytes.Repeat([]byte(`{"id":"m","type":"counter","delta":1}`), 128)
	hash, _ := signer.Hash(body)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mw := m.VerifySignature(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		r.Header.Set(m.HashHeader, hash)
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}

func BenchmarkSignatureMiddleware_Add(b *testing.B) {
	signer := signature.NewSign("bench_key")
	m := NewSignatureMiddleware(signer, zap.NewNop())
	body := bytes.Repeat([]byte("resp-body"), 2048)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	})
	mw := m.AddSignature(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/value/gauge/m", nil)
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}
