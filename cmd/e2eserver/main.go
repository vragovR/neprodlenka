package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"go.uber.org/fx"

	"neprodlenka/internal/app"
	"neprodlenka/internal/auth"
	"neprodlenka/internal/bot"
	"neprodlenka/internal/clock"
	"neprodlenka/internal/config"
	"neprodlenka/internal/maxapi"
	"neprodlenka/internal/message"
	"neprodlenka/internal/store"
)

const (
	adminLogin    = "admin"
	adminPassword = "admin-password"
)

var args struct {
	Addr    string    `arg:"--addr,required"`
	DB      string    `arg:"--db,required"`
	Now     time.Time `arg:"--now,required"`
	Members string    `arg:"--members" default:"[]"`
}

type fakeMAX struct {
	members []maxapi.Member
}

func (f fakeMAX) Send(context.Context, message.Message) (string, error) { return "m", nil }
func (f fakeMAX) Edit(context.Context, string, message.Message) error   { return nil }
func (f fakeMAX) Delete(context.Context, string) error                  { return nil }
func (f fakeMAX) AnswerCallback(context.Context, string, string) error  { return nil }
func (f fakeMAX) Members(context.Context) ([]maxapi.Member, error)      { return f.members, nil }

func main() {
	arg.MustParse(&args)
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var members []struct {
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
	}
	if err := json.Unmarshal([]byte(args.Members), &members); err != nil {
		return fmt.Errorf("parse --members: %w", err)
	}
	fake := fakeMAX{}
	for _, m := range members {
		fake.members = append(fake.members, maxapi.Member{UserID: m.UserID, Name: m.Name})
	}

	cfg := config.Config{
		Mode:             config.ModePolling,
		HTTPAddr:         args.Addr,
		DBPath:           args.DB,
		LogLevel:         "error",
		LogFormat:        "console",
		PostTime:         config.TimeOfDay{Hour: 7},
		PublishUntil:     config.TimeOfDay{Hour: 8},
		CloseTime:        config.TimeOfDay{Hour: 18},
		CloseUntil:       config.TimeOfDay{Hour: 19},
		SummaryTime:      config.TimeOfDay{Hour: 20},
		SummaryUntil:     config.TimeOfDay{Hour: 21},
		JWTSecret:        "e2e-secret-e2e-secret-e2e-secret",
		BaseURL:          "http://localhost",
		MaxChatID:        -100,
		PaymentCard:      "2202 0000 0000 0000",
		PaymentRecipient: "Иванова Мария Петровна",
	}

	var s *store.Store
	fxApp := fx.New(
		app.Module,
		fx.Replace(cfg),
		fx.Replace(fx.Annotate(fake, fx.As(new(bot.MAX)))),
		fx.Replace(clock.Clock(func() time.Time { return args.Now.In(time.Local) })),
		fx.Invoke(func(*sql.DB, *http.Server) {}),
		fx.Populate(&s),
		fx.NopLogger,
	)
	ctx := context.Background()
	if err := fxApp.Start(ctx); err != nil {
		return err
	}
	defer fxApp.Stop(ctx)

	hash, err := auth.HashPassword(adminPassword)
	if err != nil {
		return err
	}
	if err := s.AddAdmin(ctx, adminLogin, hash); err != nil {
		return err
	}
	fmt.Println("ready")

	<-fxApp.Done()
	return nil
}
