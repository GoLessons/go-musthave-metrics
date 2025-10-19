package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"go.uber.org/zap"
)

type SignatureMiddleware struct {
	signer     *signature.Signer
	HashHeader string
	logger     *zap.Logger
}

type signatureWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *signatureWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *signatureWriter) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}

func NewSignatureMiddleware(signer *signature.Signer, logger *zap.Logger) *SignatureMiddleware {
	return &SignatureMiddleware{
		signer:     signer,
		HashHeader: "HashSHA256",
		logger:     logger,
	}
}

func (m *SignatureMiddleware) VerifySignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hash := r.Header.Get(m.HashHeader)
		if hash == "" {
			m.logger.Error("no hash header", zap.String("hashHeader", m.HashHeader))
			next.ServeHTTP(w, r)
			//http.Error(w, "Signature is missing", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			m.logger.Error("failed to read body", zap.Error(err))
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if !m.signer.Check(hash, body) {
			m.logger.Error("invalid signature", zap.String("hashHeader", m.HashHeader), zap.String("signature", hash), zap.String("body", string(body)))
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *SignatureMiddleware) AddSignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &signatureWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		if len(rec.body) == 0 {
			w.WriteHeader(rec.statusCode)
			_, _ = w.Write(rec.body)
			return
		}

		hash, err := m.signer.Hash(rec.body)
		if err != nil {
			http.Error(w, "failed to sign response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HashSHA256", hash)

		w.WriteHeader(rec.statusCode)
		_, _ = w.Write(rec.body)
	})
}
