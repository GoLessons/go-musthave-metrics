package audit

import (
	"context"
)

type Auditor interface {
	Journal(context.Context, *JournalItem) bool
}

type JournalItem struct {
	TS      int64    `json:"ts"`
	Metrics []string `json:"metrics"`
	IP      string   `json:"ip_address"`
}

func NewJournalItem(ts int64, metrics []string, ipAddress string) *JournalItem {
	return &JournalItem{
		TS:      ts,
		Metrics: metrics,
		IP:      ipAddress,
	}
}
