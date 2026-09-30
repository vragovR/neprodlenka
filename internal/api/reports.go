package api

import (
	"net/http"
	"slices"
	"strconv"

	"neprodlenka/internal/report"
	"neprodlenka/internal/store"
)

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
	prev, err := h.store.FindReportUID(r.Context(), month.AddDate(0, -1, 0))
	if err != nil {
		h.internalError(w, err)
		return
	}
	next, err := h.store.FindReportUID(r.Context(), month.AddDate(0, 1, 0))
	if err != nil {
		h.internalError(w, err)
		return
	}
	marks = slices.DeleteFunc(marks, func(m store.Mark) bool { return m.ChildID != child.ID })

	rep := report.BuildChild(month, days, child, marks)
	row := rep.Children[0]
	writeJSON(w, http.StatusOK, map[string]any{
		"month":    rep.Month,
		"days":     rep.Days,
		"child":    row,
		"price":    row.MealPrice(),
		"prev_uid": prev,
		"next_uid": next,
	})
}
