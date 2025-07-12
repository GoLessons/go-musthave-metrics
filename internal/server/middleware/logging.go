package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

func NewLoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для ResponseWriter
			rw := &loggingDecorator{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Логируем информацию о запросе и ответе
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Duration("duration", duration),
				zap.Int("status_code", rw.statusCode),
				zap.Int("response_size", rw.size),
			)
		})
	}
}

type loggingDecorator struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *loggingDecorator) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingDecorator) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
