package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync/atomic"
	"time"

	"marquee/server/internal/domain"
	"marquee/server/internal/jellyfin"
)

// Poller fetches /Sessions on an interval, maps it, and publishes a snapshot to
// the hub. Failures keep the last good snapshot rather than blanking the UI.
type Poller struct {
	client  *jellyfin.Client
	mapper  *jellyfin.Mapper
	hub     *Hub
	every   time.Duration
	log     *slog.Logger
	latest  atomic.Pointer[[]byte]
	failing int // only touched by the single Run goroutine
}

func NewPoller(client *jellyfin.Client, mapper *jellyfin.Mapper, hub *Hub, every time.Duration, log *slog.Logger) *Poller {
	return &Poller{client: client, mapper: mapper, hub: hub, every: every, log: log}
}

// Latest returns the most recently published snapshot, or nil before the first poll.
func (p *Poller) Latest() []byte {
	if b := p.latest.Load(); b != nil {
		return *b
	}
	return nil
}

// Run polls until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	t := time.NewTicker(p.every)
	defer t.Stop()

	p.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.tick(ctx)
		}
	}
}

func (p *Poller) tick(ctx context.Context) {
	start := time.Now()

	sessions, err := p.client.Sessions(ctx)
	if err != nil {
		p.failing++
		p.log.Warn("poll failed", "err", err, "consecutive_failures", p.failing)
		return
	}

	mapped := p.mapper.Map(sessions)

	snap := domain.Snapshot{
		Sessions:  mapped,
		UpdatedAt: time.Now().UnixMilli(),
	}
	snap.Summarize()

	msg, err := json.Marshal(snap)
	if err != nil {
		p.log.Error("marshal snapshot", "err", err)
		return
	}

	p.latest.Store(&msg)
	p.hub.Publish(msg)

	p.log.Debug("poll",
		"sessions", len(mapped),
		"serving_kbps", snap.ServingKbps,
		"duration_ms", time.Since(start).Milliseconds())

	if p.failing > 0 {
		p.log.Info("poll recovered", "after_failures", p.failing)
		p.failing = 0
	}
}
