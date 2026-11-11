package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestDayUpcomingSkipsChildOutsideWeekdays(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `UPDATE children SET weekdays = 10 WHERE id = 2`)

	d := getDay(t, e, c, "2026-09-26")
	if !slices.Equal(dayIDs(d), []int64{1}) || d.Message == nil || strings.Contains(d.Message.Text, "Козлова") {
		t.Errorf("saturday: children = %v, message %+v", dayIDs(d), d.Message)
	}
	if d := getDay(t, e, c, "2026-09-29"); !slices.Equal(dayIDs(d), []int64{1, 2}) {
		t.Errorf("tuesday: children = %v, want 1 2", dayIDs(d))
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

type clubChange struct {
	ChildID   int64   `json:"child_id"`
	LastName  string  `json:"last_name"`
	FirstName string  `json:"first_name"`
	Name      string  `json:"name"`
	Start     *string `json:"start"`
}

type dayChanges struct {
	Changes  []clubChange `json:"changes"`
	Editable bool         `json:"changes_editable"`
	Message  struct {
		Text string `json:"text"`
	} `json:"message"`
}

func getDayChanges(t *testing.T, e *env, c *http.Cookie, date string) dayChanges {
	t.Helper()
	w := e.do(t, "GET", "/api/days/"+date, "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("get day: status = %d, body %s", w.Code, w.Body)
	}
	return decode[dayChanges](t, w.Body.String())
}

func putClubChanges(t *testing.T, e *env, c *http.Cookie, date string, childID int, body string) {
	t.Helper()
	path := "/api/days/" + date + "/children/" + strconv.Itoa(childID) + "/club-changes"
	if w := e.do(t, "PUT", path, body, c); w.Code != http.StatusNoContent {
		t.Fatalf("put %s: status = %d, body %s", path, w.Code, w.Body)
	}
}

func storedClubChanges(t *testing.T, e *env) []string {
	t.Helper()
	rows, err := e.a.DB.Query(`SELECT date, child_id, name, COALESCE(start, '-') FROM club_changes ORDER BY date, child_id, name`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var got []string
	for rows.Next() {
		var date, name, start string
		var childID int
		if err := rows.Scan(&date, &childID, &name, &start); err != nil {
			t.Fatal(err)
		}
		got = append(got, fmt.Sprintf("%s %d %s %s", date, childID, name, start))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestDayClubChanges(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO child_clubs (child_id, weekday, start, name) VALUES
		(1, 2, '14:00', 'Шахматы'), (2, 2, '12:30', 'Танцы'), (2, 4, '12:30', 'Танцы')`)

	if d := getDayChanges(t, e, c, "2026-09-29"); !d.Editable || len(d.Changes) != 0 {
		t.Fatalf("before changes: %+v", d)
	}
	putClubChanges(t, e, c, "2026-09-29", 2, `{"changes":[{"name":"Танцы","start":null}]}`)
	putClubChanges(t, e, c, "2026-09-29", 1, `{"changes":[{"name":" Шахматы ","start":"15:00"},{"name":"Футбол","start":"16:00"}]}`)
	putClubChanges(t, e, c, "2026-09-29", 1, `{"changes":[{"name":"Шахматы","start":"15:00"}]}`)
	putClubChanges(t, e, c, "2026-10-01", 2, `{"changes":[{"name":"Танцы","start":"13:00"}]}`)
	e.a.Exec(t, `INSERT INTO club_changes (date, child_id, name, start) VALUES ('2026-09-28', 1, 'Шахматы', NULL)`)

	d := getDayChanges(t, e, c, "2026-09-29")
	var changes []string
	for _, ch := range d.Changes {
		start := "-"
		if ch.Start != nil {
			start = *ch.Start
		}
		changes = append(changes, fmt.Sprintf("%d %s %s %s %s", ch.ChildID, ch.LastName, ch.FirstName, ch.Name, start))
	}
	if !d.Editable {
		t.Error("changes_editable = false, want true")
	}
	if want := []string{"1 Иванов Петя Шахматы 15:00", "2 Козлова Аня Танцы -"}; !slices.Equal(changes, want) {
		t.Errorf("changes = %v, want %v", changes, want)
	}
	wantClubs := "🎨 Кружки и секции:\n<b>15:00 Шахматы</b>: Иванов Петя\n\n"
	if !strings.Contains(d.Message.Text, wantClubs) {
		t.Errorf("preview:\n%s", d.Message.Text)
	}
	var schedule int
	if err := e.a.DB.QueryRow(`SELECT COUNT(*) FROM child_clubs`).Scan(&schedule); err != nil || schedule != 3 {
		t.Errorf("schedule clubs = %d (%v), want 3 unchanged", schedule, err)
	}

	e.a.SetNow(time.Date(2026, 9, 29, 10, 0, 0, 0, time.Local))
	if w := e.do(t, "POST", "/api/debug/day/2026-09-29", `{}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("publish: status = %d, body %s", w.Code, w.Body)
	}
	if text := e.a.MAX.Sent()[0].Text; !strings.Contains(text, wantClubs) {
		t.Errorf("sent:\n%s", text)
	}
	if got, want := storedClubChanges(t, e), []string{"2026-10-01 2 Танцы 13:00"}; !slices.Equal(got, want) {
		t.Errorf("club changes after publish = %v, want %v", got, want)
	}

	d = getDayChanges(t, e, c, "2026-09-29")
	if d.Editable || len(d.Changes) != 0 {
		t.Errorf("after publish: %+v", d)
	}
	w := e.do(t, "PUT", "/api/days/2026-09-29/children/1/club-changes", `{"changes":[]}`, c)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "Сообщение за этот день уже отправлено — правьте кружки в карточке ребёнка") {
		t.Errorf("put after publish: status = %d, body %s", w.Code, w.Body)
	}
}

func TestDayClubChangesSurviveSendError(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO child_clubs (child_id, weekday, start, name) VALUES (1, 6, '10:00', 'Шахматы')`)
	putClubChanges(t, e, c, "2026-09-26", 1, `{"changes":[{"name":"Шахматы","start":"11:00"}]}`)
	e.a.MAX.SetSendError(errors.New("max unavailable"))

	if w := e.do(t, "POST", "/api/debug/day/2026-09-26", `{}`, c); w.Code != http.StatusBadGateway {
		t.Fatalf("publish: status = %d, want 502, body %s", w.Code, w.Body)
	}
	if got, want := storedClubChanges(t, e), []string{"2026-09-26 1 Шахматы 11:00"}; !slices.Equal(got, want) {
		t.Errorf("club changes after send error = %v, want %v", got, want)
	}

	e.a.MAX.SetSendError(nil)
	if w := e.do(t, "POST", "/api/debug/day/2026-09-26", `{}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("retry: status = %d, body %s", w.Code, w.Body)
	}
	if text := e.a.MAX.Sent()[0].Text; !strings.Contains(text, "<b>11:00 Шахматы</b>: Иванов Петя") || strings.Contains(text, "10:00") {
		t.Errorf("sent:\n%s", text)
	}
	if got := storedClubChanges(t, e); len(got) != 0 {
		t.Errorf("club changes after retry = %v, want none", got)
	}
}

func TestDayClubChangeForChildNotAttending(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `UPDATE children SET weekdays = 31 WHERE id = 2`)
	putClubChanges(t, e, c, "2026-09-26", 2, `{"changes":[{"name":"Футбол","start":"16:00"}]}`)

	wantClubs := "🎨 Кружки и секции:\n<b>16:00 Футбол</b>: Козлова Аня\n\n"
	text := getDayChanges(t, e, c, "2026-09-26").Message.Text
	if !strings.Contains(text, wantClubs) || !strings.Contains(text, "✅ Обедают все 1 человек:\nИванов Петя") {
		t.Errorf("preview:\n%s", text)
	}

	if w := e.do(t, "POST", "/api/debug/day/2026-09-26", `{}`, c); w.Code != http.StatusNoContent {
		t.Fatalf("publish: status = %d, body %s", w.Code, w.Body)
	}
	if text := e.a.MAX.Sent()[0].Text; !strings.Contains(text, wantClubs) || !strings.Contains(text, "✅ Обедают все 1 человек:\nИванов Петя") {
		t.Errorf("sent:\n%s", text)
	}
	if d := getDay(t, e, c, "2026-09-26"); !slices.Equal(dayIDs(d), []int64{1, 2}) || d.Children[1].Status != nil ||
		!slices.Equal(d.Children[1].Clubs, []dayClub{{"16:00", "Футбол"}}) {
		t.Errorf("after publish: %+v", d.Children)
	}
}

