package audit

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

func TestURLAuditor_SendsJSONAndSuccess(t *testing.T) {
	var received []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("unexpected content-type: %s", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		received = body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := NewURLAuditor(srv.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	item := NewJournalItem(12345678, []string{"Alloc", "Frees"}, "192.168.0.42")
	ok := a.Journal(ctx, item)
	if !ok {
		t.Fatalf("expected success")
	}

	data, _ := json.Marshal(item)
	if string(received) != string(data) {
		t.Fatalf("unexpected payload:\nexpected: %s\ngot:      %s", string(data), string(received))
	}
}

func TestURLAuditor_FailureStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := NewURLAuditor(srv.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ok := a.Journal(ctx, NewJournalItem(1, []string{"A"}, "127.0.0.1"))
	if ok {
		t.Fatalf("expected failure for 5xx response")
	}
}

func TestURLAuditor_EmptyURL(t *testing.T) {
	a := NewURLAuditor("", nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ok := a.Journal(ctx, NewJournalItem(1, []string{"A"}, "127.0.0.1"))
	if ok {
		t.Fatalf("expected false for empty url")
	}
}
