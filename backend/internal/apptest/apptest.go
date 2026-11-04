package apptest

import (
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"neprodlenka/internal/app"
	"neprodlenka/internal/bot"
	"neprodlenka/internal/clock"
	"neprodlenka/internal/config"
)

type App struct {
	DB  *sql.DB
	MAX *FakeMAX

	mu  sync.Mutex
	now time.Time
}

func New(t *testing.T, targets ...any) *App {
	t.Helper()

	cfg := config.Config{
		DBPath:           filepath.Join(t.TempDir(), "test.db"),
		PostTime:         config.TimeOfDay{Hour: 7},
		PublishUntil:     config.TimeOfDay{Hour: 8},
		CloseTime:        config.TimeOfDay{Hour: 9},
		CloseUntil:       config.TimeOfDay{Hour: 10},
		SummaryTime:      config.TimeOfDay{Hour: 12},
		SummaryUntil:     config.TimeOfDay{Hour: 13},
		JWTSecret:        "test-secret-test-secret-test-secret",
		BaseURL:          "https://school.test",
		MaxChatID:        -100,
		WebhookSecret:    "test-webhook-secret",
		PaymentCard:      "2202 0000 0000 0000",
		PaymentRecipient: "Иванова Мария Петровна",
	}

	a := &App{MAX: NewFakeMAX(), now: time.Now()}
	fxApp := fxtest.New(t,
		app.Module,
		fx.Replace(cfg, zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))),
		fx.Replace(fx.Annotate(a.MAX, fx.As(new(bot.MAX)))),
		fx.Replace(clock.Clock(a.Now)),
		fx.Populate(append(targets, &a.DB)...),
	)
	fxApp.RequireStart()
	t.Cleanup(fxApp.RequireStop)

	return a
}

func (a *App) Now() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.now
}

func (a *App) SetNow(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.now = t
}

func (a *App) Exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := a.DB.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
