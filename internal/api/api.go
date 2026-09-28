package api

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"neprodlenka/internal/auth"
	"neprodlenka/internal/bot"
	"neprodlenka/internal/clock"
	"neprodlenka/internal/config"
	"neprodlenka/internal/store"
)

type handler struct {
	auth     *auth.Service
	store    *store.Store
	redrawer *bot.Redrawer
	bot      *bot.Bot
	max      bot.MAX
	now      clock.Clock
	baseURL  string
	secure   bool
	log      *zap.Logger
}

func NewHandler(a *auth.Service, s *store.Store, r *bot.Redrawer, b *bot.Bot, m bot.MAX, wh *bot.Webhook, now clock.Clock, cfg config.Config, log *zap.Logger) http.Handler {
	h := &handler{auth: a, store: s, redrawer: r, bot: b, max: m, now: now, baseURL: cfg.BaseURL, secure: strings.HasPrefix(cfg.BaseURL, "https://"), log: log}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", h.login)
	mux.Handle("DELETE /api/logout", h.requireAdmin(h.logout))
	mux.Handle("GET /api/children", h.requireAdmin(h.listChildren))
	mux.Handle("POST /api/children", h.requireAdmin(h.addChild))
	mux.Handle("PATCH /api/children/{id}", h.requireAdmin(h.updateChild))
	mux.Handle("GET /api/months/{month}", h.requireAdmin(h.getMonth))
	mux.Handle("PUT /api/months/{month}", h.requireAdmin(h.putMonth))
	mux.Handle("GET /api/attendance/{month}", h.requireAdmin(h.getAttendance))
	mux.Handle("PUT /api/attendance/{month}", h.requireAdmin(h.putAttendance))
	mux.Handle("POST /api/debug/day/{date}", h.requireAdmin(h.debugDay))
	mux.Handle("POST /api/debug/month/{month}", h.requireAdmin(h.debugMonth))
	mux.Handle("DELETE /api/debug/day/{date}", h.requireAdmin(h.debugDeleteDay))
	mux.Handle("DELETE /api/debug/month/{month}", h.requireAdmin(h.debugDeleteMonth))
	mux.Handle("POST /api/debug/close/{date}", h.requireAdmin(h.debugClose))
	mux.Handle("GET /api/debug/summaries", h.requireAdmin(h.debugSummaries))
	mux.HandleFunc("GET /api/reports/{uid}", h.getReport)
	mux.HandleFunc("GET /api/reports/{uid}/children/{id}", h.getChildReport)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "Нет такого запроса")
	})

	root := http.NewServeMux()
	root.Handle("POST /webhook", wh)
	root.Handle("/", requireJSON(mux))
	return logRequests(log, root)
}

func requireJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if mediaType != "application/json" {
				writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Нужен Content-Type: application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "Неверный формат запроса")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}
