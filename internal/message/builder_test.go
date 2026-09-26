package message_test

import (
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/apptest"
	"neprodlenka/internal/message"
)

var day = time.Date(2026, 9, 26, 0, 0, 0, 0, time.Local)

func seedChildren(t *testing.T, a *apptest.App) {
	t.Helper()
	a.Exec(t, `INSERT INTO children (id, last_name, first_name, active) VALUES
		(1, 'Иванов', 'Петя', 1),
		(2, 'Козлова', 'Аня', 1),
		(3, 'Сидоров', 'Коля', 0),
		(4, 'Орлов', 'Гриша', 0),
		(5, 'Петрова', 'Маша', 1)`)
	a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-26', 1, 'present', 13000),
		('2026-09-26', 3, 'absent', 0),
		('2026-09-25', 2, 'absent', 0),
		('2026-09-25', 4, 'present', 13000)`)
}

func TestBuildNoPost(t *testing.T) {
	t.Parallel()
	var b *message.Builder
	a := apptest.New(t, &b)
	seedChildren(t, a)

	_, _, ok, err := b.Build(t.Context(), day)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("ok = true, want false without daily post")
	}
}

func TestBuildOpen(t *testing.T) {
	t.Parallel()
	var b *message.Builder
	a := apptest.New(t, &b)
	seedChildren(t, a)
	a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-26', 'open', 'm1')`)

	msg, _, ok, err := b.Build(t.Context(), day)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ok = false, want true")
	}

	want := `Доброе утро!

Сегодня, <b>26 сентября (суббота)</b>, обед в школе.
По умолчанию обедают все — отвечать не нужно.

Кнопка ниже — только для тех, чей ребёнок сегодня НЕ обедает. Ждём до 09:00. Нажали по ошибке — нажмите ещё раз, отметка снимется.

❌ Не обедают (1):
Сидоров Коля
✅ Обедают (3):
Иванов Петя
Козлова Аня
Петрова Маша

<a href="https://school.test/report/`
	if !strings.HasPrefix(msg.Text, want) {
		t.Errorf("text:\n%s\n\nwant:\n%s", msg.Text, want)
	}
	if len(msg.Buttons) != 1 {
		t.Errorf("buttons = %v, want 1", msg.Buttons)
	}
}

func TestBuildClosed(t *testing.T) {
	t.Parallel()
	var b *message.Builder
	a := apptest.New(t, &b)
	seedChildren(t, a)
	a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-26', 'closed', 'm1')`)

	msg, _, ok, err := b.Build(t.Context(), day)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ok = false, want true")
	}

	want := `Спасибо! Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.

❌ Не обедают (1):
Сидоров Коля
✅ Обедают (3):
Иванов Петя
Козлова Аня
Петрова Маша

<a href="https://school.test/report/`
	if !strings.HasPrefix(msg.Text, want) {
		t.Errorf("text:\n%s\n\nwant:\n%s", msg.Text, want)
	}
	if len(msg.Buttons) != 0 {
		t.Errorf("buttons = %v, want none", msg.Buttons)
	}
}
