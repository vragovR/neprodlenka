package api

import (
	"net/http"
	"slices"
	"strconv"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/report"
	"neprodlenka/internal/store"
)

type childClubChangeJSON struct {
	Date  string  `json:"date"`
	Name  string  `json:"name"`
	Start *string `json:"start"`
}

func (h *handler) getReport(w http.ResponseWriter, r *http.Request) {
	month, ok, err := h.store.MonthByReportUID(r.Context(), r.PathValue("uid"))
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Отчёт не найден")
		return
	}

	days, err := h.store.NoSchoolDays(r.Context(), month)
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

	writeJSON(w, http.StatusOK, report.Build(month, days, children, marks))
}

func (h *handler) getChildReport(w http.ResponseWriter, r *http.Request) {
	month, ok, err := h.store.MonthByReportUID(r.Context(), r.PathValue("uid"))
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Отчёт не найден")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	}
	child, ok, err := h.store.ChildByID(r.Context(), id)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	}
	days, err := h.store.NoSchoolDays(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	marks, err := h.store.MonthMarks(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	clubs, err := h.store.ChildClubs(r.Context(), child.ID)
	if err != nil {
		h.internalError(w, err)
		return
	}
	changes, err := h.store.ChildClubChanges(r.Context(), child.ID, clock.Today(h.now()))
	if err != nil {
		h.internalError(w, err)
		return
	}
	marks = slices.DeleteFunc(marks, func(m store.Mark) bool { return m.ChildID != child.ID })

	clubsJSON := []clubJSON{}
	for _, c := range clubs {
		if slices.Contains(child.Weekdays, c.Weekday) {
			clubsJSON = append(clubsJSON, clubJSON(c))
		}
	}
	changesJSON := make([]childClubChangeJSON, 0, len(changes))
	for _, c := range changes {
		row := childClubChangeJSON{Date: c.Date, Name: c.Name}
		if !c.Canceled() {
			row.Start = new(c.Start)
		}
		changesJSON = append(changesJSON, row)
	}

	rep := report.BuildChild(month, days, child, marks)
	row := rep.Children[0]
	writeJSON(w, http.StatusOK, map[string]any{
		"month":        rep.Month,
		"days":         rep.Days,
		"child":        row,
		"price":        row.MealPrice(),
		"clubs":        clubsJSON,
		"club_changes": changesJSON,
	})
}
