package handler

import (
	"database/sql"
	"go.uber.org/zap"
	"net/http"
)

type pingHandler struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPingHandler(db *sql.DB, logger *zap.Logger) *pingHandler {
	return &pingHandler{db: db, logger: logger}
}

func (h *pingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		h.logger.Error("DB ping failed", zap.Error(err))
		http.Error(w, "DB ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
