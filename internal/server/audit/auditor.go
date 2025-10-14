package audit

import (
	"context"
)

type Auditor interface {
	Journal(context.Context, *JournalItem) bool
}

type JournalItem struct {
	TS      int64    `json:"ts"`      // unix timestamp события
	Metrics []string `json:"metrics"` // наименование полученных метрик
	IP      string   `json:"ip"`      // IP адрес входящего запроса
}

func NewJournalItem(ts int64, metrics []string, ipAddress string) *JournalItem {
	return &JournalItem{
		TS:      ts,
		Metrics: metrics,
		IP:      ipAddress,
	}
}
