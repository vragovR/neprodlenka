package api_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"neprodlenka/internal/maxapi"
)

func adminEnv(t *testing.T) (*env, *http.Cookie) {
	t.Helper()
	e := setup(t)
	e.a.SetNow(time.Date(2026, 9, 26, 10, 0, 0, 0, time.Local))
	return e, e.session(t)
}

func decode[T any](t *testing.T, body string) T {
	t.Helper()
	var v T
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return v
}

type child struct {
	ID         int64   `json:"id"`
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName string  `json:"middle_name"`
	Price      int64   `json:"price"`
	Active     bool    `json:"active"`
	Weekdays   []int   `json:"weekdays"`
	ParentIDs  []int64 `json:"parent_ids"`
	Clubs      []club  `json:"clubs"`
}

type club struct {
	Weekday int    `json:"weekday"`
	Start   string `json:"start"`
	Name    string `json:"name"`
}

func TestAdminRequiresLogin(t *testing.T) {
	t.Parallel()
	e := setup(t)

	for _, req := range []struct{ method, path, body string }{
		{"GET", "/api/children", ""},
		{"POST", "/api/children", `{}`},
		{"PATCH", "/api/children/1", `{}`},
		{"GET", "/api/months/2026-09", ""},
		{"PUT", "/api/months/2026-09", `{}`},
		{"GET", "/api/attendance/2026-09", ""},
		{"PUT", "/api/attendance/2026-09", `{}`},
		{"GET", "/api/days/2026-09-25", ""},
		{"PUT", "/api/days/2026-09-25/children/1/clubs", `{}`},
		{"POST", "/api/debug/day/2026-09-25", `{}`},
		{"POST", "/api/debug/month/2026-09", `{}`},
		{"POST", "/api/debug/close/2026-09-25", `{}`},
		{"DELETE", "/api/debug/day/2026-09-25", ""},
		{"DELETE", "/api/debug/month/2026-09", ""},
		{"GET", "/api/debug/summaries", ""},
	} {
		if w := e.do(t, req.method, req.path, req.body, nil); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want 401", req.method, req.path, w.Code)
		}
	}
}

