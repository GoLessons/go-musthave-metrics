package audit

import (
	"context"
	"sync"
)

// Subject — издатель: управляет подпиской и уведомлениями наблюдателей.
type Subject interface {
	Register(o Auditor)
	Deregister(o Auditor)
	NotifyAll(ctx context.Context, item *JournalItem) bool
}

type AuditSubject struct {
	mu        sync.RWMutex
	observers []Auditor
}

func NewAuditSubject(observers ...Auditor) *AuditSubject {
	return &AuditSubject{observers: observers}
}

func (s *AuditSubject) Register(o Auditor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, o)
}

func (s *AuditSubject) Deregister(o Auditor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obs := s.observers
	for i, current := range obs {
		if current == o {
			obs[len(obs)-1], obs[i] = obs[i], obs[len(obs)-1]
			s.observers = obs[:len(obs)-1]
			return
		}
	}
}

func (s *AuditSubject) NotifyAll(ctx context.Context, item *JournalItem) bool {
	if item == nil {
		return false
	}

	// Снимок списка наблюдателей под RLock, чтобы не держать блокировку во время уведомлений.
	s.mu.RLock()
	observers := make([]Auditor, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	if len(observers) == 0 {
		return false
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(observers))

	for _, o := range observers {
		wg.Add(1)
		go func(obs Auditor) {
			defer wg.Done()
			if ctx.Err() != nil {
				results <- false
				return
			}
			results <- obs.Journal(ctx, item)
		}(o)
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

func (s *AuditSubject) Journal(ctx context.Context, item *JournalItem) bool {
	return s.NotifyAll(ctx, item)
}
