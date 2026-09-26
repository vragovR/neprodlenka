package message

import (
	_ "embed"
	"fmt"
	"html/template"
	"slices"
	"strings"
	"time"

	"neprodlenka/internal/names"
	"neprodlenka/internal/store"
)

type Button struct {
	Text    string
	Payload string
}

type Message struct {
	Text    string
	Buttons []Button
}

//go:embed day.tmpl
var dayText string

var dayTemplate = template.Must(template.New("day").Parse(dayText))

type dayData struct {
	Date      string
	CloseTime string
	Closed    bool
	Absent    []string
	URL       string
	Rest      int
}

type Day struct {
	Date      time.Time
	CloseTime string
	Closed    bool
	Children  []store.DayChild
	URL       string
}

func Render(d Day) Message {
	children := slices.Clone(d.Children)
	names.Sort(children, func(c store.DayChild) (string, string) { return c.LastName, c.FirstName })

	var absent []string
	for _, c := range children {
		if c.Status == store.Absent {
			absent = append(absent, c.LastName+" "+c.FirstName)
		}
	}

	var b strings.Builder
	err := dayTemplate.Execute(&b, dayData{
		Date:      fmt.Sprintf("%d %s (%s)", d.Date.Day(), monthsGenitive[d.Date.Month()-1], weekdays[d.Date.Weekday()]),
		CloseTime: d.CloseTime,
		Closed:    d.Closed,
		Absent:    absent,
		Rest:      len(children) - len(absent),
		URL:       d.URL,
	})
	if err != nil {
		panic(err)
	}
	msg := Message{Text: strings.TrimSuffix(b.String(), "\n")}
	if !d.Closed {
		msg.Buttons = []Button{{Text: "❌ Не обедает", Payload: string(store.Absent)}}
	}
	return msg
}

var weekdays = [...]string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}

var monthsGenitive = [...]string{"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря"}
