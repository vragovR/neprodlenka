package report

import (
	"slices"
	"time"

	"neprodlenka/internal/store"
)

type Report struct {
	Month    string `json:"month"`
	Days     []Day  `json:"days"`
	Children []Row  `json:"children"`
	Total    int64  `json:"total"`
}

type Day struct {
	Date   string `json:"date"`
	School bool   `json:"school"`
}

type Row struct {
	LastName   string    `json:"last_name"`
	FirstName  string    `json:"first_name"`
	MiddleName string    `json:"middle_name"`
	Marks      []*string `json:"marks"`
	Present    int       `json:"present"`
	Absent     int       `json:"absent"`
	Total      int64     `json:"total"`
	Price      int64     `json:"-"`
}

func Build(month time.Time, noSchoolDays []string, children []store.Child, marks []store.Mark) Report {
	byChild := map[int64]map[string]store.Mark{}
	for _, m := range marks {
		if byChild[m.ChildID] == nil {
			byChild[m.ChildID] = map[string]store.Mark{}
		}
		byChild[m.ChildID][m.Date] = m
	}

	r := Report{Month: month.Format("2006-01"), Days: []Day{}, Children: []Row{}}
	for d := month; d.Month() == month.Month(); d = d.AddDate(0, 0, 1) {
		date := d.Format(time.DateOnly)
		r.Days = append(r.Days, Day{Date: date, School: !slices.Contains(noSchoolDays, date)})
	}

	for _, c := range children {
		childMarks := byChild[c.ID]
		if !c.Active && len(childMarks) == 0 {
			continue
		}
		row := Row{
			LastName:   c.LastName,
			FirstName:  c.FirstName,
			MiddleName: c.MiddleName,
			Price:      c.Price,
			Marks:      make([]*string, len(r.Days)),
		}
		for i, day := range r.Days {
			m, ok := childMarks[day.Date]
			if !ok {
				continue
			}
			status := string(m.Status)
			row.Marks[i] = &status
			if m.Status == store.Present {
				row.Present++
			} else {
				row.Absent++
			}
			row.Total += m.Price
		}
		r.Total += row.Total
		r.Children = append(r.Children, row)
	}
	return r
}
