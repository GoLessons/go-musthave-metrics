package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe(`:8080`, router())
}

func router() *chi.Mux {
	r := chi.NewRouter()
	r.Route("/update/counter/{metricName:.*}", func(r chi.Router) {
		r.Use(MetricCtx)
		r.Get("/", updateCounter)
		r.Post("/", updateGauge)
	})
	r.Route("/update/gauge/{metricName:.*}", func(r chi.Router) {
		r.Use(MetricCtx)
		r.Get("/", updateGauge)
		r.Post("/", updateGauge)
	})

	return r
}

func MetricCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		ctx := r.Context()
		ctx = context.WithValue(r.Context(), "metricName", metricName)
		next.ServeHTTP(w, r.WithContext(ctx))
		//metricType := chi.URLParam(r, "metricType")
		/*article, err := dbGetArticle(articleID)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "article", metric)
		next.ServeHTTP(w, r.WithContext(ctx))*/
	})
}

func updateCounter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricName, ok := ctx.Value("metricName").(string)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Write([]byte(fmt.Sprintf("Update counter: %s", metricName)))
}

func updateGauge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricName, ok := ctx.Value("metricName").(string)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Write([]byte(fmt.Sprintf("Update gauge: %s", metricName)))
}
