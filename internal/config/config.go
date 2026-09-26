package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/caarlos0/env/v11"
)

const (
	ModePolling = "polling"
	ModeWebhook = "webhook"
)

var webhookSecret = regexp.MustCompile(`^[A-Za-z0-9_-]{5,256}$`)

type Config struct {
	MaxToken      string    `env:"MAX_TOKEN,notEmpty"`
	MaxChatID     int64     `env:"MAX_CHAT_ID,notEmpty"`
	Mode          string    `env:"MODE,notEmpty"`
	WebhookURL    string    `env:"WEBHOOK_URL,required"`
	WebhookSecret string    `env:"WEBHOOK_SECRET,required"`
	TZ            string    `env:"TZ,notEmpty"`
	JWTSecret     string    `env:"JWT_SECRET,notEmpty"`
	BaseURL       string    `env:"BASE_URL,notEmpty"`
	HTTPAddr      string    `env:"HTTP_ADDR,notEmpty"`
	DBPath        string    `env:"DB_PATH,notEmpty"`
	LogLevel      string    `env:"LOG_LEVEL,notEmpty"`
	LogFormat     string    `env:"LOG_FORMAT,notEmpty"`
	PostTime      TimeOfDay `env:"POST_TIME,notEmpty"`
	PublishUntil  TimeOfDay `env:"PUBLISH_UNTIL,notEmpty"`
	CloseTime     TimeOfDay `env:"CLOSE_TIME,notEmpty"`
	CloseUntil    TimeOfDay `env:"CLOSE_UNTIL,notEmpty"`
	SummaryTime   TimeOfDay `env:"SUMMARY_TIME,notEmpty"`
	SummaryUntil  TimeOfDay `env:"SUMMARY_UNTIL,notEmpty"`

	PaymentCard      string `env:"PAYMENT_CARD,notEmpty"`
	PaymentRecipient string `env:"PAYMENT_RECIPIENT,notEmpty"`
}

func New() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	switch cfg.Mode {
	case ModePolling:
	case ModeWebhook:
		if !strings.HasPrefix(cfg.WebhookURL, "https://") {
			return Config{}, fmt.Errorf("WEBHOOK_URL must start with https:// in webhook mode")
		}
		if !webhookSecret.MatchString(cfg.WebhookSecret) {
			return Config{}, fmt.Errorf("WEBHOOK_SECRET must be 5-256 characters A-Z, a-z, 0-9, _ or - in webhook mode")
		}
	default:
		return Config{}, fmt.Errorf("unknown MODE %q: want polling or webhook", cfg.Mode)
	}
	if !cfg.PostTime.Before(cfg.PublishUntil) || cfg.CloseTime.Before(cfg.PublishUntil) || !cfg.CloseTime.Before(cfg.CloseUntil) {
		return Config{}, fmt.Errorf("want POST_TIME < PUBLISH_UNTIL <= CLOSE_TIME < CLOSE_UNTIL, got %s, %s, %s, %s",
			cfg.PostTime, cfg.PublishUntil, cfg.CloseTime, cfg.CloseUntil)
	}
	if !cfg.SummaryTime.Before(cfg.SummaryUntil) {
		return Config{}, fmt.Errorf("want SUMMARY_TIME < SUMMARY_UNTIL, got %s, %s", cfg.SummaryTime, cfg.SummaryUntil)
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	loc, err := time.LoadLocation(cfg.TZ)
	if err != nil {
		return Config{}, fmt.Errorf("load TZ: %w", err)
	}
	now := time.Now()
	if got, want := now.Format(time.DateTime+" -07:00"), now.In(loc).Format(time.DateTime+" -07:00"); got != want {
		panic(fmt.Sprintf("app time %s differs from TZ=%s time %s", got, cfg.TZ, want))
	}
	return cfg, nil
}
