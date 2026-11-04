package report

import (
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/store"
)

func row(r Row) string {
	var cells []string
	for _, m := range r.Marks {
		switch {
		case m == nil:
			cells = append(cells, "-")
		case *m == "present":
			cells = append(cells, "+")
		case *m == "pending":
			cells = append(cells, "?")
		default:
			cells = append(cells, "x")
		}
	}
	return r.LastName + " " + strings.Join(cells, "")
}

func TestBuild(t *testing.T) {
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	var noSchool []string
	for d := feb; d.Month() == time.February; d = d.AddDate(0, 0, 1) {
		if d.Day() > 3 {
			noSchool = append(noSchool, d.Format(time.DateOnly))
		}
	}

	children := []store.Child{
		{ID: 1, LastName: "Иванов", Active: true},
		{ID: 2, LastName: "Козлова", Active: true},
		{ID: 3, LastName: "Сидоров", Active: false},
		{ID: 4, LastName: "Орлов", Active: false},
		{ID: 5, LastName: "Петров", Active: true},
	}
	marks := []store.Mark{
		{ChildID: 1, Date: "2026-02-01", Status: store.Present, Price: 13000},
		{ChildID: 1, Date: "2026-02-02", Status: store.Present, Price: 14000},
		{ChildID: 1, Date: "2026-02-03", Status: store.Absent},
		{ChildID: 2, Date: "2026-02-02", Status: store.Absent},
		{ChildID: 2, Date: "2026-02-03", Status: store.Pending},
		{ChildID: 3, Date: "2026-02-01", Status: store.Present, Price: 13000},
		{ChildID: 1, Date: "2026-02-10", Status: store.Present, Price: 13000},
	}

	r := Build(feb, noSchool, children, marks)

	if r.Month != "2026-02" {
		t.Errorf("month = %s", r.Month)
	}
	var school []string
	for _, d := range r.Days {
		if d.School {
			school = append(school, d.Date)
		}
	}
	if len(r.Days) != 28 || r.Days[27].Date != "2026-02-28" || strings.Join(school, ",") != "2026-02-01,2026-02-02,2026-02-03" {
		t.Errorf("days = %+v", r.Days)
	}

	var rows []string
	for _, c := range r.Children {
		rows = append(rows, row(c))
	}
	if got, want := strings.Join(rows, "; "), "Иванов ++x------+------------------; Козлова -x?-------------------------; Сидоров +---------------------------"; got != want {
		t.Errorf("rows = %s, want %s", got, want)
	}

	totals := []struct {
		present, absent int
		total           int64
	}{{3, 1, 40000}, {0, 1, 0}, {1, 0, 13000}}
	for i, want := range totals {
		c := r.Children[i]
		if c.Present != want.present || c.Absent != want.absent || c.Total != want.total {
			t.Errorf("%s: present %d absent %d total %d, want %+v", c.LastName, c.Present, c.Absent, c.Total, want)
		}
	}
	if r.Total != 53000 {
		t.Errorf("class total = %d, want 53000", r.Total)
	}
}

func TestBuildEmpty(t *testing.T) {
	r := Build(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), nil, nil, nil)
	if len(r.Days) != 30 || len(r.Children) != 0 || r.Total != 0 {
		t.Errorf("report = %+v", r)
	}
}
