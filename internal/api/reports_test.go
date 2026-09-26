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
