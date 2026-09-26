package api

import (
	"net/http"

	"neprodlenka/internal/report"
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
