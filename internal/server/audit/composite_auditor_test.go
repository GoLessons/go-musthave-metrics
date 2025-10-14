package audit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type stubAuditor struct {
	result bool
	delay  time.Duration
	calls  int32
}

func (s *stubAuditor) Journal(ctx context.Context, item *JournalItem) bool {
	atomic.AddInt32(&s.calls, 1)
	select {
	case <-ctx.Done():
		return false
	case <-time.After(s.delay):
		return s.result
	}
}

func TestCompositeAuditor_AggregatesResultsAndCallsAll(t *testing.T) {
	a1 := &stubAuditor{result: false, delay: 100 * time.Millisecond}
	a2 := &stubAuditor{result: true, delay: 50 * time.Millisecond}
	a3 := &stubAuditor{result: false, delay: 150 * time.Millisecond}

	comp := NewCompositeAuditor(a1, a2, a3)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ok := comp.Journal(ctx, NewJournalItem(1, []string{"A"}, "127.0.0.1"))
	elapsed := time.Since(start)

	if !ok {
		t.Fatalf("expected true when at least one succeeds")
	}

	if atomic.LoadInt32(&a1.calls) != 1 || atomic.LoadInt32(&a2.calls) != 1 || atomic.LoadInt32(&a3.calls) != 1 {
		t.Fatalf("expected all child auditors to be called")
	}

	if elapsed > 300*time.Millisecond {
		t.Fatalf("composite should run children concurrently, elapsed: %v", elapsed)
	}
}

func TestCompositeAuditor_NoChildrenOrNilItem(t *testing.T) {
	comp := NewCompositeAuditor()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if ok := comp.Journal(ctx, nil); ok {
		t.Fatalf("expected false for nil item")
	}
}
