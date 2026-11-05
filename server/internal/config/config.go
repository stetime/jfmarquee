package config

import (
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"strings"
	"time"
)

// Config is everything the binary needs from the environment. No env-file
// parsing: the dev scripts export .env.local.
type Config struct {
	JellyfinURL     string
	JellyfinKey     string
	ListenAddr      string
	PollEvery       time.Duration
	LogLevel        slog.Level
	LogFormat       string // "json" | "text"
	TailscaleRanges []netip.Prefix
}

func Load() (Config, error) {
	c := Config{
		JellyfinURL: os.Getenv("JELLYFIN_URL"),
		JellyfinKey: os.Getenv("JELLYFIN_API_KEY"),
		ListenAddr:  envOr("LISTEN_ADDR", ":3000"),
		LogFormat:   strings.ToLower(envOr("LOG_FORMAT", "json")),
	}

	if c.JellyfinURL == "" {
		return c, fmt.Errorf("JELLYFIN_URL is required")
	}
	if c.JellyfinKey == "" {
		return c, fmt.Errorf("JELLYFIN_API_KEY is required (use an admin key, otherwise hardware acceleration is always reported as none)")
	}

	d, err := time.ParseDuration(envOr("POLL_INTERVAL", "3s"))
	if err != nil {
		return c, fmt.Errorf("POLL_INTERVAL: %w", err)
	}
	if d < 500*time.Millisecond {
		return c, fmt.Errorf("POLL_INTERVAL must be at least 500ms, got %s", d)
	}
	c.PollEvery = d

	lvl, err := parseLevel(envOr("LOG_LEVEL", "info"))
	if err != nil {
		return c, err
	}
	c.LogLevel = lvl

	if c.LogFormat != "json" && c.LogFormat != "text" {
		return c, fmt.Errorf("LOG_FORMAT must be json or text, got %q", c.LogFormat)
	}

	if raw := os.Getenv("TAILSCALE_RANGES"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			p, err := netip.ParsePrefix(strings.TrimSpace(part))
			if err != nil {
				return c, fmt.Errorf("TAILSCALE_RANGES: %w", err)
			}
			c.TailscaleRanges = append(c.TailscaleRanges, p)
		}
	}

	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error, got %q", s)
	}
}
