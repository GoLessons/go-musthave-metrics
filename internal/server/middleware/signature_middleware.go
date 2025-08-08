package middleware

import (
	"bytes"
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type SignatureMiddleware struct {
	signer     *signature.Signer
	HashHeader string
	logger     *zap.Logger
}

type signatureWriter struct {
	http.ResponseWriter
	signer     *signature.Signer
	body       bytes.Buffer
	hashHeader string
	headerSent bool
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
			http.Error(w, "Signature is missing", http.StatusBadRequest)
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
		next.ServeHTTP(w, r)
	})
}
