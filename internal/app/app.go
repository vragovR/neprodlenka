package app

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"neprodlenka/internal/api"
	"neprodlenka/internal/auth"
	"neprodlenka/internal/bot"
	"neprodlenka/internal/clock"
	"neprodlenka/internal/config"
	"neprodlenka/internal/db"
	"neprodlenka/internal/httpserver"
	"neprodlenka/internal/logger"
	"neprodlenka/internal/maxapi"
	"neprodlenka/internal/message"
	"neprodlenka/internal/store"
)

var Module = fx.Options(
	fx.Provide(
		config.New,
		logger.New,
		clock.New,
		db.New,
		httpserver.New,
		api.NewHandler,
		auth.New,
		store.New,
		message.NewBuilder,
		maxapi.New,
		func(c *maxapi.Client) bot.MAX { return c },
		bot.NewRedrawer,
		bot.New,
		bot.NewWebhook,
	),
	fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		l := &fxevent.ZapLogger{Logger: log}
		l.UseLogLevel(zapcore.DebugLevel)
		return l
	}),
)
