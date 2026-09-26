package api

import (
	"errors"
	"net"
	"net/http"

	"go.uber.org/zap"

	"neprodlenka/internal/auth"
)

const cookieName = "session"

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	ip := clientIP(r)
	token, err := h.auth.Login(r.Context(), req.Login, req.Password, ip)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		h.log.Warn("login failed", zap.String("login", req.Login), zap.String("ip", ip))
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный логин или пароль")
		return
	case errors.Is(err, auth.ErrTooManyAttempts):
		h.log.Warn("login rate limited", zap.String("login", req.Login), zap.String("ip", ip))
		writeError(w, http.StatusTooManyRequests, "too_many_attempts", "Слишком много попыток, попробуйте позже")
		return
	case err != nil:
		h.internalError(w, err)
		return
	}

	h.log.Info("login", zap.String("login", req.Login), zap.String("ip", ip))
	setCookie(w, token, int(auth.TokenTTL.Seconds()))
	writeJSON(w, http.StatusOK, map[string]string{"login": req.Login})
}

func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	setCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) requireAdmin(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Нужно войти")
			return
		}
		admin, err := h.auth.Authenticate(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Нужно войти")
			return
		}
		if err != nil {
			h.internalError(w, err)
			return
		}

		if r.Method != http.MethodGet {
			h.log.Info("admin request", zap.String("login", admin.Login), zap.String("method", r.Method), zap.String("path", r.URL.Path))
		}
		next(w, r)
	})
}

func (h *handler) internalError(w http.ResponseWriter, err error) {
	h.log.Error("request failed", zap.Error(err))
	writeError(w, http.StatusInternalServerError, "internal", "Ошибка сервера")
}

func setCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/api",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
