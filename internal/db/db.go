package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"

	"neprodlenka/internal/config"
)

//go:embed migrations/*.sql
var migrations embed.FS

func New(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	dsn := "file:" + cfg.DBPath +
		"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return migrate(ctx, db, log)
		},
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}

func migrate(ctx context.Context, db *sql.DB, log *zap.Logger) error {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, fsys)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	for _, r := range results {
		log.Info("migration applied", zap.String("file", r.Source.Path), zap.Duration("took", r.Duration))
	}

	return nil
}
