package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"neprodlenka/internal/config"
)

func New(lc fx.Lifecycle, cfg config.Config, h http.Handler, log *zap.Logger) *http.Server {
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("http server started", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("http server stopped", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
