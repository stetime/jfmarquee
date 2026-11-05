package api

import (
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"marquee/server/internal/jellyfin"
)

// Deps is everything the HTTP layer needs.
type Deps struct {
	Log    *slog.Logger
	Hub    *Hub
	Poller *Poller
	Client *jellyfin.Client
	// Static is the built client; may be nil while developing with Vite.
	Static fs.FS
}

// NewRouter wires the API and the SPA.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /api/sessions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := d.Poller.Latest()
		if msg == nil {
			// Before the first successful poll there is nothing to say; the client
			// treats an empty list as "connecting".
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"sessions":[],"updatedAt":0,"servingKbps":0,"transcoding":0,"paused":0}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(msg)
	}))

	mux.Handle("GET /api/sessions/stream", http.HandlerFunc(d.Hub.HandleSSE))

	mux.Handle("GET /api/thumb/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		width := jellyfin.ThumbWidth
		if v := r.URL.Query().Get("w"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 2000 {
				width = n
			}
		}

		res, err := d.Client.Image(r.Context(), id, r.URL.Query().Get("tag"), width)
		if err != nil {
			if he, ok := err.(*jellyfin.HTTPError); ok && he.Status == http.StatusNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			LoggerFrom(r.Context()).Warn("thumb failed", "id", id, "err", err)
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		defer res.Body.Close()

		ct := res.Header.Get("Content-Type")
		if ct == "" {
			ct = "image/jpeg"
		}
		body, err := io.ReadAll(io.LimitReader(res.Body, 16<<20))
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		// The tag changes whenever the artwork does, so the response is immutable.
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		_, _ = w.Write(body)
	}))

	mux.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	}))

	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if d.Static == nil {
			http.Error(w, "client not built: run `bun run build`, or use the Vite dev server", http.StatusNotFound)
			return
		}
		serveSPA(d.Static, w, r)
	}))

	return WithLogger(d.Log, AccessLog(d.Log, mux))
}

// serveSPA serves files from fsys and falls back to index.html for client routes.
func serveSPA(fsys fs.FS, w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}

	if f, err := fsys.Open(name); err == nil {
		f.Close()
		if ct := mime.TypeByExtension(path.Ext(name)); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeFileFS(w, r, fsys, name)
		return
	}

	// Client-side route: hand back the shell.
	if _, err := fs.Stat(fsys, "index.html"); err != nil {
		http.Error(w, "client not built", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFileFS(w, r, fsys, "index.html")
}
