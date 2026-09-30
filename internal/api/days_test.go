package api_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

type dayClub struct {
	Start string `json:"start"`
	Name  string `json:"name"`
}

type dayRow struct {
	ChildID int64     `json:"child_id"`
	Active  bool      `json:"active"`
	Status  *string   `json:"status"`
	Clubs   []dayClub `json:"clubs"`
}

type day struct {
	Date      string   `json:"date"`
	School    bool     `json:"school"`
	Published bool     `json:"published"`
	Closed    bool     `json:"closed"`
	Children  []dayRow `json:"children"`
	Message   *struct {
		Text    string   `json:"text"`
		Buttons []string `json:"buttons"`
	} `json:"message"`
}

func getDay(t *testing.T, e *env, c *http.Cookie, date string) day {
	t.Helper()
	w := e.do(t, "GET", "/api/days/"+date, "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("get day: status = %d, body %s", w.Code, w.Body)
	}
	return decode[day](t, w.Body.String())
}

func dayIDs(d day) []int64 {
	var ids []int64
	for _, r := range d.Children {
		ids = append(ids, r.ChildID)
	}
	return ids
}

func TestDay(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO child_clubs (child_id, weekday, start, name) VALUES
		(1, 6, '10:00', 'Шахматы'), (1, 1, '15:00', 'Шахматы'), (3, 6, '11:00', 'Футбол')`)

	d := getDay(t, e, c, "2026-09-26")
	if d.Date != "2026-09-26" || !d.School || d.Published || d.Closed || !slices.Equal(dayIDs(d), []int64{1, 2}) {
		t.Fatalf("before publish: %+v", d)
	}
	if r := d.Children[0]; r.Status != nil || !slices.Equal(r.Clubs, []dayClub{{"10:00", "Шахматы"}}) {
		t.Errorf("before publish child 1: %+v", r)
	}
	if d.Message == nil || !strings.Contains(d.Message.Text, "Доброе утро!") || !strings.Contains(d.Message.Text, "Козлова Аня") ||
		!strings.Contains(d.Message.Text, "🎨 Кружки и секции:\n<b>10:00 Шахматы</b>: Иванов Петя\n\n") {
		t.Errorf("message before publish: %+v", d.Message)
	}
	if w := e.do(t, "PUT", "/api/days/2026-09-26/children/1/clubs", `{"clubs":[]}`, c); w.Code != http.StatusBadRequest {
		t.Errorf("put before publish: status = %d, want 400", w.Code)
	}

	if w := e.do(t, "POST", "/api/debug/day/2026-09-26", `{}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("publish: status = %d, body %s", w.Code, w.Body)
	}
	d = getDay(t, e, c, "2026-09-26")
	if !d.Published || d.Closed || !slices.Equal(dayIDs(d), []int64{1, 2}) {
		t.Fatalf("after publish: %+v", d)
	}
	for _, r := range d.Children {
		if r.Status == nil || *r.Status != "pending" {
			t.Errorf("after publish child %d status = %v, want pending", r.ChildID, r.Status)
		}
	}

	if sent := e.a.MAX.Sent()[0]; d.Message == nil || d.Message.Text != sent.Text || !slices.Equal(d.Message.Buttons, []string{sent.Buttons[0].Text}) {
		t.Errorf("message = %+v, want sent %+v", d.Message, sent)
	}

	for _, req := range []struct{ path, body string }{
		{"/api/days/2026-09-26/children/1/clubs", `{"clubs":[]}`},
		{"/api/days/2026-09-26/children/2/clubs", `{"clubs":[{"start":"12:30","name":" Танцы "},{"start":"12:30","name":"Танцы"}]}`},
		{"/api/days/2026-09-26/children/3/clubs", `{"clubs":[{"start":"11:00","name":"Футбол"}]}`},
	} {
		if w := e.do(t, "PUT", req.path, req.body, c); w.Code != http.StatusNoContent {
			t.Fatalf("put %s: status = %d, body %s", req.path, w.Code, w.Body)
		}
	}
	d = getDay(t, e, c, "2026-09-26")
	if !slices.Equal(dayIDs(d), []int64{1, 2, 3}) {
		t.Fatalf("after put: children = %v, want 1 2 3", dayIDs(d))
	}
	want := [][]dayClub{{}, {{"12:30", "Танцы"}}, {{"11:00", "Футбол"}}}
	for i, r := range d.Children {
		if !slices.Equal(r.Clubs, want[i]) {
			t.Errorf("child %d clubs = %v, want %v", r.ChildID, r.Clubs, want[i])
		}
	}
	if want := "🎨 Кружки и секции:\n<b>11:00 Футбол</b>: Сидоров Коля\n<b>12:30 Танцы</b>: Козлова Аня\n\n"; !strings.Contains(d.Message.Text, want) {
		t.Errorf("message after put:\n%s", d.Message.Text)
	}
	if d.Children[2].Active || d.Children[2].Status != nil {
		t.Errorf("inactive child with club: %+v", d.Children[2])
	}
}

func TestDayPastShowsOnlyMarked(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO child_clubs (child_id, weekday, start, name) VALUES (1, 5, '10:00', 'Шахматы'), (2, 5, '11:00', 'Футбол')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES ('2026-09-25', 1, 'present', 13000)`)

	d := getDay(t, e, c, "2026-09-25")
	if !slices.Equal(dayIDs(d), []int64{1}) {
		t.Errorf("children = %v, want 1", dayIDs(d))
	}
	if d.Message == nil || !strings.Contains(d.Message.Text, "Иванов Петя") || strings.Contains(d.Message.Text, "Козлова") {
		t.Errorf("message: %+v", d.Message)
	}

	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-26', 'open', 'm1')`)
	e.a.Exec(t, `INSERT INTO attendance (date, child_id, status, price) VALUES ('2026-09-26', 1, 'pending', 0)`)
	if d := getDay(t, e, c, "2026-09-26"); !slices.Equal(dayIDs(d), []int64{1}) || strings.Contains(d.Message.Text, "Козлова") {
		t.Errorf("published today: children = %v, message %+v", dayIDs(d), d.Message)
	}
}

func TestDayErrors(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-25', 'open', 'm1')`)

	for _, tt := range []struct{ name, method, path, body string }{
		{"bad date", "GET", "/api/days/25.09.2026", ""},
		{"blank name", "PUT", "/api/days/2026-09-25/children/1/clubs", `{"clubs":[{"start":"10:00","name":" "}]}`},
		{"bad start", "PUT", "/api/days/2026-09-25/children/1/clubs", `{"clubs":[{"start":"25:00","name":"Шахматы"}]}`},
		{"long name", "PUT", "/api/days/2026-09-25/children/1/clubs", `{"clubs":[{"start":"10:00","name":"` + strings.Repeat("я", 101) + `"}]}`},
	} {
		if w := e.do(t, tt.method, tt.path, tt.body, c); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", tt.name, w.Code)
		}
	}
	for _, path := range []string{"/api/days/2026-09-25/children/99/clubs", "/api/days/2026-09-25/children/x/clubs"} {
		if w := e.do(t, "PUT", path, `{"clubs":[]}`, c); w.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", path, w.Code)
		}
	}
}