func TestDayClubChangesOfInactiveChild(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	putClubChanges(t, e, c, "2026-09-29", 1, `{"changes":[{"name":"Футбол","start":"16:00"}]}`)
	if d := getDayChanges(t, e, c, "2026-09-29"); len(d.Changes) != 1 || !strings.Contains(d.Message.Text, "16:00 Футбол") {
		t.Fatalf("before deactivation: changes = %+v, preview:\n%s", d.Changes, d.Message.Text)
	}

	e.a.Exec(t, `UPDATE children SET active = 0 WHERE id = 1`)
	if d := getDayChanges(t, e, c, "2026-09-29"); len(d.Changes) != 0 || strings.Contains(d.Message.Text, "Футбол") {
		t.Errorf("after deactivation: changes = %+v, preview:\n%s", d.Changes, d.Message.Text)
	}
}

func TestDayClubChangesErrors(t *testing.T) {
	t.Parallel()
	e, c := attendanceEnv(t)
	e.a.Exec(t, `INSERT INTO no_school_days (date) VALUES ('2026-10-02')`)

	for _, date := range []string{"2026-09-25", "2026-10-02"} {
		if getDayChanges(t, e, c, date).Editable {
			t.Errorf("%s: changes_editable = true, want false", date)
		}
	}
	for _, tt := range []struct{ name, path, body string }{
		{"past date", "/api/days/2026-09-25/children/1/club-changes", `{"changes":[]}`},
		{"no school day", "/api/days/2026-10-02/children/1/club-changes", `{"changes":[{"name":"Шахматы","start":"10:00"}]}`},
		{"blank name", "/api/days/2026-09-28/children/1/club-changes", `{"changes":[{"name":" ","start":null}]}`},
		{"long name", "/api/days/2026-09-28/children/1/club-changes", `{"changes":[{"name":"` + strings.Repeat("я", 101) + `","start":"10:00"}]}`},
		{"bad start", "/api/days/2026-09-28/children/1/club-changes", `{"changes":[{"name":"Шахматы","start":"25:00"}]}`},
		{"duplicate", "/api/days/2026-09-28/children/1/club-changes", `{"changes":[{"name":"Шахматы","start":"10:00"},{"name":" Шахматы","start":null}]}`},
		{"inactive child", "/api/days/2026-09-28/children/3/club-changes", `{"changes":[]}`},
		{"bad date", "/api/days/28.09.2026/children/1/club-changes", `{"changes":[]}`},
	} {
		if w := e.do(t, "PUT", tt.path, tt.body, c); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400, body %s", tt.name, w.Code, w.Body)
		}
	}
	for _, path := range []string{"/api/days/2026-09-28/children/99/club-changes", "/api/days/2026-09-28/children/x/club-changes"} {
		if w := e.do(t, "PUT", path, `{"changes":[]}`, c); w.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", path, w.Code)
		}
	}
	if got := storedClubChanges(t, e); len(got) != 0 {
		t.Errorf("club changes = %v, want none", got)
	}
}
