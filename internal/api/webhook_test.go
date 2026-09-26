package api_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/apptest"
)

func callback(chatID, userID string) string {
	return `{
		"update_type": "message_callback",
		"timestamp": 1790395200000,
		"callback": {"callback_id": "cb1", "payload": "absent", "user": {"user_id": ` + userID + `}},
		"message": {"recipient": {"chat_id": ` + chatID + `}, "body": {"mid": "m1"}}
	}`
}

func webhookEnv(t *testing.T) *env {
	t.Helper()
	e := setup(t)
	e.a.SetNow(time.Date(2026, 9, 26, 7, 30, 0, 0, time.Local))
	e.a.Exec(t, `INSERT INTO children (id, last_name, first_name) VALUES (1, 'Иванов', 'Петя')`)
	e.a.Exec(t, `INSERT INTO child_parents (user_id, child_id) VALUES (101, 1)`)
	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-26', 'open', 'm1')`)
	return e
}

func status(t *testing.T, e *env) string {
	t.Helper()
	var s string
	if err := e.a.DB.QueryRow(`SELECT COALESCE((SELECT status FROM attendance WHERE child_id = 1), '')`).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestWebhook(t *testing.T) {
	t.Parallel()
	e := webhookEnv(t)

	w := e.do(t, "POST", "/webhook", callback("-100", "101"), nil, "X-Max-Bot-Api-Secret", "test-webhook-secret")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	if got := status(t, e); got != "absent" {
		t.Errorf("status = %q, want absent", got)
	}
	if got := e.a.MAX.Answered(); len(got) != 1 || got[0] != (apptest.Answer{CallbackID: "cb1", Notification: "Отмечено: не обедает"}) {
		t.Errorf("answered = %v, want cb1 with notification", got)
	}
	if edit := e.a.MAX.NextEdit(t); !strings.Contains(edit.Message.Text, "❌ Не обедают (1):\nИванов Петя") {
		t.Errorf("text:\n%s", edit.Message.Text)
	}
}

func TestWebhookRejected(t *testing.T) {
	t.Parallel()
	e := webhookEnv(t)

	tests := []struct {
		name, body, secret string
		want               int
	}{
		{"no secret", callback("-100", "101"), "", http.StatusUnauthorized},
		{"wrong secret", callback("-100", "101"), "wrong-secret", http.StatusUnauthorized},
		{"other chat", callback("-200", "101"), "test-webhook-secret", http.StatusOK},
	}
	for _, tt := range tests {
		w := e.do(t, "POST", "/webhook", tt.body, nil, "X-Max-Bot-Api-Secret", tt.secret)
		if w.Code != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.name, w.Code, tt.want)
		}
	}
	if got := status(t, e); got != "" {
		t.Errorf("status = %q, want no mark", got)
	}
	if got := e.a.MAX.Answered(); len(got) != 0 {
		t.Errorf("answered = %v, want none", got)
	}
}