func TestChildren(t *testing.T) {
	t.Parallel()
	e, c := adminEnv(t)
	e.a.MAX.SetMembers(maxapi.Member{UserID: 101, Name: "Мама Пети"}, maxapi.Member{UserID: 102, Name: "Папа Пети"})

	w := e.do(t, "POST", "/api/children", `{"last_name":" Семёнов ","first_name":"Петя"}`, c)
	if w.Code != http.StatusCreated {
		t.Fatalf("add: status = %d, body %s", w.Code, w.Body)
	}
	petya := decode[child](t, w.Body.String())
	if petya.LastName != "Семёнов" || petya.Price != 13000 || !petya.Active || !slices.Equal(petya.Weekdays, []int{1, 2, 3, 4, 5}) ||
		len(petya.ParentIDs) != 0 || petya.Clubs == nil || len(petya.Clubs) != 0 {
		t.Errorf("added = %+v", petya)
	}

	w = e.do(t, "POST", "/api/children", `{"last_name":"Ёлкина","first_name":"Аня","middle_name":"Олеговна","price":15000}`, c)
	if w.Code != http.StatusCreated {
		t.Fatalf("add: status = %d, body %s", w.Code, w.Body)
	}
	anya := decode[child](t, w.Body.String())

	w = e.do(t, "PATCH", "/api/children/"+itoa(anya.ID), `{"parent_ids":[]}`, c)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"parent_ids":[]`) {
		t.Errorf("unlink all: status = %d, body %s", w.Code, w.Body)
	}

	w = e.do(t, "PATCH", "/api/children/"+itoa(petya.ID), `{"price":14000,"parent_ids":[102,101,101]}`, c)
	if w.Code != http.StatusOK {
		t.Fatalf("patch: status = %d, body %s", w.Code, w.Body)
	}
	if got := decode[child](t, w.Body.String()); got.Price != 14000 || got.LastName != "Семёнов" || !equalIDs(got.ParentIDs, 101, 102) {
		t.Errorf("patched = %+v", got)
	}

	w = e.do(t, "PATCH", "/api/children/"+itoa(petya.ID),
		`{"clubs":[{"weekday":3,"start":"15:00","name":" Шахматы "},{"weekday":1,"start":"14:30","name":"Робототехника"},{"weekday":3,"start":"15:00","name":"Шахматы"}]}`, c)
	if w.Code != http.StatusOK {
		t.Fatalf("patch clubs: status = %d, body %s", w.Code, w.Body)
	}
	wantClubs := []club{{1, "14:30", "Робототехника"}, {3, "15:00", "Шахматы"}}
	if got := decode[child](t, w.Body.String()); !slices.Equal(got.Clubs, wantClubs) || !equalIDs(got.ParentIDs, 101, 102) {
		t.Errorf("patched clubs = %+v", got)
	}

	w = e.do(t, "PATCH", "/api/children/"+itoa(petya.ID), `{"weekdays":[4,2,4]}`, c)
	if got := decode[child](t, w.Body.String()); w.Code != http.StatusOK || !slices.Equal(got.Weekdays, []int{2, 4}) || got.Price != 14000 {
		t.Errorf("patch weekdays: status = %d, child %+v", w.Code, got)
	}

	w = e.do(t, "PATCH", "/api/children/"+itoa(petya.ID), `{"middle_name":" Олегович ","active":false}`, c)
	if w.Code != http.StatusOK {
		t.Fatalf("deactivate: status = %d, body %s", w.Code, w.Body)
	}
	if got := decode[child](t, w.Body.String()); got.MiddleName != "Олегович" || got.Active || got.Price != 14000 || !slices.Equal(got.Weekdays, []int{2, 4}) ||
		!equalIDs(got.ParentIDs, 101, 102) || !slices.Equal(got.Clubs, wantClubs) {
		t.Errorf("deactivated = %+v", got)
	}
	w = e.do(t, "PATCH", "/api/children/"+itoa(petya.ID), `{"active":true}`, c)
	if got := decode[child](t, w.Body.String()); w.Code != http.StatusOK || !got.Active || !slices.Equal(got.Clubs, wantClubs) {
		t.Errorf("activate: status = %d, child %+v", w.Code, got)
	}

	w = e.do(t, "GET", "/api/children", "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("list: status = %d, body %s", w.Code, w.Body)
	}
	list := decode[struct {
		Children []child `json:"children"`
		Members  []struct {
			UserID int64  `json:"user_id"`
			Name   string `json:"name"`
		} `json:"members"`
	}](t, w.Body.String())
	if len(list.Children) != 2 || list.Children[0].LastName != "Ёлкина" || list.Children[1].LastName != "Семёнов" {
		t.Errorf("children = %+v", list.Children)
	}
	if !equalIDs(list.Children[1].ParentIDs, 101, 102) || !slices.Equal(list.Children[1].Clubs, wantClubs) || len(list.Children[0].Clubs) != 0 || len(list.Members) != 2 || list.Members[0].Name != "Мама Пети" {
		t.Errorf("list = %+v", list)
	}
}

func TestChildrenMAXUnavailable(t *testing.T) {
	t.Parallel()
	e, c := adminEnv(t)
	e.a.MAX.SetError(errors.New("max unavailable"))

	if w := e.do(t, "GET", "/api/children", "", c); w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502, body %s", w.Code, w.Body)
	}
}

func TestChildErrors(t *testing.T) {
	t.Parallel()
	e, c := adminEnv(t)
	e.a.Exec(t, `INSERT INTO children (id, last_name, first_name) VALUES (1, 'Иванов', 'Петя'), (2, 'Козлова', 'Аня')`)

	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"no last name", "POST", "/api/children", `{"first_name":"Петя"}`, http.StatusBadRequest},
		{"blank first name", "POST", "/api/children", `{"last_name":"Иванов","first_name":"  "}`, http.StatusBadRequest},
		{"negative price", "POST", "/api/children", `{"last_name":"Иванов","first_name":"Петя","price":-1}`, http.StatusBadRequest},
		{"unknown field", "POST", "/api/children", `{"last_name":"Иванов","first_name":"Петя","age":7}`, http.StatusBadRequest},
		{"clear last name", "PATCH", "/api/children/2", `{"last_name":""}`, http.StatusBadRequest},
		{"blank club name", "PATCH", "/api/children/2", `{"clubs":[{"weekday":1,"start":"14:00","name":" "}]}`, http.StatusBadRequest},
		{"long club name", "PATCH", "/api/children/2", `{"clubs":[{"weekday":1,"start":"14:00","name":"` + strings.Repeat("я", 101) + `"}]}`, http.StatusBadRequest},
		{"bad club weekday", "PATCH", "/api/children/2", `{"clubs":[{"weekday":8,"start":"14:00","name":"Шахматы"}]}`, http.StatusBadRequest},
		{"bad club start", "PATCH", "/api/children/2", `{"clubs":[{"weekday":1,"start":"25:00","name":"Шахматы"}]}`, http.StatusBadRequest},
		{"no weekdays", "PATCH", "/api/children/2", `{"weekdays":[]}`, http.StatusBadRequest},
		{"bad weekday", "PATCH", "/api/children/2", `{"weekdays":[2,8]}`, http.StatusBadRequest},
		{"unknown child", "PATCH", "/api/children/99", `{"price":1}`, http.StatusNotFound},
		{"bad id", "PATCH", "/api/children/abc", `{"price":1}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		if w := e.do(t, tt.method, tt.path, tt.body, c); w.Code != tt.want {
			t.Errorf("%s: status = %d, want %d, body %s", tt.name, w.Code, tt.want, w.Body)
		}
	}
}

