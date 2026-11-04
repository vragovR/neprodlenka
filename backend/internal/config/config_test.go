package config

import (
	"maps"
	"strings"
	"testing"
	"time"
)

func setEnv(t *testing.T, override map[string]string) {
	t.Helper()
	vars := map[string]string{
		"MAX_TOKEN":         "token",
		"MAX_CHAT_ID":       "-100",
		"MODE":              "polling",
		"WEBHOOK_URL":       "",
		"WEBHOOK_SECRET":    "",
		"TZ":                "UTC",
		"POST_TIME":         "07:00",
		"PUBLISH_UNTIL":     "08:00",
		"CLOSE_TIME":        "08:00",
		"CLOSE_UNTIL":       "09:00",
		"SUMMARY_TIME":      "12:00",
		"SUMMARY_UNTIL":     "13:00",
		"PAYMENT_CARD":      "1234",
		"PAYMENT_RECIPIENT": "Иванова",
		"JWT_SECRET":        strings.Repeat("s", 32),
		"BASE_URL":          "https://example.com/",
		"HTTP_ADDR":         ":8080",
		"DB_PATH":           "test.db",
		"LOG_LEVEL":         "info",
		"LOG_FORMAT":        "json",
	}
	maps.Copy(vars, override)
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func TestNew(t *testing.T) {
	time.Local = time.UTC

	for _, override := range []map[string]string{
		{},
		{"MODE": "webhook", "WEBHOOK_URL": "https://example.com/webhook", "WEBHOOK_SECRET": "abc_DEF-123"},
	} {
		setEnv(t, override)
		cfg, err := New()
		if err != nil {
			t.Fatalf("%v: %v", override, err)
		}
		if cfg.BaseURL != "https://example.com" || cfg.PostTime != (TimeOfDay{7, 0}) || cfg.CloseUntil != (TimeOfDay{9, 0}) {
			t.Errorf("%v: cfg = %+v", override, cfg)
		}
	}
}

func TestNewErrors(t *testing.T) {
	tests := []struct {
		name     string
		override map[string]string
		want     string
	}{
		{"missing var", map[string]string{"PAYMENT_CARD": ""}, "PAYMENT_CARD"},
		{"bad time", map[string]string{"POST_TIME": "7am"}, "want HH:MM"},
		{"unknown mode", map[string]string{"MODE": "push"}, "unknown MODE"},
		{"webhook over http", map[string]string{"MODE": "webhook", "WEBHOOK_URL": "http://example.com/webhook", "WEBHOOK_SECRET": "secret"}, "WEBHOOK_URL"},
		{"bad webhook secret", map[string]string{"MODE": "webhook", "WEBHOOK_URL": "https://example.com/webhook", "WEBHOOK_SECRET": "bad secret"}, "WEBHOOK_SECRET"},
		{"publish until before post", map[string]string{"PUBLISH_UNTIL": "07:00"}, "POST_TIME < PUBLISH_UNTIL"},
		{"close before publish until", map[string]string{"CLOSE_TIME": "07:59"}, "POST_TIME < PUBLISH_UNTIL"},
		{"close until before close", map[string]string{"CLOSE_UNTIL": "08:00"}, "POST_TIME < PUBLISH_UNTIL"},
		{"summary until before summary", map[string]string{"SUMMARY_UNTIL": "12:00"}, "SUMMARY_TIME < SUMMARY_UNTIL"},
		{"short jwt secret", map[string]string{"JWT_SECRET": strings.Repeat("s", 31)}, "JWT_SECRET"},
		{"bad tz", map[string]string{"TZ": "Europe/Nowhere"}, "TZ"},
	}
	for _, tt := range tests {
		setEnv(t, tt.override)
		if _, err := New(); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want containing %q", tt.name, err, tt.want)
		}
	}
}
