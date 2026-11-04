package api

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const maxBody = 4096

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func logRequests(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fields := []zap.Field{zap.String("method", r.Method), zap.String("path", r.URL.Path)}
		if r.Body != nil && r.URL.Path != "/api/login" && strings.HasPrefix(r.URL.Path, "/api/") {
			body, _ := io.ReadAll(io.LimitReader(r.Body, maxBody))
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), r.Body))
			if len(body) > 0 {
				fields = append(fields, zap.ByteString("body", body))
			}
		}
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Info("http request", append(fields,
			zap.Int("status", sw.status),
			zap.Duration("took", time.Since(start)),
			zap.String("ip", clientIP(r)),
		)...)
	})
}