func TestChildRedrawsTodayOpenPost(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"open", "closed"} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			e, c := adminEnv(t)
			e.a.Exec(t, `INSERT INTO children (id, last_name, first_name) VALUES (1, 'Иванов', 'Петя')`)
			e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-26', ?, 'm1')`, status)
			e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES ('2026-09-26', 1, 'absent', 0)`)

			if w := e.do(t, "PATCH", "/api/children/1", `{"first_name":"Пётр"}`, c); w.Code != http.StatusOK {
				t.Fatalf("status = %d, body %s", w.Code, w.Body)
			}

			if status == "closed" {
				e.a.MAX.NoEdit(t)
				return
			}
			if edit := e.a.MAX.NextEdit(t); !strings.Contains(edit.Message.Text, "Иванов Пётр") {
				t.Errorf("text:\n%s", edit.Message.Text)
			}
		})
	}
}

func TestMonth(t *testing.T) {
	t.Parallel()
	e, c := adminEnv(t)
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-08-31'), ('2026-10-01')`)

	type month struct {
		Month        string   `json:"month"`
		NoSchoolDays []string `json:"no_school_days"`
		ReportURL    string   `json:"report_url"`
	}

	w := e.do(t, "GET", "/api/months/2026-09", "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	first := decode[month](t, w.Body.String())
	if first.Month != "2026-09" || len(first.NoSchoolDays) != 0 || !strings.HasPrefix(first.ReportURL, "https://school.test/report/") {
		t.Errorf("month = %+v", first)
	}

	if w := e.do(t, "PUT", "/api/months/2026-09", `{"no_school_days":["2026-09-05","2026-09-06"]}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("put: status = %d, body %s", w.Code, w.Body)
	}
	if w := e.do(t, "PUT", "/api/months/2026-09", `{"no_school_days":["2026-09-06","2026-09-13"]}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("put: status = %d, body %s", w.Code, w.Body)
	}

	second := decode[month](t, e.do(t, "GET", "/api/months/2026-09", "", c).Body.String())
	if strings.Join(second.NoSchoolDays, ",") != "2026-09-06,2026-09-13" {
		t.Errorf("no school days = %v", second.NoSchoolDays)
	}
	if second.ReportURL != first.ReportURL {
		t.Errorf("report url changed: %s → %s", first.ReportURL, second.ReportURL)
	}
	other := decode[month](t, e.do(t, "GET", "/api/months/2026-10", "", c).Body.String())
	if other.ReportURL == first.ReportURL || strings.Join(other.NoSchoolDays, ",") != "2026-10-01" {
		t.Errorf("october = %+v", other)
	}

	for _, tt := range []struct{ method, path, body string }{
		{"GET", "/api/months/2026-13", ""},
		{"GET", "/api/months/september", ""},
		{"PUT", "/api/months/2026-09", `{"no_school_days":["2026-10-01"]}`},
		{"PUT", "/api/months/2026-09", `{"no_school_days":["26.09.2026"]}`},
	} {
		if w := e.do(t, tt.method, tt.path, tt.body, c); w.Code != http.StatusBadRequest {
			t.Errorf("%s %s %s: status = %d, want 400", tt.method, tt.path, tt.body, w.Code)
		}
	}
}

