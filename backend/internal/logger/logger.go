package logger

import (
	"fmt"

	"go.uber.org/zap"

	"neprodlenka/internal/config"
)

func New(cfg config.Config) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parse LOG_LEVEL: %w", err)
	}

	var zcfg zap.Config
	switch cfg.LogFormat {
	case "console":
		zcfg = zap.NewDevelopmentConfig()
	case "json":
		zcfg = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("unknown LOG_FORMAT %q: want console or json", cfg.LogFormat)
	}
	zcfg.Level = level

	return zcfg.Build()
}
