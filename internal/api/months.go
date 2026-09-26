package api

import (
	"net/http"
	"time"
)

func (h *handler) month(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	month, err := time.ParseInLocation("2006-01", r.PathValue("month"), time.Local)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation", "Месяц — в формате ГГГГ-ММ")
		return time.Time{}, false
	}
	return month, true
}

func (h *handler) getMonth(w http.ResponseWriter, r *http.Request) {
	month, ok := h.month(w, r)
	if !ok {
		return
	}

	days, err := h.store.NoSchoolDays(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	uid, err := h.store.ReportUID(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"month":          month.Format("2006-01"),
		"no_school_days": days,
		"report_url":     h.baseURL + "/report/" + uid,
	})
}

func (h *handler) putMonth(w http.ResponseWriter, r *http.Request) {
	month, ok := h.month(w, r)
	if !ok {
		return
	}

	var req struct {
		NoSchoolDays []string `json:"no_school_days"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	days := make([]time.Time, 0, len(req.NoSchoolDays))
	for _, s := range req.NoSchoolDays {
		d, err := time.ParseInLocation(time.DateOnly, s, time.Local)
		if err != nil || d.Year() != month.Year() || d.Month() != month.Month() {
			writeError(w, http.StatusBadRequest, "validation", "Дата "+s+" не из месяца "+month.Format("2006-01"))
			return
		}
		days = append(days, d)
	}

	if err := h.store.SetNoSchoolDays(r.Context(), month, days); err != nil {
		h.internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
