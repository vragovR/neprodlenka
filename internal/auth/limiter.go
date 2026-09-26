package auth

import (
	"sync"
	"time"
)

type limiter struct {
	window   time.Duration
	mu       sync.Mutex
	failures map[string][]time.Time
}

func newLimiter(window time.Duration) *limiter {
	return &limiter{window: window, failures: map[string][]time.Time{}}
}

func (l *limiter) allowed(now time.Time, key string, limit int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recent(now, key)) < limit
}

func (l *limiter) fail(now time.Time, keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, key := range keys {
		l.failures[key] = append(l.recent(now, key), now)
	}
	for key := range l.failures {
		l.recent(now, key)
	}
}

func (l *limiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

func (l *limiter) recent(now time.Time, key string) []time.Time {
	times := l.failures[key]
	i := 0
	for i < len(times) && !now.Before(times[i].Add(l.window)) {
		i++
	}
	if i == len(times) {
		delete(l.failures, key)
		return nil
	}
	l.failures[key] = times[i:]
	return times[i:]
}
