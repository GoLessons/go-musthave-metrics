package audit

import (
	"context"
	"github.com/goccy/go-json"
	"resty.dev/v3"
	"time"
)

type RemoteAuditor struct {
	url    string
	client *resty.Client
}

func NewRemoteAuditor(url string, client *resty.Client) *RemoteAuditor {
	if client == nil {
		client = resty.New()
		client.SetTimeout(5 * time.Second)
	}
	return &RemoteAuditor{url: url, client: client}
}

func (a *RemoteAuditor) Journal(ctx context.Context, item *JournalItem) bool {
	if a.url == "" || item == nil {
		return false
	}

	data, err := json.Marshal(item)
	if err != nil {
		return false
	}

	resp, err := a.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(a.url)
	if err != nil {
		return false
	}

	return resp.IsSuccess()
}
