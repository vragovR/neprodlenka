package admincli_test

import (
	"bytes"
	"strings"
	"testing"

	"neprodlenka/internal/admincli"
	"neprodlenka/internal/apptest"
	"neprodlenka/internal/store"
)

func passwords(values ...string) admincli.PasswordReader {
	return func(string) (string, error) {
		v := values[0]
		values = values[1:]
		return v, nil
	}
}

func list(t *testing.T, s *store.Store) string {
	t.Helper()
	var out bytes.Buffer
	if err := admincli.List(t.Context(), s, &out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestAdminCommands(t *testing.T) {
	t.Parallel()
	var s *store.Store
	apptest.New(t, &s)
	ctx, out := t.Context(), &bytes.Buffer{}

	if err := admincli.Add(ctx, s, "anna", passwords("secret-pass", "secret-pass"), out); err != nil {
		t.Fatal(err)
	}
	if err := admincli.Add(ctx, s, "boris", passwords("other-pass", "other-pass"), out); err != nil {
		t.Fatal(err)
	}
	if err := admincli.Add(ctx, s, "anna", passwords("secret-pass", "secret-pass"), out); err == nil {
		t.Error("add existing admin succeeded")
	}
	if got := list(t, s); got != "anna\nboris\n" {
		t.Errorf("list = %q", got)
	}

	before, _, _ := s.AdminByLogin(ctx, "anna")
	if err := admincli.Passwd(ctx, s, "anna", passwords("new-password", "new-password"), out); err != nil {
		t.Fatal(err)
	}
	after, _, _ := s.AdminByLogin(ctx, "anna")
	if before.PasswordHash == after.PasswordHash {
		t.Error("password hash not changed")
	}

	if err := admincli.Remove(ctx, s, "boris", out); err != nil {
		t.Fatal(err)
	}
	if got := list(t, s); got != "anna\n" {
		t.Errorf("list after remove = %q", got)
	}
}

func TestAdminCommandErrors(t *testing.T) {
	t.Parallel()
	var s *store.Store
	apptest.New(t, &s)
	ctx, out := t.Context(), &bytes.Buffer{}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"short password", admincli.Add(ctx, s, "anna", passwords("short", "short"), out), "at least 8"},
		{"passwords differ", admincli.Add(ctx, s, "anna", passwords("secret-pass", "secret-pas"), out), "do not match"},
		{"passwd unknown", admincli.Passwd(ctx, s, "nobody", passwords("secret-pass", "secret-pass"), out), "not found"},
		{"remove unknown", admincli.Remove(ctx, s, "nobody", out), "not found"},
	}
	for _, tt := range tests {
		if tt.err == nil || !strings.Contains(tt.err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want %q", tt.name, tt.err, tt.want)
		}
	}
}
