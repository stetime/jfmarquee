package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"marquee/server/internal/api"
	"marquee/server/internal/config"
	"marquee/server/internal/jellyfin"
	"marquee/server/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	log := newLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := jellyfin.NewClient(cfg.JellyfinURL, cfg.JellyfinKey)
	mapper := &jellyfin.Mapper{Log: log, Tailscale: cfg.TailscaleRanges}
	hub := api.NewHub(log)
	poller := api.NewPoller(client, mapper, hub, cfg.PollEvery, log)

	fsys, err := web.FS()
	if err != nil {
		log.Error("embed client", "err", err)
		os.Exit(1)
	}

	go hub.Run(ctx)
	go poller.Run(ctx)

	handler := api.NewRouter(api.Deps{
		Log:    log,
		Hub:    hub,
		Poller: poller,
		Client: client,
		Static: fsys,
	})

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		// No write timeout: SSE connections are long-lived and heartbeat-checked.
		IdleTimeout: 60 * time.Second,
	}

	log.Info("starting",
		"addr", cfg.ListenAddr,
		"jellyfin", redactURL(cfg.JellyfinURL),
		"poll_interval", cfg.PollEvery.String())

	errs := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutting down", "reason", "signal")
	case err := <-errs:
		log.Error("server", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Warn("shutdown", "err", err)
	}
}

func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var h slog.Handler
	if cfg.LogFormat == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	log := slog.New(h)
	slog.SetDefault(log)
	return log
}

// redactURL keeps scheme+host only so credentials never reach the logs.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid>"
	}
	u.User = nil
	u.RawQuery = ""
	return u.Scheme + "://" + u.Host
}
