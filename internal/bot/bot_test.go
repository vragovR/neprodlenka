package bot_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/apptest"
	"neprodlenka/internal/bot"
)

func at(day, hour, minute int) time.Time {
	return time.Date(2026, 9, day, hour, minute, 0, 0, time.Local)
}

func setup(t *testing.T) (*apptest.App, *bot.Bot) {
	t.Helper()
	var b *bot.Bot
	a := apptest.New(t, &b)
	a.Exec(t, `INSERT INTO children (id, last_name, first_name, price, active) VALUES
		(1, 'Иванов', 'Петя', 13000, 1),
		(2, 'Козлова', 'Аня', 15000, 1),
		(3, 'Сидоров', 'Коля', 13000, 0)`)
	a.Exec(t, `INSERT INTO child_parents (user_id, child_id) VALUES (101, 1), (102, 1), (201, 2), (301, 3)`)
	return a, b
}

func tick(t *testing.T, a *apptest.App, b *bot.Bot, now time.Time) {
	t.Helper()
	a.SetNow(now)
	if err := b.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func press(t *testing.T, b *bot.Bot, id string, userID int64, payload string) {
	t.Helper()
	err := b.HandleCallback(t.Context(), bot.Callback{ID: id, MessageID: "m1", UserID: userID, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
}

type mark struct {
	status string
	price  int
}

func marks(t *testing.T, a *apptest.App, date string) map[int]mark {
	t.Helper()
	rows, err := a.DB.Query(`SELECT child_id, status, price FROM attendance WHERE date = ?`, date)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	res := map[int]mark{}
	for rows.Next() {
		var (
			id int
			m  mark
		)
		if err := rows.Scan(&id, &m.status, &m.price); err != nil {
			t.Fatal(err)
		}
		res[id] = m
	}
	return res
}

type post struct {
	status    string
	messageID string
}

func getPost(t *testing.T, a *apptest.App, date string) (post, bool) {
	t.Helper()
	var (
		p  post
		id sql.NullString
	)
	err := a.DB.QueryRow(`SELECT status, message_id FROM daily_posts WHERE date = ?`, date).Scan(&p.status, &id)
	if err == sql.ErrNoRows {
		return post{}, false
	}
	if err != nil {
		t.Fatal(err)
	}
	p.messageID = id.String
	return p, true
}

func equal(a, b map[int]mark) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestTickPublish(t *testing.T) {
	t.Parallel()
	a, b := setup(t)

	tick(t, a, b, at(26, 6, 59))
	if got := len(a.MAX.Sent()); got != 0 {
		t.Fatalf("sent %d before POST_TIME, want 0", got)
	}

	tick(t, a, b, at(26, 7, 0))
	tick(t, a, b, at(26, 8, 30))

	sent := a.MAX.Sent()
	if len(sent) != 1 {
		t.Fatalf("sent %d, want 1", len(sent))
	}
	if !strings.Contains(sent[0].Text, "✅ Обедают все 2 человека:") {
		t.Errorf("text:\n%s", sent[0].Text)
	}
	if len(sent[0].Buttons) != 1 {
		t.Errorf("buttons = %v, want 1", sent[0].Buttons)
	}
	if p, _ := getPost(t, a, "2026-09-26"); p != (post{"open", "m1"}) {
		t.Errorf("post = %v, want open m1", p)
	}
	var months int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM report_months WHERE month = '2026-09'`).Scan(&months); err != nil || months != 1 {
		t.Errorf("report months = %d (%v), want 1", months, err)
	}
}

func TestTickNoPublish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		now      time.Time
		noSchool bool
	}{
		{name: "no school day", now: at(26, 7, 30), noSchool: true},
		{name: "after publish deadline", now: at(26, 8, 0)},
		{name: "after close time", now: at(26, 9, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := setup(t)
			if tt.noSchool {
				a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-26')`)
			}

			tick(t, a, b, tt.now)

			if got := len(a.MAX.Sent()); got != 0 {
				t.Errorf("sent %d, want 0", got)
			}
			if _, ok := getPost(t, a, "2026-09-26"); ok {
				t.Error("post created, want none")
			}
		})
	}
}

func TestTickClose(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	tick(t, a, b, at(26, 7, 0))
	press(t, b, "c1", 101, "absent")
	a.MAX.NextEdit(t)

	tick(t, a, b, at(26, 8, 59))
	if p, _ := getPost(t, a, "2026-09-26"); p.status != "open" {
		t.Fatalf("status = %q before CLOSE_TIME, want open", p.status)
	}

	tick(t, a, b, at(26, 9, 0))

	if p, _ := getPost(t, a, "2026-09-26"); p.status != "closed" {
		t.Errorf("status = %q, want closed", p.status)
	}
	if got, want := marks(t, a, "2026-09-26"), map[int]mark{1: {"absent", 0}, 2: {"present", 15000}}; !equal(got, want) {
		t.Errorf("marks = %v, want %v", got, want)
	}

	edit := a.MAX.NextEdit(t)
	if !strings.HasPrefix(edit.Message.Text, "Спасибо! Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.") {
		t.Errorf("text:\n%s", edit.Message.Text)
	}
	if len(edit.Message.Buttons) != 0 {
		t.Errorf("buttons = %v, want none", edit.Message.Buttons)
	}
}

func TestTickClosesOnlyTodayInWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		date string
		now  time.Time
		want string
	}{
		{name: "today in window", date: "2026-09-26", now: at(26, 9, 30), want: "closed"},
		{name: "today after window", date: "2026-09-26", now: at(26, 10, 0), want: "open"},
		{name: "past day", date: "2026-09-25", now: at(26, 9, 30), want: "open"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := setup(t)
			a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES (?, 'open', 'm1')`, tt.date)

			tick(t, a, b, tt.now)

			if p, _ := getPost(t, a, tt.date); p.status != tt.want {
				t.Errorf("status = %q, want %q", p.status, tt.want)
			}
			if tt.want == "open" {
				a.MAX.NoEdit(t)
				return
			}
			if got := marks(t, a, tt.date); len(got) != 2 {
				t.Errorf("marks = %v, want 2 present", got)
			}
			if edit := a.MAX.NextEdit(t); !strings.Contains(edit.Message.Text, "Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.") {
				t.Errorf("text:\n%s", edit.Message.Text)
			}
		})
	}
}

func TestTickSendError(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.MAX.SetSendError(errors.New("max unavailable"))

	a.SetNow(at(26, 7, 0))
	if err := b.Tick(t.Context()); err == nil {
		t.Fatal("tick succeeded, want send error")
	}
	if _, ok := getPost(t, a, "2026-09-26"); ok {
		t.Fatal("post left after send error, want deleted")
	}

	a.MAX.SetSendError(nil)
	tick(t, a, b, at(26, 7, 1))

	if got := len(a.MAX.Sent()); got != 1 {
		t.Errorf("sent %d, want 1", got)
	}
	if p, _ := getPost(t, a, "2026-09-26"); p != (post{"open", "m1"}) {
		t.Errorf("post = %v, want open m1", p)
	}
}

func TestTickPostWithoutMessageID(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO daily_posts (date) VALUES ('2026-09-26')`)

	tick(t, a, b, at(26, 7, 30))
	tick(t, a, b, at(26, 9, 30))

	if got := len(a.MAX.Sent()); got != 0 {
		t.Errorf("sent %d, want 0", got)
	}
	if p, _ := getPost(t, a, "2026-09-26"); p != (post{"open", ""}) {
		t.Errorf("post = %v, want untouched", p)
	}
	if got := marks(t, a, "2026-09-26"); len(got) != 0 {
		t.Errorf("marks = %v, want none", got)
	}
}

func TestPress(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	tick(t, a, b, at(26, 7, 0))

	a.SetNow(at(26, 7, 30))
	press(t, b, "c1", 101, "absent")
	press(t, b, "c2", 201, "absent")

	if got, want := marks(t, a, "2026-09-26"), map[int]mark{1: {"absent", 0}, 2: {"absent", 0}}; !equal(got, want) {
		t.Errorf("marks = %v, want %v", got, want)
	}
	if got := a.MAX.Answered(); len(got) != 2 || got[0].Notification != "Отмечено: не обедает" {
		t.Errorf("answered = %v, want 2 callbacks", got)
	}

	edit := a.MAX.NextEdit(t)
	if edit.MessageID != "m1" {
		t.Errorf("edited %q, want m1", edit.MessageID)
	}
	if !strings.Contains(edit.Message.Text, "❌ Не обедают (") || !strings.Contains(edit.Message.Text, "\nИванов Петя") {
		t.Errorf("text:\n%s", edit.Message.Text)
	}
}

func TestPressAgainUnmarks(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	tick(t, a, b, at(26, 7, 0))

	press(t, b, "c1", 101, "absent")
	press(t, b, "c2", 102, "absent")

	if got := marks(t, a, "2026-09-26"); len(got) != 0 {
		t.Errorf("marks = %v, want none", got)
	}
	if got := a.MAX.Answered(); len(got) != 2 || got[1].Notification != "Отметка снята: обедает" {
		t.Errorf("answered = %v", got)
	}
}

