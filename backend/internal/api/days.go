package api

import (
	"cmp"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/message"
	"neprodlenka/internal/store"
)

type dayClubJSON struct {
	Start string `json:"start"`
	Name  string `json:"name"`
}

type clubChangeJSON struct {
	ChildID   int64   `json:"child_id"`
	LastName  string  `json:"last_name"`
	FirstName string  `json:"first_name"`
	Name      string  `json:"name"`
	Start     *string `json:"start"`
}

type messageJSON struct {
	Text    string   `json:"text"`
	Buttons []string `json:"buttons"`
}

func toMessageJSON(msg message.Message) *messageJSON {
	res := &messageJSON{Text: msg.Text, Buttons: []string{}}
	for _, b := range msg.Buttons {
		res.Buttons = append(res.Buttons, b.Text)
	}
	return res
}

type dayRowJSON struct {
	ChildID    int64         `json:"child_id"`
	LastName   string        `json:"last_name"`
	FirstName  string        `json:"first_name"`
	MiddleName string        `json:"middle_name"`
	Active     bool          `json:"active"`
	Status     *string       `json:"status"`
	Clubs      []dayClubJSON `json:"clubs"`
}

func (h *handler) getDay(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	school, err := h.store.IsSchoolDay(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	post, hasPost, err := h.store.Post(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	published := hasPost && post.MessageID != ""
	upcoming := !hasPost && !date.Before(clock.Today(h.now()))
	children, err := h.store.Children(r.Context())
	if err != nil {
		h.internalError(w, err)
		return
	}
	dayChildren, err := h.store.DayChildren(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	statuses := map[int64]store.Status{}
	for _, c := range dayChildren {
		statuses[c.ID] = c.Status
	}

	dayClubs := h.store.PlannedClubs
	if published {
		dayClubs = h.store.DayClubs
	}
	list, err := dayClubs(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	clubs := map[int64][]dayClubJSON{}
	for _, c := range list {
		clubs[c.ChildID] = append(clubs[c.ChildID], dayClubJSON{Start: c.Start, Name: c.Name})
	}

	rows := []dayRowJSON{}
	for _, c := range children {
		status := statuses[c.ID]
		childClubs := clubs[c.ID]
		marked := status != ""
		expected := upcoming && c.Active && c.Attends(date)
		withClubs := published && len(childClubs) > 0
		if !marked && !expected && !withClubs {
			continue
		}
		if childClubs == nil {
			childClubs = []dayClubJSON{}
		}
		row := dayRowJSON{
			ChildID:    c.ID,
			LastName:   c.LastName,
			FirstName:  c.FirstName,
			MiddleName: c.MiddleName,
			Active:     c.Active,
			Clubs:      childClubs,
		}
		if marked {
			row.Status = new(string(status))
		}
		rows = append(rows, row)
	}

	var msgJSON *messageJSON
	if school {
		msg, _, ok, err := h.builder.Build(r.Context(), date)
		if err == nil && !ok {
			msg, err = h.builder.BuildDay(r.Context(), date, false, !upcoming)
		}
		if err != nil {
			h.internalError(w, err)
			return
		}
		msgJSON = toMessageJSON(msg)
	}

	changes, err := h.store.ClubChanges(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	changesJSON := make([]clubChangeJSON, 0, len(changes))
	for _, c := range changes {
		row := clubChangeJSON{ChildID: c.ChildID, LastName: c.LastName, FirstName: c.FirstName, Name: c.Name}
		if !c.Canceled() {
			row.Start = new(c.Start)
		}
		changesJSON = append(changesJSON, row)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"date":             date.Format(time.DateOnly),
		"school":           school,
		"published":        published,
		"closed":           hasPost && post.Status == store.PostClosed,
		"children":         rows,
		"message":          msgJSON,
		"changes":          changesJSON,
		"changes_editable": school && upcoming,
	})
}

func (h *handler) putDayClubs(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	childID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	}
	var req struct {
		Clubs []dayClubJSON `json:"clubs"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	clubs := make([]store.DayClub, 0, len(req.Clubs))
	for _, c := range req.Clubs {
		start, name, err := normalizeClub(c.Start, c.Name)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation", err.Error())
			return
		}
		clubs = append(clubs, store.DayClub{ChildID: childID, Start: start, Name: name})
	}
	slices.SortFunc(clubs, func(a, b store.DayClub) int {
		return cmp.Or(strings.Compare(a.Start, b.Start), strings.Compare(a.Name, b.Name))
	})
	clubs = slices.Compact(clubs)

	post, hasPost, err := h.store.Post(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !hasPost || post.MessageID == "" {
		writeError(w, http.StatusBadRequest, "validation", "Кружки дня можно править после отправки сообщения за день")
		return
	}

	err = h.store.SetDayClubs(r.Context(), date, childID, clubs)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	}
	if err != nil {
		h.internalError(w, err)
		return
	}
	h.redrawer.Redraw(date)
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) putClubChanges(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	childID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	}
	var req struct {
		Changes []struct {
			Name  string  `json:"name"`
			Start *string `json:"start"`
		} `json:"changes"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if date.Before(clock.Today(h.now())) {
		writeError(w, http.StatusBadRequest, "validation", "Изменения кружков вносятся только на сегодня и будущие дни")
		return
	}
	school, err := h.store.IsSchoolDay(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !school {
		writeError(w, http.StatusBadRequest, "validation", "Изменения кружков вносятся только на учебные дни")
		return
	}

	changes := make([]store.ClubChange, 0, len(req.Changes))
	for _, c := range req.Changes {
		var (
			start string
			name  string
			err   error
		)
		if c.Start == nil {
			name, err = normalizeClubName(c.Name)
		} else {
			start, name, err = normalizeClub(*c.Start, c.Name)
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation", err.Error())
			return
		}
		if slices.ContainsFunc(changes, func(ch store.ClubChange) bool { return ch.Name == name }) {
			writeError(w, http.StatusBadRequest, "validation", "Кружок «"+name+"» указан дважды")
			return
		}
		changes = append(changes, store.ClubChange{ChildID: childID, Start: start, Name: name})
	}

	err = h.store.SetClubChanges(r.Context(), date, childID, changes)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
	case errors.Is(err, store.ErrChildInactive):
		writeError(w, http.StatusBadRequest, "validation", "Ребёнок деактивирован")
	case errors.Is(err, store.ErrPostExists):
		writeError(w, http.StatusBadRequest, "validation", "Сообщение за этот день уже отправлено — правьте кружки в карточке ребёнка")
	case err != nil:
		h.internalError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
