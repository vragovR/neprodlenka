package api

import (
	"cmp"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"neprodlenka/internal/message"
	"neprodlenka/internal/store"
)

type dayClubJSON struct {
	Start string `json:"start"`
	Name  string `json:"name"`
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

	clubs := map[int64][]dayClubJSON{}
	if published {
		dayClubs, err := h.store.DayClubs(r.Context(), date)
		if err != nil {
			h.internalError(w, err)
			return
		}
		for _, c := range dayClubs {
			clubs[c.ChildID] = append(clubs[c.ChildID], dayClubJSON{Start: c.Start, Name: c.Name})
		}
	} else {
		weekday := store.Weekday(date)
		for _, c := range children {
			for _, club := range c.Clubs {
				if club.Weekday == weekday && c.Active {
					clubs[c.ID] = append(clubs[c.ID], dayClubJSON{Start: club.Start, Name: club.Name})
				}
			}
		}
	}

	rows := []dayRowJSON{}
	for _, c := range children {
		status := statuses[c.ID]
		childClubs := clubs[c.ID]
		if !c.Active && status == "" && len(childClubs) == 0 {
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
		if status != "" {
			row.Status = new(string(status))
		}
		rows = append(rows, row)
	}

	var msgJSON *messageJSON
	if school {
		msg, _, ok, err := h.builder.Build(r.Context(), date)
		if err == nil && !ok {
			msg, err = h.builder.BuildDay(r.Context(), date, false)
		}
		if err != nil {
			h.internalError(w, err)
			return
		}
		msgJSON = toMessageJSON(msg)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"date":      date.Format(time.DateOnly),
		"school":    school,
		"published": published,
		"closed":    hasPost && post.Status == store.PostClosed,
		"children":  rows,
		"message":   msgJSON,
	})
}

func (h *handler) putDayClubs(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	var req struct {
		Clubs []struct {
			ChildID int64  `json:"child_id"`
			Start   string `json:"start"`
			Name    string `json:"name"`
		} `json:"clubs"`
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
		clubs = append(clubs, store.DayClub{ChildID: c.ChildID, Start: start, Name: name})
	}
	slices.SortFunc(clubs, func(a, b store.DayClub) int {
		return cmp.Or(cmp.Compare(a.ChildID, b.ChildID), strings.Compare(a.Start, b.Start), strings.Compare(a.Name, b.Name))
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

	err = h.store.SetDayClubs(r.Context(), date, clubs)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "validation", "Неизвестный ребёнок")
		return
	}
	if err != nil {
		h.internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
