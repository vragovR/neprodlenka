package api

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/store"
)

type attendanceDayJSON struct {
	Date      string `json:"date"`
	Published bool   `json:"published"`
	Closed    bool   `json:"closed"`
	School    bool   `json:"school"`
}

type attendanceRowJSON struct {
	ChildID    int64             `json:"child_id"`
	LastName   string            `json:"last_name"`
	FirstName  string            `json:"first_name"`
	MiddleName string            `json:"middle_name"`
	Active     bool              `json:"active"`
	Marks      map[string]string `json:"marks"`
	Total      int64             `json:"total"`
	Paid       bool              `json:"paid"`
}

func (h *handler) getAttendance(w http.ResponseWriter, r *http.Request) {
	month, ok := h.month(w, r)
	if !ok {
		return
	}

	noSchool, err := h.store.NoSchoolDays(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	posts, err := h.store.MonthPosts(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	children, err := h.store.Children(r.Context())
	if err != nil {
		h.internalError(w, err)
		return
	}
	marks, err := h.store.MonthMarks(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	paid, err := h.store.MonthPaid(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}

	days := []attendanceDayJSON{}
	for d := month; d.Month() == month.Month(); d = d.AddDate(0, 0, 1) {
		date := d.Format(time.DateOnly)
		post, hasPost := posts[date]
		days = append(days, attendanceDayJSON{
			Date:      date,
			Published: hasPost && post.MessageID != "",
			Closed:    hasPost && post.Status == store.PostClosed,
			School:    !slices.Contains(noSchool, date),
		})
	}

	byChild := map[int64]map[string]string{}
	totals := map[int64]int64{}
	for _, m := range marks {
		if byChild[m.ChildID] == nil {
			byChild[m.ChildID] = map[string]string{}
		}
		byChild[m.ChildID][m.Date] = string(m.Status)
		totals[m.ChildID] += m.Price
	}

	rows := []attendanceRowJSON{}
	for _, c := range children {
		if !c.Active && len(byChild[c.ID]) == 0 {
			continue
		}
		childMarks := byChild[c.ID]
		if childMarks == nil {
			childMarks = map[string]string{}
		}
		rows = append(rows, attendanceRowJSON{
			ChildID:    c.ID,
			LastName:   c.LastName,
			FirstName:  c.FirstName,
			MiddleName: c.MiddleName,
			Active:     c.Active,
			Marks:      childMarks,
			Total:      totals[c.ID],
			Paid:       paid[c.ID],
		})
	}

	var msgJSON *messageJSON
	msg, rm, ok, err := h.builder.BuildMonth(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if ok {
		msgJSON = toMessageJSON(msg)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"month":    month.Format("2006-01"),
		"days":     days,
		"children": rows,
		"message":  msgJSON,
		"sent":     rm.MessageID != "",
	})
}

func (h *handler) putAttendance(w http.ResponseWriter, r *http.Request) {
	month, ok := h.month(w, r)
	if !ok {
		return
	}

	var req struct {
		Marks []struct {
			Date    string  `json:"date"`
			ChildID int64   `json:"child_id"`
			Status  *string `json:"status"`
		} `json:"marks"`
		Payments []struct {
			ChildID int64 `json:"child_id"`
			Paid    bool  `json:"paid"`
		} `json:"payments"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	noSchool, err := h.store.NoSchoolDays(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}

	today := clock.Today(h.now())
	changes := make([]store.MarkChange, 0, len(req.Marks))
	var dates []time.Time
	for _, m := range req.Marks {
		date, err := time.ParseInLocation(time.DateOnly, m.Date, time.Local)
		switch {
		case err != nil || date.Year() != month.Year() || date.Month() != month.Month():
			writeError(w, http.StatusBadRequest, "validation", "Дата "+m.Date+" не из месяца "+month.Format("2006-01"))
			return
		case date.After(today):
			writeError(w, http.StatusBadRequest, "validation", "Нельзя отмечать будущие дни")
			return
		case slices.Contains(noSchool, m.Date):
			writeError(w, http.StatusBadRequest, "validation", "Неучебный день "+m.Date)
			return
		}

		var status store.Status
		if m.Status != nil {
			status = store.Status(*m.Status)
			if status != store.Present && status != store.Absent {
				writeError(w, http.StatusBadRequest, "validation", "Статус — present, absent или null")
				return
			}
		}
		changes = append(changes, store.MarkChange{Date: date, ChildID: m.ChildID, Status: status})

		if !slices.ContainsFunc(dates, date.Equal) {
			dates = append(dates, date)
		}
	}

	payments := make([]store.PaymentChange, 0, len(req.Payments))
	for _, p := range req.Payments {
		payments = append(payments, store.PaymentChange{ChildID: p.ChildID, Paid: p.Paid})
	}

	err = h.store.SaveMarks(r.Context(), month, changes, payments)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "validation", "Неизвестный ребёнок")
		return
	}
	if err != nil {
		h.internalError(w, err)
		return
	}

	for _, date := range dates {
		h.redrawer.Redraw(date)
	}
	if len(dates) == 0 && len(payments) > 0 {
		h.redrawer.RedrawMonth(month)
	}
	w.WriteHeader(http.StatusNoContent)
}
