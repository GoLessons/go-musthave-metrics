package audit

import (
	"bufio"
	"context"
	"github.com/goccy/go-json"
	"os"
	"testing"
	"time"
)

func TestFileAuditor_WriteAndAppend(t *testing.T) {
	tmp, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatalf("temp file error: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	a := NewFileAuditor(tmp.Name())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	item1 := NewJournalItem(1, []string{"A"}, "127.0.0.1")
	if ok := a.Journal(ctx, item1); !ok {
		t.Fatalf("first write failed")
	}

	item2 := NewJournalItem(2, []string{"B"}, "127.0.0.2")
	if ok := a.Journal(ctx, item2); !ok {
		t.Fatalf("second write failed")
	}

	f, err := os.Open(tmp.Name())
	if err != nil {
		t.Fatalf("open file error: %v", err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var got1, got2 JournalItem
	if err := json.Unmarshal([]byte(lines[0]), &got1); err != nil {
		t.Fatalf("unmarshal line1 error: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &got2); err != nil {
		t.Fatalf("unmarshal line2 error: %v", err)
	}

	if got1.TS != 1 || got1.IP != "127.0.0.1" || len(got1.Metrics) != 1 || got1.Metrics[0] != "A" {
		t.Fatalf("line1 content mismatch: %+v", got1)
	}
	if got2.TS != 2 || got2.IP != "127.0.0.2" || len(got2.Metrics) != 1 || got2.Metrics[0] != "B" {
		t.Fatalf("line2 content mismatch: %+v", got2)
	}
}

func TestFileAuditor_EmptyPath(t *testing.T) {
	a := NewFileAuditor("")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ok := a.Journal(ctx, NewJournalItem(1, []string{"A"}, "127.0.0.1"))
	if ok {
		t.Fatalf("expected false for empty path")
	}
}
