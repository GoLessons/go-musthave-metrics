package audit

import (
	"testing"

	"github.com/goccy/go-json"
)

func TestJournalItemJSON(t *testing.T) {
	item := NewJournalItem(12345678, []string{"Alloc", "Frees"}, "192.168.0.42")
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	expected := `{"ts":12345678,"metrics":["Alloc","Frees"],"ip_address":"192.168.0.42"}`
	if string(data) != expected {
		t.Fatalf("unexpected json:\nexpected: %s\ngot:      %s", expected, string(data))
	}
}
