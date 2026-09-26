package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/apptest"
	"neprodlenka/internal/auth"
	"neprodlenka/internal/store"
)

type env struct {
	a     *apptest.App
	h     http.Handler
	store *store.Store
}

func setup(t *testing.T) *env {
	t.Helper()
	e := &env{}
	e.a = apptest.New(t, &e.h, &e.store)
	addAdmin(t, e, "anna", "secret-pass")
	return e
}

func addAdmin(t *testing.T, e *env, login, password string) {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.store.AddAdmin(t.Context(), login, hash); err != nil {
		t.Fatal(err)
	}
}

func (e *env) do(t *testing.T, method, path, body string, cookie *http.Cookie, headers ...string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		r.Header.Set(headers[i], headers[i+1])
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	e.h.ServeHTTP(w, r)
	return w
}

func (e *env) login(t *testing.T, login, password string) *httptest.ResponseRecorder {
	t.Helper()
	return e.do(t, "POST", "/api/login", `{"login":"`+login+`","password":"`+password+`"}`, nil)
}

func (e *env) session(t *testing.T) *http.Cookie {
	t.Helper()
	w := e.login(t, "anna", "secret-pass")
	if w.Code != http.StatusOK {
		t.Fatalf("login status = %d, body %s", w.Code, w.Body)
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			return c
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func (e *env) logout(t *testing.T, c *http.Cookie) int {
	t.Helper()
	return e.do(t, "DELETE", "/api/logout", "", c).Code
}

func TestLogin(t *testing.T) {
	t.Parallel()
	e := setup(t)

	w := e.login(t, "anna", "secret-pass")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	if got := strings.TrimSpace(w.Body.String()); got != `{"login":"anna"}` {
		t.Errorf("body = %s", got)
	}

	c := w.Result().Cookies()[0]
	if c.Name != "session" || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode || c.Path != "/api" {
		t.Errorf("cookie = %+v", c)
	}
	if c.MaxAge != int((30 * 24 * time.Hour).Seconds()) {
		t.Errorf("max age = %d", c.MaxAge)
	}

	if code := e.logout(t, c); code != http.StatusNoContent {
		t.Errorf("logout status = %d, want 204", code)
	}
}

func TestLoginFails(t *testing.T) {
	t.Parallel()
	e := setup(t)

	for _, tt := range []struct{ login, password string }{
		{"anna", "wrong-pass"},
		{"nobody", "secret-pass"},
	} {
		w := e.login(t, tt.login, tt.password)
		if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "invalid_credentials") {
			t.Errorf("login(%s, %s) = %d %s", tt.login, tt.password, w.Code, w.Body)
		}
		if len(w.Result().Cookies()) != 0 {
			t.Error("cookie set on failed login")
		}
	}
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()
	e := setup(t)
	c := e.session(t)

	tampered := *c
	tampered.Value = c.Value[:len(c.Value)-2] + "xx"

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "no cookie"},
		{name: "garbage", cookie: &http.Cookie{Name: "session", Value: "garbage"}},
		{name: "tampered", cookie: &tampered},
	}
	for _, tt := range tests {
		if code := e.logout(t, tt.cookie); code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", tt.name, code)
		}
	}
}

func TestSessionEnds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(t *testing.T, e *env)
	}{
		{name: "expired", change: func(t *testing.T, e *env) {
			e.a.SetNow(e.a.Now().Add(30*24*time.Hour + time.Second))
		}},
		{name: "password changed", change: func(t *testing.T, e *env) {
			hash, _ := auth.HashPassword("new-password")
			if _, err := e.store.SetAdminPassword(t.Context(), "anna", hash); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "admin removed", change: func(t *testing.T, e *env) {
			if _, err := e.store.RemoveAdmin(t.Context(), "anna"); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := setup(t)
			addAdmin(t, e, "boris", "boris-pass")
			c := e.session(t)

			tt.change(t, e)

			if code := e.logout(t, c); code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", code)
			}
		})
	}
}

func TestLoginRateLimit(t *testing.T) {
	t.Parallel()

	t.Run("by login", func(t *testing.T) {
		t.Parallel()
		e := setup(t)
		for i := range 5 {
			if w := e.login(t, "anna", "wrong-pass"); w.Code != http.StatusUnauthorized {
				t.Fatalf("attempt %d: status = %d", i, w.Code)
			}
		}
		if w := e.login(t, "anna", "secret-pass"); w.Code != http.StatusTooManyRequests {
			t.Errorf("status = %d, want 429", w.Code)
		}

		e.a.SetNow(e.a.Now().Add(15 * time.Minute))
		if w := e.login(t, "anna", "secret-pass"); w.Code != http.StatusOK {
			t.Errorf("after window: status = %d, want 200", w.Code)
		}
	})

	t.Run("by ip", func(t *testing.T) {
		t.Parallel()
		e := setup(t)
		for i := range 20 {
			w := e.do(t, "POST", "/api/login", `{"login":"user`+string(rune('a'+i))+`","password":"x"}`, nil, "X-Real-IP", "10.0.0.1")
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("attempt %d: status = %d", i, w.Code)
			}
		}
		w := e.do(t, "POST", "/api/login", `{"login":"anna","password":"secret-pass"}`, nil, "X-Real-IP", "10.0.0.1")
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("same ip: status = %d, want 429", w.Code)
		}
		w = e.do(t, "POST", "/api/login", `{"login":"anna","password":"secret-pass"}`, nil, "X-Real-IP", "10.0.0.2")
		if w.Code != http.StatusOK {
			t.Errorf("other ip: status = %d, want 200", w.Code)
		}
	})
}

func TestRequests(t *testing.T) {
	t.Parallel()
	e := setup(t)

	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		want        int
	}{
		{name: "login without json content type", method: "POST", path: "/api/login", body: `{"login":"anna","password":"secret-pass"}`, contentType: "text/plain", want: http.StatusUnsupportedMediaType},
		{name: "login with bad json", method: "POST", path: "/api/login", body: `{`, contentType: "application/json", want: http.StatusBadRequest},
		{name: "unknown path", method: "GET", path: "/api/nothing", want: http.StatusNotFound},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		if tt.contentType != "" {
			r.Header.Set("Content-Type", tt.contentType)
		}
		w := httptest.NewRecorder()
		e.h.ServeHTTP(w, r)
		if w.Code != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.name, w.Code, tt.want)
		}
	}
}
