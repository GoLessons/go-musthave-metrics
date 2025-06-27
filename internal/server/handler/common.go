package handler

import "net/http"

type MetricController interface {
	Get(http.ResponseWriter, *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}
