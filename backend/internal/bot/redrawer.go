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
	pending map[string]job
	wake    chan struct{}
}

type job struct {
	date time.Time
	day  bool
}

func NewRedrawer(lc fx.Lifecycle, b *message.Builder, m MAX, log *zap.Logger) *Redrawer {
	r := &Redrawer{
		builder: b,
		max:     m,
		log:     log,
		pending: map[string]job{},
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
	r.add(date.Format(time.DateOnly), job{date: date, day: true})
}

func (r *Redrawer) RedrawMonth(month time.Time) {
	r.add(month.Format("2006-01"), job{date: month})
}

func (r *Redrawer) add(key string, j job) {
	r.mu.Lock()
	r.pending[key] = j
	r.mu.Unlock()

	select {
	case r.wake <- struct{}{}:
	default:
	}
}

func (r *Redrawer) run(ctx context.Context) {
	for {
		j, ok := r.next()
		if !ok {
			select {
			case <-ctx.Done():
				return
			case <-r.wake:
				continue
			}
		}

		if err := r.redraw(ctx, j); err != nil {
			r.log.Error("redraw", zap.Time("date", j.date), zap.Bool("day", j.day), zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(editInterval):
		}
	}
}

func (r *Redrawer) next() (job, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, j := range r.pending {
		delete(r.pending, key)
		return j, true
	}
	return job{}, false
}

func (r *Redrawer) redraw(ctx context.Context, j job) error {
	if !j.day {
		return r.redrawMonth(ctx, j.date)
	}
	return errors.Join(r.redrawDay(ctx, j.date), r.redrawMonth(ctx, clock.MonthOf(j.date)))
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
