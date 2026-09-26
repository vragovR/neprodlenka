package bot

import (
	"context"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"neprodlenka/internal/config"
	"neprodlenka/internal/maxapi"
)

const (
	maxTickWait = time.Hour
	tickRetry   = time.Minute
)

func Start(lc fx.Lifecycle, api *maxapi.Client, b *Bot, cfg config.Config, log *zap.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			wg.Go(func() { schedule(ctx, b, log) })
			if cfg.Mode == config.ModeWebhook {
				wg.Go(func() { subscribe(ctx, api, cfg, log) })
			} else {
				wg.Go(func() { poll(ctx, api, b, log) })
			}
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			wg.Wait()
			return nil
		},
	})
}

func schedule(ctx context.Context, b *Bot, log *zap.Logger) {
	for {
		err := b.Tick(ctx)
		wait := min(time.Until(b.nextTick(b.now())), maxTickWait)
		if err != nil && ctx.Err() == nil {
			log.Error("tick, retry in a minute", zap.Error(err))
			wait = tickRetry
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

func subscribe(ctx context.Context, api *maxapi.Client, cfg config.Config, log *zap.Logger) {
	for {
		err := api.Subscribe(ctx, cfg.WebhookURL, cfg.WebhookSecret)
		if err == nil {
			log.Info("webhook subscribed", zap.String("url", cfg.WebhookURL))
			return
		}
		if ctx.Err() != nil {
			return
		}
		log.Error("subscribe webhook, retry in a minute", zap.Error(err))
		select {
		case <-ctx.Done():
			return
		case <-time.After(tickRetry):
		}
	}
}

func poll(ctx context.Context, api *maxapi.Client, b *Bot, log *zap.Logger) {
	var marker int64
	for {
		updates, next, err := api.Updates(ctx, marker)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Error("get updates", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryDelay):
			}
			continue
		}
		marker = next

		for _, u := range updates {
			b.HandleUpdate(ctx, u)
		}
	}
}
