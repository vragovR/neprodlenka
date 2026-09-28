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
	if d.Message == nil || !strings.Contains(d.Message.Text, "Доброе утро!") || !strings.Contains(d.Message.Text, "Козлова Аня") {
		t.Errorf("message before publish: %+v", d.Message)
	}
	if w := e.do(t, "PUT", "/api/days/2026-09-26/clubs", `{"clubs":[]}`, c); w.Code != http.StatusBadRequest {
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

	w := e.do(t, "PUT", "/api/days/2026-09-26/clubs", `{"clubs":[
		{"child_id":2,"start":"12:30","name":" Танцы "},
		{"child_id":2,"start":"12:30","name":"Танцы"},
		{"child_id":3,"start":"11:00","name":"Футбол"}]}`, c)
	if w.Code != http.StatusNoContent {
		t.Fatalf("put: status = %d, body %s", w.Code, w.Body)
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
	if d.Children[2].Active || d.Children[2].Status != nil {
		t.Errorf("inactive child with club: %+v", d.Children[2])
	}
}

func TestDayErrors(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO daily_posts (date, status, message_id) VALUES ('2026-09-25', 'open', 'm1')`)

	for _, tt := range []struct{ name, method, path, body string }{
		{"bad date", "GET", "/api/days/25.09.2026", ""},
		{"blank name", "PUT", "/api/days/2026-09-25/clubs", `{"clubs":[{"child_id":1,"start":"10:00","name":" "}]}`},
		{"bad start", "PUT", "/api/days/2026-09-25/clubs", `{"clubs":[{"child_id":1,"start":"25:00","name":"Шахматы"}]}`},
		{"unknown child", "PUT", "/api/days/2026-09-25/clubs", `{"clubs":[{"child_id":99,"start":"10:00","name":"Шахматы"}]}`},
	} {
		if w := e.do(t, tt.method, tt.path, tt.body, c); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", tt.name, w.Code)
		}
	}
}
