package bot

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/message"
)

const (
	editInterval = 500 * time.Millisecond
	retryDelay   = 5 * time.Second
)

type Redrawer struct {
	builder *message.Builder
	max     MAX
	log     *zap.Logger

	mu      sync.Mutex
	pending map[string]time.Time
	wake    chan struct{}
}

func NewRedrawer(lc fx.Lifecycle, b *message.Builder, m MAX, log *zap.Logger) *Redrawer {
	r := &Redrawer{
		builder: b,
		max:     m,
		log:     log,
		pending: map[string]time.Time{},
		wake:    make(chan struct{}, 1),
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				r.run(ctx)
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			<-done
			return nil
		},
	})

	return r
}

func (r *Redrawer) Redraw(date time.Time) {
	r.mu.Lock()
	r.pending[date.Format(time.DateOnly)] = date
	r.mu.Unlock()

	select {
	case r.wake <- struct{}{}:
	default:
	}
}

func (r *Redrawer) run(ctx context.Context) {
	for {
		date, ok := r.next()
		if !ok {
			select {
			case <-ctx.Done():
				return
			case <-r.wake:
				continue
			}
		}

		delay := editInterval
		if err := r.redraw(ctx, date); err != nil {
			r.log.Error("redraw", zap.Time("date", date), zap.Error(err))
			r.Redraw(date)
			delay = retryDelay
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (r *Redrawer) next() (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, date := range r.pending {
		delete(r.pending, key)
		return date, true
	}
	return time.Time{}, false
}

func (r *Redrawer) redraw(ctx context.Context, date time.Time) error {
	return errors.Join(r.redrawDay(ctx, date), r.redrawMonth(ctx, clock.MonthOf(date)))
}

func (r *Redrawer) redrawDay(ctx context.Context, date time.Time) error {
	msg, post, ok, err := r.builder.Build(ctx, date)
	if err != nil {
		return err
	}
	if !ok || post.MessageID == "" {
		return nil
	}
	return r.max.Edit(ctx, post.MessageID, msg)
}

func (r *Redrawer) redrawMonth(ctx context.Context, month time.Time) error {
	msg, rm, ok, err := r.builder.BuildMonth(ctx, month)
	if err != nil {
		return err
	}
	if !ok || rm.MessageID == "" {
		return nil
	}
	return r.max.Edit(ctx, rm.MessageID, msg)
}
