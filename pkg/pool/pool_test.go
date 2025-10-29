package pool

import (
	"bytes"
	"testing"
)

type TestItem struct {
	value      int
	resetCount int
}

func (t *TestItem) Reset() {
	t.value = 0
	t.resetCount++
}

func TestNewAndGetCreates(t *testing.T) {
	p := New(func() *TestItem {
		return &TestItem{value: 10}
	})
	obj := p.Get()
	if obj == nil {
		t.Fatalf("Get returned nil")
	}
	if obj.value != 10 {
		t.Fatalf("expected value=10, got %d", obj.value)
	}
	if obj.resetCount != 0 {
		t.Fatalf("expected resetCount=0, got %d", obj.resetCount)
	}
}

func TestPutResets(t *testing.T) {
	p := New(func() *TestItem { return &TestItem{} })
	obj := p.Get()
	obj.value = 42

	p.Put(obj)

	if obj.value != 0 {
		t.Fatalf("expected value reset to 0, got %d", obj.value)
	}
	if obj.resetCount != 1 {
		t.Fatalf("expected resetCount=1 after Put, got %d", obj.resetCount)
	}

	// Дополнительно: следующий Get возвращает объект с "нулевым" состоянием.
	obj2 := p.Get()
	if obj2 != nil && obj2.value != 0 {
		t.Fatalf("expected value=0 on Get after Put, got %d", obj2.value)
	}
}

func TestPutNilPointer(t *testing.T) {
	p := New(func() *TestItem { return &TestItem{value: 7} })

	p.Put(nil)

	got := p.Get()
	if got == nil {
		t.Fatalf("expected non-nil object from Get")
	}
	if got.value != 7 || got.resetCount != 0 {
		t.Fatalf("unexpected object from Get: value=%d resetCount=%d", got.value, got.resetCount)
	}
}

func TestBytesBufferReset(t *testing.T) {
	p := New(func() *bytes.Buffer { return &bytes.Buffer{} })

	buf := p.Get()
	if buf == nil {
		t.Fatalf("Get returned nil buffer")
	}
	_, _ = buf.WriteString("hello")
	if buf.Len() == 0 {
		t.Fatalf("expected buffer to have content before Put")
	}

	p.Put(buf)

	if buf.Len() != 0 {
		t.Fatalf("buffer should be reset after Put, got len=%d", buf.Len())
	}

	buf2 := p.Get()
	if buf2.Len() != 0 {
		t.Fatalf("expected clean buffer on Get, got len=%d", buf2.Len())
	}
}