func TestPressMarksAllChildrenOfParent(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO child_parents (user_id, child_id) VALUES (101, 2)`)
	tick(t, a, b, at(26, 7, 0))

	press(t, b, "c1", 101, "absent")

	if got, want := marks(t, a, "2026-09-26"), map[int]mark{1: {"absent", 0}, 2: {"absent", 0}}; !equal(got, want) {
		t.Errorf("marks = %v, want %v", got, want)
	}

	press(t, b, "c2", 201, "absent")
	press(t, b, "c3", 101, "absent")

	if got, want := marks(t, a, "2026-09-26"), map[int]mark{1: {"absent", 0}, 2: {"absent", 0}}; !equal(got, want) {
		t.Errorf("after partial unmark: marks = %v, want %v", got, want)
	}

	press(t, b, "c4", 101, "absent")

	if got := marks(t, a, "2026-09-26"); len(got) != 0 {
		t.Errorf("after unmark: marks = %v, want none", got)
	}
}

func TestPressOpenPostAnyTime(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-20', 'open', 'm1')`)
	a.SetNow(at(26, 23, 0))

	press(t, b, "c1", 101, "absent")

	if got, want := marks(t, a, "2026-09-20"), map[int]mark{1: {"absent", 0}}; !equal(got, want) {
		t.Errorf("marks = %v, want %v", got, want)
	}
}

func TestPressIgnored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userID  int64
		payload string
		now     time.Time
		closed  bool
	}{
		{name: "unbound parent", userID: 999, payload: "absent"},
		{name: "inactive child", userID: 301, payload: "absent"},
		{name: "unknown payload", userID: 101, payload: "maybe"},
		{name: "present payload", userID: 101, payload: "present"},
		{name: "closed post", userID: 101, payload: "absent", closed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := setup(t)
			tick(t, a, b, at(26, 7, 0))
			if tt.closed {
				a.Exec(t, `UPDATE daily_posts SET status = 'closed'`)
			}
			if !tt.now.IsZero() {
				a.SetNow(tt.now)
			}

			press(t, b, "c1", tt.userID, tt.payload)

			if got := marks(t, a, "2026-09-26"); len(got) != 0 {
				t.Errorf("marks = %v, want none", got)
			}
			if got := a.MAX.Answered(); len(got) != 1 || got[0].Notification == "" {
				t.Errorf("answered = %v, want 1 callback", got)
			}
		})
	}
}

func TestTickSummary(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-29', 1, 'present', 13000),
		('2026-09-29', 2, 'absent', 0)`)
	a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-30', 'closed', 'm0')`)

	tick(t, a, b, at(29, 12, 0))
	tick(t, a, b, at(30, 11, 59))
	if got := len(a.MAX.Sent()); got != 0 {
		t.Fatalf("sent %d before SUMMARY_TIME on last school day, want 0", got)
	}

	tick(t, a, b, at(30, 12, 0))
	tick(t, a, b, at(30, 12, 1))

	sent := a.MAX.Sent()
	if len(sent) != 1 {
		t.Fatalf("sent %d, want 1", len(sent))
	}
	if want := "Питание за сентябрь 2026.\nПроверьте, пожалуйста, свою сумму и оплатите.\n\n• <b>Иванов Петя</b>: 1 день × 130\u00a0₽ = 130\u00a0₽\n• <b>Козлова Аня</b>: 0 дней × 150\u00a0₽ = 0\u00a0₽\n\nИтого по классу: 130\u00a0₽ (1 обед)."; !strings.HasPrefix(sent[0].Text, want) {
		t.Errorf("text:\n%s", sent[0].Text)
	}
	var messageID string
	if err := a.DB.QueryRow(`SELECT message_id FROM report_months WHERE month = '2026-09'`).Scan(&messageID); err != nil || messageID != "m1" {
		t.Errorf("message id = %q (%v), want m1", messageID, err)
	}
}

func TestTickSummaryLastSchoolDay(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES ('2026-09-28', 1, 'present', 13000)`)
	a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-27'), ('2026-09-30')`)

	tick(t, a, b, at(28, 12, 0))
	if got := len(a.MAX.Sent()); got != 0 {
		t.Fatalf("sent %d before last school day, want 0", got)
	}

	tick(t, a, b, at(29, 12, 0))
	if got := len(a.MAX.Sent()); got != 1 {
		t.Errorf("sent %d on last school day, want 1", got)
	}
}

func TestTickSummaryWindow(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES ('2026-09-29', 1, 'present', 13000)`)
	a.MAX.SetSendError(errors.New("max unavailable"))

	a.SetNow(at(30, 12, 59))
	if err := b.Tick(t.Context()); err == nil {
		t.Fatal("tick succeeded, want send error")
	}

	a.MAX.SetSendError(nil)
	tick(t, a, b, at(30, 13, 0))

	if got := len(a.MAX.Sent()); got != 0 {
		t.Errorf("sent %d after SUMMARY_UNTIL, want 0", got)
	}
}

func TestTickSummaryNoMarks(t *testing.T) {
	t.Parallel()
	a, b := setup(t)
	a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-30')`)

	tick(t, a, b, at(30, 12, 0))

	if got := len(a.MAX.Sent()); got != 0 {
		t.Errorf("sent %d, want 0", got)
	}
}
