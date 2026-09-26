package message

import (
	"testing"
	"time"

	"neprodlenka/internal/report"
)

func TestRenderMonth(t *testing.T) {
	month := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	r := report.Report{
		Children: []report.Row{
			{LastName: "Иванов", FirstName: "Петя", Present: 21, Total: 273000},
			{LastName: "Козлова", FirstName: "Аня", Present: 2, Total: 28000},
			{LastName: "Орлов", FirstName: "Гриша", Present: 11, Total: 143050},
			{LastName: "Петрова", FirstName: "Маша", Present: 0, Price: 13000},
		},
		Total: 444050,
	}

	got := RenderMonth(month, r, "https://school.test/report/u1", Payment{Card: "2202 0000 0000 0000", Recipient: "Иванова Мария Петровна"}).Text

	want := `Питание за сентябрь 2026.
Проверьте, пожалуйста, свою сумму и оплатите.

• <b>Иванов Петя</b>: 21 день × 130 ₽ = 2 730 ₽
• <b>Козлова Аня</b>: 2 дня × 140 ₽ = 280 ₽
• <b>Орлов Гриша</b>: 11 дней = 1 430,50 ₽
• <b>Петрова Маша</b>: 0 дней × 130 ₽ = 0 ₽

Итого по классу: 4 440,50 ₽ (34 обеда).
<a href="https://school.test/report/u1">Открыть отчёт за месяц</a>

Оплата на карту Сбербанк:
<b>2202 0000 0000 0000</b>
Получатель: Иванова Мария Петровна`
	if got != want {
		t.Errorf("text:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestPlural(t *testing.T) {
	want := map[int]string{0: "0 обедов", 1: "1 обед", 2: "2 обеда", 5: "5 обедов", 11: "11 обедов", 14: "14 обедов",
		21: "21 обед", 22: "22 обеда", 99: "99 обедов", 101: "101 обед", 111: "111 обедов"}
	for n, w := range want {
		if got := plural(n, "обед", "обеда", "обедов"); got != w {
			t.Errorf("plural(%d) = %q, want %q", n, got, w)
		}
	}
}
