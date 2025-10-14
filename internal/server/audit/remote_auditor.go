package audit

import (
	"context"
)

type RemoteAuditor struct {
	url string
}

func NewRemoteAuditor(url string) *RemoteAuditor {
	return &RemoteAuditor{
		url: url,
	}
}

func (a *RemoteAuditor) Journal(ctx context.Context, item *JournalItem) bool {
	return true
}
