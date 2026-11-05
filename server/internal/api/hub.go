package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Hub fans the latest snapshot out to every connected SSE subscriber.
// Each subscriber holds a single-slot buffer: a slow client simply gets the
// newest snapshot next time rather than accumulating a backlog.
type Hub struct {
	log  *slog.Logger
	subs map[chan []byte]struct{}
	add  chan chan []byte
	drop chan chan []byte
	pub  chan []byte
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		log:  log,
		subs: make(map[chan []byte]struct{}),
		add:  make(chan chan []byte),
		drop: make(chan chan []byte),
		pub:  make(chan []byte, 1),
	}
}

// Run owns the subscriber map so no mutex is needed. It returns when ctx is done.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for ch := range h.subs {
				close(ch)
				delete(h.subs, ch)
			}
			return

		case ch := <-h.add:
			h.subs[ch] = struct{}{}
			h.log.Debug("sse subscriber added", "subscribers", len(h.subs))

		case ch := <-h.drop:
			if _, ok := h.subs[ch]; ok {
				delete(h.subs, ch)
				close(ch)
			}
			h.log.Debug("sse subscriber removed", "subscribers", len(h.subs))

		case msg := <-h.pub:
			for ch := range h.subs {
				select {
				case ch <- msg:
				default:
					// Latest wins: replace whatever is queued.
					select {
					case <-ch:
					default:
					}
					select {
					case ch <- msg:
					default:
					}
				}
			}
		}
	}
}

// Publish hands a pre-marshalled snapshot to the hub without blocking.
func (h *Hub) Publish(msg []byte) {
	select {
	case h.pub <- msg:
	default:
		select {
		case <-h.pub:
		default:
		}
		select {
		case h.pub <- msg:
		default:
		}
	}
}

// HandleSSE streams every published snapshot to one client.
func (h *Hub) HandleSSE(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	hdr := w.Header()
	hdr.Set("Content-Type", "text/event-stream")
	hdr.Set("Cache-Control", "no-cache")
	hdr.Set("Connection", "keep-alive")
	hdr.Set("X-Accel-Buffering", "no")

	ch := make(chan []byte, 1)
	select {
	case h.add <- ch:
	case <-r.Context().Done():
		return
	}
	defer func() {
		select {
		case h.drop <- ch:
		case <-time.After(time.Second):
		}
	}()

	// Tell the browser to retry quickly; the poll loop is the real cadence.
	eventLog := LoggerFrom(r.Context())

	start := time.Now()
	defer func() {
		eventLog.Info("sse closed", "sse_duration_ms", time.Since(start).Milliseconds())
	}()

	if _, err := w.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	fl.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case msg, ok := <-ch:
			if !ok {
				return
			}
			frame := make([]byte, 0, len(msg)+24)
			frame = append(frame, "event: snapshot\ndata: "...)
			frame = append(frame, msg...)
			frame = append(frame, "\n\n"...)
			if _, err := w.Write(frame); err != nil {
				return
			}
			fl.Flush()

		case <-heartbeat.C:
			if _, err := w.Write([]byte(": heartbeat\n\n")); err != nil {
				return
			}
			fl.Flush()
		}
	}
}
