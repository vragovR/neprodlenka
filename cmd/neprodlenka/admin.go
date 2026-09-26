package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/fx"
	"golang.org/x/term"

	"neprodlenka/internal/admincli"
	"neprodlenka/internal/app"
	"neprodlenka/internal/store"
)

type AdminCmd struct {
	Add    *AdminAddCmd    `arg:"subcommand:add" help:"добавить администратора"`
	Passwd *AdminPasswdCmd `arg:"subcommand:passwd" help:"сменить пароль"`
	Remove *AdminRemoveCmd `arg:"subcommand:remove" help:"удалить администратора"`
	List   *AdminListCmd   `arg:"subcommand:list" help:"список логинов"`
}

type AdminAddCmd struct {
	Login string `arg:"positional,required"`
}

func (c *AdminAddCmd) Execute() error {
	return withStore(func(ctx context.Context, s *store.Store) error {
		return admincli.Add(ctx, s, c.Login, readPassword(), os.Stdout)
	})
}

type AdminPasswdCmd struct {
	Login string `arg:"positional,required"`
}

func (c *AdminPasswdCmd) Execute() error {
	return withStore(func(ctx context.Context, s *store.Store) error {
		return admincli.Passwd(ctx, s, c.Login, readPassword(), os.Stdout)
	})
}

type AdminRemoveCmd struct {
	Login string `arg:"positional,required"`
}

func (c *AdminRemoveCmd) Execute() error {
	return withStore(func(ctx context.Context, s *store.Store) error {
		return admincli.Remove(ctx, s, c.Login, os.Stdout)
	})
}

type AdminListCmd struct{}

func (c *AdminListCmd) Execute() error {
	return withStore(func(ctx context.Context, s *store.Store) error {
		return admincli.List(ctx, s, os.Stdout)
	})
}

func withStore(f func(context.Context, *store.Store) error) error {
	var s *store.Store
	fxApp := fx.New(app.Module, fx.Populate(&s), fx.NopLogger)
	ctx := context.Background()
	if err := fxApp.Start(ctx); err != nil {
		return err
	}
	defer fxApp.Stop(ctx)

	return f(ctx, s)
}

func readPassword() admincli.PasswordReader {
	stdin := bufio.NewReader(os.Stdin)
	return func(prompt string) (string, error) {
		fmt.Fprint(os.Stderr, prompt)
		fd := int(os.Stdin.Fd())
		if term.IsTerminal(fd) {
			b, err := term.ReadPassword(fd)
			fmt.Fprintln(os.Stderr)
			return string(b), err
		}
		line, err := stdin.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
}
