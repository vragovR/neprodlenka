package api_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestReport(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-05'), ('2026-09-06')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-01', 1, 'present', 13000),
		('2026-09-02', 1, 'present', 12000),
		('2026-09-01', 2, 'absent', 0),
		('2026-09-02', 3, 'present', 13000),
		('2026-08-31', 4, 'present', 13000),
		('2026-10-01', 1, 'present', 13000)`)

	month := decode[struct {
		ReportURL string `json:"report_url"`
	}](t, e.do(t, "GET", "/api/months/2026-09", "", c).Body.String())
	path := "/api" + strings.TrimPrefix(month.ReportURL, "https://school.test")
	path = strings.Replace(path, "/report/", "/reports/", 1)

	w := e.do(t, "GET", path, "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	r := decode[struct {
		Month string `json:"month"`
		Days  []struct {
			School bool `json:"school"`
		} `json:"days"`
		Children []struct {
			LastName string    `json:"last_name"`
			Marks    []*string `json:"marks"`
			Present  int       `json:"present"`
			Absent   int       `json:"absent"`
			Total    int64     `json:"total"`
		} `json:"children"`
		Total int64 `json:"total"`
	}](t, w.Body.String())

	if r.Month != "2026-09" || len(r.Days) != 30 || r.Days[4].School || !r.Days[6].School {
		t.Errorf("month %s, days %+v", r.Month, r.Days)
	}
	var got []string
	for _, ch := range r.Children {
		got = append(got, ch.LastName+":"+itoa(int64(ch.Present))+"/"+itoa(int64(ch.Absent))+"/"+itoa(ch.Total))
	}
	if want := "Иванов:2/0/25000 Козлова:0/1/0 Сидоров:1/0/13000"; strings.Join(got, " ") != want {
		t.Errorf("children = %s, want %s", strings.Join(got, " "), want)
	}
	if r.Total != 38000 {
		t.Errorf("total = %d, want 38000", r.Total)
	}
}

func TestReportNotFound(t *testing.T) {
	t.Parallel()
	e := setup(t)

	if w := e.do(t, "GET", "/api/reports/unknown", "", nil); w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestChildReport(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-05')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-01', 1, 'present', 13000),
		('2026-09-02', 1, 'absent', 0),
		('2026-09-03', 1, 'present', 13000),
		('2026-09-01', 2, 'present', 15000)`)

	type childReport struct {
		Days []struct {
			School bool `json:"school"`
		} `json:"days"`
		Child struct {
			ID       int64     `json:"id"`
			LastName string    `json:"last_name"`
			Marks    []*string `json:"marks"`
			Present  int       `json:"present"`
			Absent   int       `json:"absent"`
			Total    int64     `json:"total"`
		} `json:"child"`
		Price   int64  `json:"price"`
		PrevUID string `json:"prev_uid"`
		NextUID string `json:"next_uid"`
	}

	e.a.Exec(t, `INSERT INTO report_months (month, uid) VALUES ('2026-08', 'rep-8'), ('2026-09', 'rep-9'), ('2026-10', 'rep-10')`)

	w := e.do(t, "GET", "/api/reports/rep-9/children/1", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	r := decode[childReport](t, w.Body.String())
	if len(r.Days) != 30 || r.Days[4].School || r.Child.ID != 1 || r.Child.LastName != "Иванов" ||
		r.Child.Present != 2 || r.Child.Absent != 1 || r.Child.Total != 26000 || r.Price != 13000 || r.PrevUID != "rep-8" || r.NextUID != "rep-10" {
		t.Errorf("report = %+v", r)
	}

	r = decode[childReport](t, e.do(t, "GET", "/api/reports/rep-10/children/4", "", nil).Body.String())
	if r.Child.LastName != "Орлов" || len(r.Child.Marks) != 31 || r.Child.Present != 0 || r.Price != 13000 || r.PrevUID != "rep-9" || r.NextUID != "" {
		t.Errorf("inactive child = %+v", r)
	}

	for path, code := range map[string]int{
		"/api/reports/rep-9/children/99":  http.StatusNotFound,
		"/api/reports/rep-9/children/abc": http.StatusNotFound,
		"/api/reports/unknown/children/1": http.StatusNotFound,
	} {
		if w := e.do(t, "GET", path, "", c); w.Code != code {
			t.Errorf("%s: status = %d, want %d", path, w.Code, code)
		}
	}
}