type attendanceDay struct {
	Date      string `json:"date"`
	Published bool   `json:"published"`
	Closed    bool   `json:"closed"`
	School    bool   `json:"school"`
}

type attendanceRow struct {
	ChildID int64             `json:"child_id"`
	Active  bool              `json:"active"`
	Marks   map[string]string `json:"marks"`
	Total   int64             `json:"total"`
	Paid    bool              `json:"paid"`
}

type attendance struct {
	Days     []attendanceDay `json:"days"`
	Children []attendanceRow `json:"children"`
	Message  *struct {
		Text string `json:"text"`
	} `json:"message"`
	Sent bool `json:"sent"`
}

func attendanceEnv(t *testing.T) (*env, *http.Cookie) {
	t.Helper()
	e, c := adminEnv(t)
	e.a.Exec(t, `INSERT INTO children (id, last_name, first_name, price, active, weekdays) VALUES
		(1, 'Иванов', 'Петя', 13000, 1, 127),
		(2, 'Козлова', 'Аня', 15000, 1, 127),
		(3, 'Сидоров', 'Коля', 13000, 0, 127),
		(4, 'Орлов', 'Гриша', 13000, 0, 127)`)
	return e, c
}

func getAttendance(t *testing.T, e *env, c *http.Cookie) attendance {
	t.Helper()
	w := e.do(t, "GET", "/api/attendance/2026-09", "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("get: status = %d, body %s", w.Code, w.Body)
	}
	return decode[attendance](t, w.Body.String())
}

func putAttendance(t *testing.T, e *env, c *http.Cookie, body string) {
	t.Helper()
	if w := e.do(t, "PUT", "/api/attendance/2026-09", body, c); w.Code != http.StatusNoContent {
		t.Fatalf("put: status = %d, body %s", w.Code, w.Body)
	}
}

func dayMarks(t *testing.T, e *env, date string) string {
	t.Helper()
	rows, err := e.a.DB.Query(`SELECT child_id, status, price FROM attendance WHERE date = ? ORDER BY child_id`, date)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var parts []string
	for rows.Next() {
		var (
			id, price int64
			status    string
		)
		if err := rows.Scan(&id, &status, &price); err != nil {
			t.Fatal(err)
		}
		parts = append(parts, itoa(id)+":"+status+":"+itoa(price))
	}
	return strings.Join(parts, " ")
}

