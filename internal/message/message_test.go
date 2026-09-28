package message

import (
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/store"
)

func TestRender(t *testing.T) {
	date := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	buttons := []Button{{Text: "❌ Не обедает", Payload: "absent"}}
	intro := `Доброе утро!

Сегодня, <b>26 сентября (суббота)</b>, обед в школе.
По умолчанию обедают все — отвечать не нужно.

Кнопка ниже — только для тех, чей ребёнок сегодня НЕ обедает. Ждём до 08:20. Нажали по ошибке — нажмите ещё раз, отметка снимется.`

	tests := []struct {
		name     string
		closed   bool
		children []store.DayChild
		text     string
		buttons  []Button
	}{
		{
			name: "open, absent and rest",
			children: []store.DayChild{
				{LastName: "Сидоров", FirstName: "Коля", Status: store.Absent},
				{LastName: "Иванов", FirstName: "Петя", Status: store.Present},
				{LastName: "Козлова", FirstName: "Аня"},
			},
			text: intro + `

❌ Не обедают (1):
Сидоров Коля
✅ Обедают (2):
Иванов Петя
Козлова Аня

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`,
			buttons: buttons,
		},
		{
			name: "open, nobody absent",
			children: []store.DayChild{
				{LastName: "Козлова", FirstName: "Аня"},
			},
			text: intro + `

✅ Обедают все 1 человек:
Козлова Аня

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`,
			buttons: buttons,
		},
		{
			name: "open, all absent, sorted by last name then first name, ё as е",
			children: []store.DayChild{
				{LastName: "Жуков", FirstName: "Илья", Status: store.Absent},
				{LastName: "Ёлкин", FirstName: "Олег", Status: store.Absent},
				{LastName: "Семяшкин", FirstName: "Лёва", Status: store.Absent},
				{LastName: "Семёнов", FirstName: "Юра", Status: store.Absent},
				{LastName: "Семёнов", FirstName: "Антон", Status: store.Absent},
				{LastName: "Егоров", FirstName: "Дима", Status: store.Absent},
			},
			text: intro + `

❌ Не обедают (6):
Егоров Дима
Ёлкин Олег
Жуков Илья
Семёнов Антон
Семёнов Юра
Семяшкин Лёва

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`,
			buttons: buttons,
		},
		{
			name:   "closed, no buttons",
			closed: true,
			children: []store.DayChild{
				{LastName: "Иванов", FirstName: "Петя", Status: store.Present},
				{LastName: "Козлова", FirstName: "Аня", Status: store.Absent},
			},
			text: `Спасибо! Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.

❌ Не обедают (1):
Козлова Аня
✅ Обедают (1):
Иванов Петя

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`,
		},
		{
			name:   "closed, nobody absent",
			closed: true,
			children: []store.DayChild{
				{LastName: "Козлова", FirstName: "Аня"},
				{LastName: "Иванов", FirstName: "Петя", Status: store.Present},
			},
			text: `Спасибо! Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.

✅ Обедают все 2 человека:
Иванов Петя
Козлова Аня

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`,
		},
		{
			name:   "closed, no children",
			closed: true,
			text: `Спасибо! Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(Day{Date: date, CloseTime: "08:20", Closed: tt.closed, Children: tt.children, URL: "https://school.test/report/u1"})
			if got.Text != tt.text {
				t.Errorf("text:\n%s\n\nwant:\n%s", got.Text, tt.text)
			}
			if len(got.Buttons) != len(tt.buttons) {
				t.Fatalf("buttons = %v, want %v", got.Buttons, tt.buttons)
			}
			for i := range got.Buttons {
				if got.Buttons[i] != tt.buttons[i] {
					t.Errorf("button %d = %v, want %v", i, got.Buttons[i], tt.buttons[i])
				}
			}
		})
	}
}

func TestRenderClubs(t *testing.T) {
	date := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	got := Render(Day{
		Date:   date,
		Closed: true,
		Children: []store.DayChild{
			{ID: 1, LastName: "Иванов", FirstName: "Петя"},
			{ID: 5, LastName: "Петров", FirstName: "Гоша", Status: store.Absent},
		},
		Clubs: []store.DayClub{
			{ChildID: 2, LastName: "Козлова", FirstName: "Аня", Start: "15:00", Name: "Футбол"},
			{ChildID: 1, LastName: "Иванов", FirstName: "Петя", Start: "15:00", Name: "Футбол"},
			{ChildID: 5, LastName: "Петров", FirstName: "Гоша", Start: "15:00", Name: "Футбол"},
			{ChildID: 5, LastName: "Петров", FirstName: "Гоша", Start: "16:00", Name: "Бассейн"},
			{ChildID: 1, LastName: "Иванов", FirstName: "Петя", Start: "15:00", Name: "Дзюдо"},
			{ChildID: 3, LastName: "Ёлкин", FirstName: "Олег", Start: "14:30", Name: "Шахматы <A&B>"},
		},
		URL: "https://school.test/report/u1",
	}).Text

	want := `Спасибо! Сбор информации по обеду на <b>26 сентября (суббота)</b> завершён.

❌ Не обедают (1):
Петров Гоша
✅ Обедают (1):
Иванов Петя

🎨 Кружки и секции:
<b>14:30 Шахматы &lt;A&amp;B&gt;</b>: Ёлкин Олег
<b>15:00 Дзюдо</b>: Иванов Петя
<b>15:00 Футбол</b>: Иванов Петя, Козлова Аня

<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>`
	if got != want {
		t.Errorf("text:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestRenderWeekday(t *testing.T) {
	want := []string{"21 сентября (понедельник)", "22 сентября (вторник)", "23 сентября (среда)", "24 сентября (четверг)",
		"25 сентября (пятница)", "26 сентября (суббота)", "27 сентября (воскресенье)", "1 января (пятница)"}
	for i, w := range want {
		date := time.Date(2026, 9, 21+i, 0, 0, 0, 0, time.UTC)
		if i == 7 {
			date = time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		if text := Render(Day{Date: date, Closed: true}).Text; !strings.Contains(text, "по обеду на <b>"+w+"</b> завершён.") {
			t.Errorf("text:\n%s\nwant %q", text, w)
		}
	}
}
