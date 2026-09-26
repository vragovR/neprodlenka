package api

import (
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"neprodlenka/internal/bot"
	"neprodlenka/internal/store"
)

func (h *handler) date(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	date, err := time.ParseInLocation(time.DateOnly, r.PathValue("date"), time.Local)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation", "Дата — в формате ГГГГ-ММ-ДД")
		return time.Time{}, false
	}
	return date, true
}

func (h *handler) debugDay(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	school, err := h.store.IsSchoolDay(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !school || date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		writeError(w, http.StatusBadRequest, "validation", "В неучебный день сообщение не отправить")
		return
	}
	published, err := h.bot.Publish(r.Context(), date)
	h.debugPublished(w, published, err, "За эту дату сообщение уже есть")
}

func (h *handler) debugMonth(w http.ResponseWriter, r *http.Request) {
	month, ok := h.month(w, r)
	if !ok {
		return
	}
	marks, err := h.store.MonthMarks(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if len(marks) == 0 {
		writeError(w, http.StatusBadRequest, "validation", "За месяц нет отметок")
		return
	}
	published, err := h.bot.PublishSummary(r.Context(), month)
	h.debugPublished(w, published, err, "Итоги за этот месяц уже опубликованы")
}

func (h *handler) debugPublished(w http.ResponseWriter, published bool, err error, exists string) {
	switch {
	case errors.Is(err, bot.ErrSend):
		h.log.Error("debug send", zap.Error(err))
		writeError(w, http.StatusBadGateway, "max_unavailable", "Не удалось отправить сообщение в MAX")
	case err != nil:
		h.internalError(w, err)
	case !published:
		writeError(w, http.StatusBadRequest, "validation", exists)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *handler) debugDeleteDay(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	post, ok, err := h.store.Post(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !ok || post.MessageID == "" {
		writeError(w, http.StatusBadRequest, "validation", "За эту дату нет сообщения")
		return
	}
	h.debugDelete(w, r, post.MessageID, func() error { return h.store.RemovePost(r.Context(), date) })
}

func (h *handler) debugDeleteMonth(w http.ResponseWriter, r *http.Request) {
	month, ok := h.month(w, r)
	if !ok {
		return
	}
	rm, err := h.store.ReportMonth(r.Context(), month)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if rm.MessageID == "" {
		writeError(w, http.StatusBadRequest, "validation", "Итоги за этот месяц не опубликованы")
		return
	}
	h.debugDelete(w, r, rm.MessageID, func() error { return h.store.SetReportMessageID(r.Context(), month, "") })
}

func (h *handler) debugDelete(w http.ResponseWriter, r *http.Request, messageID string, forget func() error) {
	deleteErr := h.max.Delete(r.Context(), messageID)
	if err := forget(); err != nil {
		h.internalError(w, err)
		return
	}
	if deleteErr != nil {
		h.log.Error("debug delete", zap.String("message_id", messageID), zap.Error(deleteErr))
		writeError(w, http.StatusBadGateway, "max_unavailable", "Запись о сообщении удалена, но из MAX удалить не удалось — удалите сообщение в группе вручную")
		return
	}
	h.log.Info("message deleted by admin", zap.String("message_id", messageID))
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) debugClose(w http.ResponseWriter, r *http.Request) {
	date, ok := h.date(w, r)
	if !ok {
		return
	}
	post, ok, err := h.store.Post(r.Context(), date)
	if err != nil {
		h.internalError(w, err)
		return
	}
	if !ok || post.Status != store.PostOpen {
		writeError(w, http.StatusBadRequest, "validation", "За эту дату нет открытой отметки")
		return
	}
	if err := h.store.ClosePost(r.Context(), date); err != nil {
		h.internalError(w, err)
		return
	}
	h.log.Info("post closed by admin", zap.Time("date", date))
	h.redrawer.Redraw(date)
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) debugSummaries(w http.ResponseWriter, r *http.Request) {
	months, err := h.store.SummaryMonths(r.Context())
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"months": months})
}