func TestAttendanceGet(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-05'), ('2026-09-06')`)
	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES
		('2026-09-25', 'closed', 'm1'), ('2026-09-24', 'open', 'm2'), ('2026-09-23', 'open', NULL)`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-25', 1, 'present', 12000), ('2026-09-24', 1, 'present', 13000),
		('2026-09-25', 3, 'absent', 0), ('2026-08-31', 4, 'absent', 0), ('2026-08-31', 1, 'present', 13000)`)

	a := getAttendance(t, e, c)

	if a.Message == nil || !strings.HasPrefix(a.Message.Text, "Питание за сентябрь 2026.") || a.Sent {
		t.Errorf("message = %+v, sent = %v; want unsent summary", a.Message, a.Sent)
	}
	if len(a.Days) != 30 || a.Days[0].Date != "2026-09-01" || !a.Days[3].School || a.Days[4].School || a.Days[5].School || !a.Days[6].School {
		t.Fatalf("days = %+v", a.Days)
	}
	states := map[string]attendanceDay{}
	for _, d := range a.Days {
		states[d.Date] = d
	}
	if d := states["2026-09-25"]; !d.Published || !d.Closed {
		t.Errorf("25: %+v", d)
	}
	if d := states["2026-09-24"]; !d.Published || d.Closed {
		t.Errorf("24: %+v", d)
	}
	if d := states["2026-09-23"]; d.Published {
		t.Errorf("23: %+v", d)
	}

	var ids []int64
	for _, r := range a.Children {
		ids = append(ids, r.ChildID)
	}
	if !equalIDs(ids, 1, 2, 3) {
		t.Errorf("children = %v, want 1 2 3", ids)
	}
	if a.Children[0].Marks["2026-09-25"] != "present" || a.Children[0].Total != 25000 ||
		len(a.Children[1].Marks) != 0 || !a.Children[1].Active ||
		a.Children[2].Marks["2026-09-25"] != "absent" || a.Children[2].Active {
		t.Errorf("children = %+v", a.Children)
	}

	past := decode[attendance](t, e.do(t, "GET", "/api/attendance/2026-08", "", c).Body.String())
	ids = nil
	for _, r := range past.Children {
		ids = append(ids, r.ChildID)
	}
	if !equalIDs(ids, 1, 4) {
		t.Errorf("past month children = %v, want 1 4", ids)
	}
}

func TestAttendanceUnpublishedDaySavesOnlyGivenCells(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)

	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-24","child_id":2,"status":"absent"}]}`)

	if got, want := dayMarks(t, e, "2026-09-24"), "2:absent:0"; got != want {
		t.Errorf("marks = %s, want %s", got, want)
	}
	if got := dayMarks(t, e, "2026-09-23"); got != "" {
		t.Errorf("untouched day marks = %s, want none", got)
	}
}

func TestAttendanceFutureAbsent(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)

	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-29","child_id":2,"status":"absent"}]}`)
	if got, want := dayMarks(t, e, "2026-09-29"), "2:absent:0"; got != want {
		t.Errorf("marks = %s, want %s", got, want)
	}

	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-29","child_id":2,"status":null}]}`)
	if got := dayMarks(t, e, "2026-09-29"); got != "" {
		t.Errorf("after clearing: marks = %s, want none", got)
	}
}

func TestAttendancePublishedDayChangesOnlyGivenCells(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-25', 'closed', 'm1')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-25', 1, 'present', 12000), ('2026-09-25', 2, 'present', 15000), ('2026-09-25', 3, 'absent', 0)`)
	e.a.Exec(t, `UPDATE children SET price = 20000`)

	putAttendance(t, e, c, `{"marks":[
		{"date":"2026-09-25","child_id":1,"status":"present"},
		{"date":"2026-09-25","child_id":2,"status":"absent"},
		{"date":"2026-09-25","child_id":3,"status":"present"}]}`)

	if got, want := dayMarks(t, e, "2026-09-25"), "1:present:12000 2:absent:0 3:present:20000"; got != want {
		t.Errorf("marks = %s, want %s", got, want)
	}

	edit := e.a.MAX.NextEdit(t)
	if edit.MessageID != "m1" || !strings.Contains(edit.Message.Text, "❌ Не обедают (1):\nКозлова Аня") {
		t.Errorf("edit %s:\n%s", edit.MessageID, edit.Message.Text)
	}

	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-25","child_id":2,"status":null}]}`)
	if got, want := dayMarks(t, e, "2026-09-25"), "1:present:12000 3:present:20000"; got != want {
		t.Errorf("after clearing: marks = %s, want %s", got, want)
	}
}

