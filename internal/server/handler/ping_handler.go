package handler

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

type pingHandler struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPingHandler(db *sql.DB, logger *zap.Logger) *pingHandler {
	return &pingHandler{db: db, logger: logger}
}

func (h *pingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		http.Error(w, "DB not configured", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
