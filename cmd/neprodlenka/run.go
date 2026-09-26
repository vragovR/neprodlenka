package main

import (
	"database/sql"
	"net/http"

	"go.uber.org/fx"

	"neprodlenka/internal/app"
	"neprodlenka/internal/bot"
)

type RunCmd struct{}

func (c *RunCmd) Execute() error {
	fx.New(
		app.Module,
		fx.Invoke(func(*sql.DB, *http.Server) {}),
		fx.Invoke(bot.Start),
	).Run()
	return nil
}
