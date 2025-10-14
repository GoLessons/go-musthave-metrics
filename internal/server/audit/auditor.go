package audit

import (
	"context"
)

type Auditor interface {
	Journal(context.Context, *JournalItem) bool
}

type JournalItem struct {
	Ts        int64    `json:"ts"`         // unix timestamp события
	Metrics   []string `json:"metrics"`    // наименование полученных метрик
	IpAddress string   `json:"ip_address"` // IP адрес входящего запроса
}

func NewJournalItem(ts int64, metrics []string, ipAddress string) *JournalItem {
	return &JournalItem{
		Ts:        ts,
		Metrics:   metrics,
		IpAddress: ipAddress,
	}
}
