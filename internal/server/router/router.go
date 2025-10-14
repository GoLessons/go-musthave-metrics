package router

import (
	"database/sql"
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/handler"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func RouterFactory() container.Factory[*chi.Mux] {
	return func(c container.Container) (*chi.Mux, error) {
		logger, err := container.GetService[zap.Logger](c, "logger")
		if err != nil {
			return nil, err
		}

		counterStorage, err := container.GetService[storage.MemStorage[serverModel.Counter]](c, "counterStorage")
		if err != nil {
			return nil, err
		}

		gaugeStorage, err := container.GetService[storage.MemStorage[serverModel.Gauge]](c, "gaugeStorage")
		if err != nil {
			return nil, err
		}

		metricService, err := container.GetService[service.MetricService](c, "metricService")
		if err != nil {
			return nil, err
		}

		db, err := container.GetService[sql.DB](c, "db")
		if err != nil {
			return nil, err
		}

		cfg, err := container.GetService[config.Config](c, "config")
		if err != nil {
			return nil, err
		}

		r := chi.NewRouter()

		metricControllerJSON := handler.NewMetricsController(*metricService, handler.JSONResposeBuilder, logger)
		metricControllerPlain := handler.NewMetricsController(*metricService, handler.PlainResposeBuilder, logger)

		var signatureMiddleware *middleware.SignatureMiddleware
		if cfg.Key != "" {
			signer := signature.NewSign(cfg.Key)
			signatureMiddleware = middleware.NewSignatureMiddleware(signer, logger)
		}

		r.Route("/update/{metricType}/{metricName:[a-zA-Z0-9_-]+}/{metricValue:(-?)[a-z0-9\\.]+}",
			func(r chi.Router) {
				r.Use(middleware.MetricCtxFromPath)
				if signatureMiddleware != nil {
					r.Use(signatureMiddleware.AddSignature)
				}

				r.Post("/", metricControllerPlain.Update)
			},
		)

		r.Route("/value/{metricType}/{metricName:[a-zA-Z0-9_-]+}", func(r chi.Router) {
			r.Use(middleware.GzipMiddleware)
			r.Use(middleware.MetricCtxFromPath)
			if signatureMiddleware != nil {
				r.Use(signatureMiddleware.AddSignature)
			}

			r.Get("/", metricControllerPlain.Get)
		})

		r.Route("/update", func(r chi.Router) {
			r.Use(middleware.ValidateRoute)
			if signatureMiddleware != nil {
				r.Use(signatureMiddleware.VerifySignature)
			}
			r.Use(middleware.GzipMiddleware)
			r.Use(middleware.MetricCtxFromBody)
			if signatureMiddleware != nil {
				r.Use(signatureMiddleware.AddSignature)
			}

			r.Post("/", metricControllerJSON.Update)
			r.Post("/.+", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Not Found", http.StatusNotFound)
			})
		})

		r.Route("/updates", func(r chi.Router) {
			r.Use(middleware.ValidateRoute)
			if signatureMiddleware != nil {
				r.Use(signatureMiddleware.VerifySignature)
			}
			r.Use(middleware.GzipMiddleware)
			r.Use(middleware.MetricsListCtxFromBody)
			if signatureMiddleware != nil {
				r.Use(signatureMiddleware.AddSignature)
			}

			r.Post("/", metricControllerJSON.UpdateBatch)
			r.Post("/.+", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Not Found", http.StatusNotFound)
			})
		})

		r.Route("/value",
			func(r chi.Router) {
				if signatureMiddleware != nil {
					r.Use(signatureMiddleware.VerifySignature)
				}
				r.Use(middleware.GzipMiddleware)
				r.Use(middleware.MetricCtxFromBody)
				if signatureMiddleware != nil {
					r.Use(signatureMiddleware.AddSignature)
				}

				r.Post("/", metricControllerJSON.Get)
			},
		)

		r.Route("/ping",
			func(r chi.Router) {
				r.Get("/", handler.NewPingHandler(db, logger).Ping)
			},
		)

		r.Route("/",
			func(r chi.Router) {
				r.Use(middleware.GzipMiddleware)
				r.Get(
					"/",
					handler.NewListController(
						counterStorage,
						gaugeStorage,
					).Get,
				)
			},
		)

		return r, nil
	}
}