func TestAttendanceRedrawsMonthSummary(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO report_months (month, uid, message_id) VALUES ('2026-09', 'u1', 'm9')`)

	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-24","child_id":2,"status":"present"}]}`)

	edit := e.a.MAX.NextEdit(t)
	if edit.MessageID != "m9" || !strings.Contains(edit.Message.Text, "🔴 <b>Козлова Аня</b>: 1 день × 150\u00a0₽ = 150\u00a0₽") {
		t.Errorf("edit %s:\n%s", edit.MessageID, edit.Message.Text)
	}
	if a := getAttendance(t, e, c); a.Message == nil || a.Message.Text != edit.Message.Text || !a.Sent {
		t.Errorf("message = %+v, want %q", a.Message, edit.Message.Text)
	}
}

func TestAttendancePayments(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO report_months (month, uid, message_id) VALUES ('2026-09', 'u1', 'm9')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES ('2026-09-24', 2, 'present', 15000)`)

	putAttendance(t, e, c, `{"marks":[],"payments":[{"child_id":2,"paid":true}]}`)

	edit := e.a.MAX.NextEdit(t)
	if edit.MessageID != "m9" || !strings.Contains(edit.Message.Text, "🟢 <b>Козлова Аня</b>: 1 день × 150\u00a0₽ = 150\u00a0₽") {
		t.Errorf("edit %s:\n%s", edit.MessageID, edit.Message.Text)
	}
	if a := getAttendance(t, e, c); len(a.Children) != 2 || a.Children[0].Paid || !a.Children[1].Paid {
		t.Errorf("children = %+v, want only 2 paid", a.Children)
	}

	putAttendance(t, e, c, `{"marks":[],"payments":[{"child_id":2,"paid":false}]}`)
	if edit := e.a.MAX.NextEdit(t); !strings.Contains(edit.Message.Text, "🔴 <b>Козлова Аня</b>: 1 день × 150\u00a0₽ = 150\u00a0₽") {
		t.Errorf("after unpaid:\n%s", edit.Message.Text)
	}
	if a := getAttendance(t, e, c); a.Children[0].Paid || a.Children[1].Paid {
		t.Errorf("children = %+v, want none paid", a.Children)
	}

	if w := e.do(t, "PUT", "/api/attendance/2026-09", `{"marks":[],"payments":[{"child_id":99,"paid":true}]}`, c); w.Code != http.StatusBadRequest {
		t.Errorf("unknown child: status = %d, want 400", w.Code)
	}
}

func TestAttendanceErrors(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-20')`)

	tests := []struct {
		name, path, body string
	}{
		{"future", "2026-09", `{"marks":[{"date":"2026-09-27","child_id":1,"status":"present"}]}`},
		{"no school day", "2026-09", `{"marks":[{"date":"2026-09-20","child_id":1,"status":"present"}]}`},
		{"other month", "2026-09", `{"marks":[{"date":"2026-08-31","child_id":1,"status":"present"}]}`},
		{"bad date", "2026-09", `{"marks":[{"date":"25.09.2026","child_id":1,"status":"present"}]}`},
		{"bad month", "09.2026", `{"marks":[]}`},
		{"unknown child", "2026-09", `{"marks":[
			{"date":"2026-09-25","child_id":1,"status":"present"},
			{"date":"2026-09-25","child_id":99,"status":"present"}]}`},
		{"bad status", "2026-09", `{"marks":[{"date":"2026-09-25","child_id":1,"status":"maybe"}]}`},
	}
	for _, tt := range tests {
		if w := e.do(t, "PUT", "/api/attendance/"+tt.path, tt.body, c); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400, body %s", tt.name, w.Code, w.Body)
		}
	}

	var n int
	if err := e.a.DB.QueryRow(`SELECT COUNT(*) FROM attendance`).Scan(&n); err != nil || n != 0 {
		t.Errorf("attendance rows = %d (%v), want 0", n, err)
	}
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func equalIDs(got []int64, want ...int64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestDebugSend(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)

	if w := e.do(t, "POST", "/api/debug/month/2026-09", `{}`, c); w.Code != http.StatusBadRequest {
		t.Errorf("month without marks: status = %d, want 400", w.Code)
	}
	if w := e.do(t, "POST", "/api/debug/day/25.09.2026", `{}`, c); w.Code != http.StatusBadRequest {
		t.Errorf("bad date: status = %d, want 400", w.Code)
	}
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-09-23')`)
	for _, date := range []string{"2026-09-23", "2026-09-27"} {
		if w := e.do(t, "POST", "/api/debug/day/"+date, `{}`, c); w.Code != http.StatusBadRequest {
			t.Errorf("no school or future day %s: status = %d, want 400", date, w.Code)
		}
	}

	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-24","child_id":2,"status":"absent"}]}`)

	for _, path := range []string{"/api/debug/day/2026-09-24", "/api/debug/month/2026-09"} {
		if w := e.do(t, "POST", path, `{}`, c); w.Code != http.StatusNoContent {
			t.Fatalf("%s: status = %d, body %s", path, w.Code, w.Body)
		}
	}

	sent := e.a.MAX.Sent()
	if len(sent) != 2 || !strings.Contains(sent[0].Text, "Козлова Аня") || len(sent[0].Buttons) != 1 ||
		!strings.HasPrefix(sent[1].Text, "Питание за сентябрь 2026.") {
		t.Errorf("sent = %+v", sent)
	}

	var post, summary string
	if err := e.a.DB.QueryRow(`SELECT status || ':' || message_id FROM daily_posts WHERE date = '2026-09-24'`).Scan(&post); err != nil || post != "open:m1" {
		t.Errorf("post = %q (%v), want open:m1", post, err)
	}
	if err := e.a.DB.QueryRow(`SELECT message_id FROM report_months WHERE month = '2026-09'`).Scan(&summary); err != nil || summary != "m2" {
		t.Errorf("summary message = %q (%v), want m2", summary, err)
	}

	for _, path := range []string{"/api/debug/day/2026-09-24", "/api/debug/month/2026-09"} {
		if w := e.do(t, "POST", path, `{}`, c); w.Code != http.StatusBadRequest {
			t.Errorf("%s again: status = %d, want 400", path, w.Code)
		}
	}
	if got := len(e.a.MAX.Sent()); got != 2 {
		t.Errorf("sent %d messages, want 2", got)
	}
}

