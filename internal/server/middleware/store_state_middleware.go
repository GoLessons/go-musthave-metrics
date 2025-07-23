package middleware

import (
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"net/http"
	"sync"
	"time"
)

type storeStateMiddleware struct {
	metricService *service.MetricService
	metricDumper  service.MetricDumper
	storeInterval time.Duration
	lastStoreTime time.Time
	mutex         sync.Mutex
}

func NewStoreStateMiddleware(
	metricService *service.MetricService,
	metricDumper service.MetricDumper,
	storeInterval uint64,
) *storeStateMiddleware {
	return &storeStateMiddleware{
		metricService: metricService,
		metricDumper:  metricDumper,
		storeInterval: time.Duration(storeInterval) * time.Second,
		lastStoreTime: time.Now(),
	}
}

func (m *storeStateMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		m.mutex.Lock()
		defer m.mutex.Unlock()

		if time.Since(m.lastStoreTime) >= m.storeInterval {
			err := service.StoreState(m.metricService, m.metricDumper)
			if err == nil {
				m.lastStoreTime = time.Now()
			}
		}
	})
}
