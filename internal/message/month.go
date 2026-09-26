package message

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"neprodlenka/internal/report"
)

//go:embed month.tmpl
var monthText string

var monthTemplate = template.Must(template.New("month").Funcs(template.FuncMap{"money": money}).Parse(monthText))

var monthNames = [...]string{"январь", "февраль", "март", "апрель", "май", "июнь",
	"июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"}

type Payment struct {
	Card      string
	Recipient string
}

type monthRow struct {
	Name  string
	Days  string
	Price int64
	Total int64
}

type monthData struct {
	Month    string
	Children []monthRow
	Total    int64
	Meals    string
	URL      string
	Payment  Payment
}

func RenderMonth(month time.Time, r report.Report, url string, p Payment) Message {
	rows := make([]monthRow, 0, len(r.Children))
	meals := 0
	for _, c := range r.Children {
		price := c.Price
		if c.Present > 0 {
			price = 0
			if c.Total%int64(c.Present) == 0 {
				price = c.Total / int64(c.Present)
			}
		}
		rows = append(rows, monthRow{
			Name:  c.LastName + " " + c.FirstName,
			Days:  plural(c.Present, "день", "дня", "дней"),
			Price: price,
			Total: c.Total,
		})
		meals += c.Present
	}

	var b strings.Builder
	err := monthTemplate.Execute(&b, monthData{
		Month:    fmt.Sprintf("%s %d", monthNames[month.Month()-1], month.Year()),
		Children: rows,
		Total:    r.Total,
		Meals:    plural(meals, "обед", "обеда", "обедов"),
		URL:      url,
		Payment:  p,
	})
	if err != nil {
		panic(err)
	}
	return Message{Text: strings.TrimSuffix(b.String(), "\n")}
}

func plural(n int, one, few, many string) string {
	word := many
	switch {
	case n%100 >= 11 && n%100 <= 14:
	case n%10 == 1:
		word = one
	case n%10 >= 2 && n%10 <= 4:
		word = few
	}
	return fmt.Sprintf("%d %s", n, word)
}

func money(kopecks int64) string {
	rubles := fmt.Sprint(kopecks / 100)
	for i := len(rubles) - 3; i > 0; i -= 3 {
		rubles = rubles[:i] + " " + rubles[i:]
	}
	if k := kopecks % 100; k != 0 {
		rubles += fmt.Sprintf(",%02d", k)
	}
	return rubles + " ₽"
}