func TestDebugSendFails(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	putAttendance(t, e, c, `{"marks":[{"date":"2026-09-24","child_id":2,"status":"absent"}]}`)
	e.a.MAX.SetSendError(errors.New("max unavailable"))

	for _, path := range []string{"/api/debug/day/2026-09-25", "/api/debug/month/2026-09"} {
		if w := e.do(t, "POST", path, `{}`, c); w.Code != http.StatusBadGateway {
			t.Errorf("%s: status = %d, want 502, body %s", path, w.Code, w.Body)
		}
	}

	var posts int
	var summary sql.NullString
	if err := e.a.DB.QueryRow(`SELECT (SELECT COUNT(*) FROM daily_posts),
		(SELECT message_id FROM report_months WHERE month = '2026-09')`).Scan(&posts, &summary); err != nil {
		t.Fatal(err)
	}
	if posts != 0 || summary.Valid {
		t.Errorf("posts = %d, summary = %v; want 0, NULL", posts, summary)
	}

	e.a.MAX.SetSendError(nil)
	if w := e.do(t, "POST", "/api/debug/day/2026-09-25", `{}`, c); w.Code != http.StatusNoContent {
		t.Errorf("retry: status = %d, body %s", w.Code, w.Body)
	}
}

func TestDebugClose(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)

	if w := e.do(t, "POST", "/api/debug/close/2026-09-25", `{}`, c); w.Code != http.StatusBadRequest {
		t.Errorf("no post: status = %d, want 400", w.Code)
	}

	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-25', 'open', 'm7')`)
	if w := e.do(t, "POST", "/api/debug/close/2026-09-25", `{}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("close: status = %d, body %s", w.Code, w.Body)
	}

	edit := e.a.MAX.NextEdit(t)
	if edit.MessageID != "m7" || !strings.Contains(edit.Message.Text, "завершён") {
		t.Errorf("edit %s:\n%s", edit.MessageID, edit.Message.Text)
	}
	var n int
	if err := e.a.DB.QueryRow(`SELECT COUNT(*) FROM attendance WHERE date = '2026-09-25' AND status = 'present'`).Scan(&n); err != nil || n != 2 {
		t.Errorf("present marks = %d (%v), want 2", n, err)
	}

	if w := e.do(t, "POST", "/api/debug/close/2026-09-25", `{}`, c); w.Code != http.StatusBadRequest {
		t.Errorf("closed twice: status = %d, want 400", w.Code)
	}
}

func TestDebugSummaries(t *testing.T) {
	t.Parallel()
	e, c := adminEnv(t)
	e.a.Exec(t, `INSERT INTO report_months (month, uid, message_id) VALUES ('2026-08', 'u1', 'm1'), ('2026-09', 'u2', NULL)`)

	w := e.do(t, "GET", "/api/debug/summaries", "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	got := decode[struct{ Months []string }](t, w.Body.String())
	if len(got.Months) != 1 || got.Months[0] != "2026-08" {
		t.Errorf("months = %v, want [2026-08]", got.Months)
	}
}

func TestDebugDelete(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)

	for _, path := range []string{"/api/debug/day/2026-09-24", "/api/debug/month/2026-09"} {
		if w := e.do(t, "DELETE", path, "", c); w.Code != http.StatusBadRequest {
			t.Errorf("%s without message: status = %d, want 400", path, w.Code)
		}
	}

	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-24', 'closed', 'm7')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES
		('2026-09-24', 2, 'absent', 0), ('2026-09-24', 1, 'pending', 0)`)
	e.a.Exec(t, `INSERT INTO day_clubs (date, child_id, start, name) VALUES ('2026-09-24', 1, '10:00', 'Шахматы')`)
	e.a.Exec(t, `UPDATE report_months SET message_id = 'm8' WHERE month = '2026-09'`)

	for _, path := range []string{"/api/debug/day/2026-09-24", "/api/debug/month/2026-09"} {
		if w := e.do(t, "DELETE", path, "", c); w.Code != http.StatusNoContent {
			t.Fatalf("%s: status = %d, body %s", path, w.Code, w.Body)
		}
	}

	if got := e.a.MAX.Deleted(); len(got) != 2 || got[0] != "m7" || got[1] != "m8" {
		t.Errorf("deleted = %v, want [m7 m8]", got)
	}
	var posts, marks, clubs int
	var summary sql.NullString
	if err := e.a.DB.QueryRow(`SELECT (SELECT COUNT(*) FROM daily_posts), (SELECT COUNT(*) FROM attendance),
		(SELECT COUNT(*) FROM day_clubs), (SELECT message_id FROM report_months WHERE month = '2026-09')`,
	).Scan(&posts, &marks, &clubs, &summary); err != nil {
		t.Fatal(err)
	}
	if posts != 0 || marks != 1 || clubs != 0 || summary.Valid {
		t.Errorf("posts = %d, marks = %d, clubs = %d, summary = %v; want 0, 1, 0, NULL", posts, marks, clubs, summary)
	}

	if w := e.do(t, "POST", "/api/debug/day/2026-09-24", `{}`, c); w.Code != http.StatusNoContent {
		t.Errorf("send after delete: status = %d, body %s", w.Code, w.Body)
	}
}

func TestDebugDeleteMAXFails(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-24', 'closed', 'm7')`)
	e.a.MAX.SetError(errors.New("max unavailable"))

	if w := e.do(t, "DELETE", "/api/debug/day/2026-09-24", "", c); w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502, body %s", w.Code, w.Body)
	}
	var posts int
	if err := e.a.DB.QueryRow(`SELECT COUNT(*) FROM daily_posts`).Scan(&posts); err != nil || posts != 0 {
		t.Errorf("posts = %d (%v), want 0", posts, err)
	}
}
