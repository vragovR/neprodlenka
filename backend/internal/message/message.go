package message

import (
	"cmp"
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
	Present   []string
	People    string
	Clubs     []clubData
	URL       string
}

type clubData struct {
	Start    string
	Name     string
	Children string
}

type Day struct {
	Date      time.Time
	CloseTime string
	Closed    bool
	Children  []store.DayChild
	Clubs     []store.DayClub
	URL       string
}

func Render(d Day) Message {
	children := slices.Clone(d.Children)
	names.Sort(children, func(c store.DayChild) (string, string) { return c.LastName, c.FirstName })

	var absent, present []string
	absentIDs := map[int64]bool{}
	for _, c := range children {
		if c.Status == store.Absent {
			absentIDs[c.ID] = true
			absent = append(absent, c.LastName+" "+c.FirstName)
		} else {
			present = append(present, c.LastName+" "+c.FirstName)
		}
	}

	var b strings.Builder
	err := dayTemplate.Execute(&b, dayData{
		Date:      fmt.Sprintf("%d %s (%s)", d.Date.Day(), monthsGenitive[d.Date.Month()-1], weekdays[d.Date.Weekday()]),
		CloseTime: d.CloseTime,
		Closed:    d.Closed,
		Absent:    absent,
		Present:   present,
		People:    plural(len(present), "человек", "человека", "человек"),
		Clubs:     groupClubs(slices.DeleteFunc(slices.Clone(d.Clubs), func(c store.DayClub) bool { return absentIDs[c.ChildID] })),
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

func groupClubs(clubs []store.DayClub) []clubData {
	names.Sort(clubs, func(c store.DayClub) (string, string) { return c.LastName, c.FirstName })
	slices.SortStableFunc(clubs, func(a, b store.DayClub) int {
		return cmp.Or(strings.Compare(a.Start, b.Start), strings.Compare(a.Name, b.Name))
	})

	var groups []clubData
	var children []string
	for i, c := range clubs {
		children = append(children, c.LastName+" "+c.FirstName)
		if i+1 < len(clubs) && clubs[i+1].Start == c.Start && clubs[i+1].Name == c.Name {
			continue
		}
		groups = append(groups, clubData{Start: c.Start, Name: c.Name, Children: strings.Join(children, ", ")})
		children = nil
	}
	return groups
}

var weekdays = [...]string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}

var monthsGenitive = [...]string{"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря"}
