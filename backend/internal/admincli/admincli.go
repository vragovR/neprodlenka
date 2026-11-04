package admincli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"neprodlenka/internal/auth"
	"neprodlenka/internal/store"
)

const minPasswordLen = 8

type PasswordReader func(prompt string) (string, error)

func Add(ctx context.Context, s *store.Store, login string, readPassword PasswordReader, out io.Writer) error {
	hash, err := newPassword(readPassword)
	if err != nil {
		return err
	}
	if err := s.AddAdmin(ctx, login, hash); errors.Is(err, store.ErrAdminExists) {
		return fmt.Errorf("admin %q already exists", login)
	} else if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "admin %q added\n", login)
	return err
}

func Passwd(ctx context.Context, s *store.Store, login string, readPassword PasswordReader, out io.Writer) error {
	hash, err := newPassword(readPassword)
	if err != nil {
		return err
	}
	ok, err := s.SetAdminPassword(ctx, login, hash)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("admin %q not found", login)
	}
	_, err = fmt.Fprintf(out, "password of %q changed\n", login)
	return err
}

func Remove(ctx context.Context, s *store.Store, login string, out io.Writer) error {
	ok, err := s.RemoveAdmin(ctx, login)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("admin %q not found", login)
	}
	_, err = fmt.Fprintf(out, "admin %q removed\n", login)
	return err
}

func List(ctx context.Context, s *store.Store, out io.Writer) error {
	logins, err := s.AdminLogins(ctx)
	if err != nil {
		return err
	}
	for _, login := range logins {
		if _, err := fmt.Fprintln(out, login); err != nil {
			return err
		}
	}
	return nil
}

func newPassword(readPassword PasswordReader) (string, error) {
	password, err := readPassword("Password: ")
	if err != nil {
		return "", err
	}
	if len([]rune(password)) < minPasswordLen {
		return "", fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}
	repeat, err := readPassword("Repeat password: ")
	if err != nil {
		return "", err
	}
	if password != repeat {
		return "", errors.New("passwords do not match")
	}
	return auth.HashPassword(password)
}
