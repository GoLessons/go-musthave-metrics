package audit

import (
	"context"
	"sync"
)

type CompositeAuditor struct {
	childs []Auditor
}

func (a CompositeAuditor) Journal(ctx context.Context, item *JournalItem) bool {
	if item == nil || len(a.childs) == 0 {
		return false
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(a.childs))

	for _, child := range a.childs {
		wg.Add(1)
		go func(ch Auditor) {
			defer wg.Done()
			results <- ch.Journal(ctx, item)
		}(child)
	}

	wg.Wait()
	close(results)

	for r := range results {
		if r {
			return true
		}
	}
	return false
}

func NewCompositeAuditor(childs ...Auditor) *CompositeAuditor {
	return &CompositeAuditor{childs: childs}
}
