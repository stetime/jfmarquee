package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// logKey is the context key for the request-scoped logger.
type logKey struct{}

// LoggerFrom returns the request-scoped logger, falling back to a discard logger.
func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(logKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.New(slog.DiscardHandler)
}

// WithLogger stores the base logger on the request context so handlers can
// attach request-scoped attributes with slog.FromContext.
func WithLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestID()
		l := log.With("request_id", id)
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), logKey{}, l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AccessLog logs one line per request. Streams (SSE) log their own open/close
// events instead, so they are marked and skipped here.
func AccessLog(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip static assets; they are noisy and uninteresting.
		if isAsset(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", rec.bytes,
		}
		if rec.status >= 500 {
			log.Error("http", attrs...)
		} else if rec.status >= 400 {
			log.Warn("http", attrs...)
		} else {
			log.Info("http", attrs...)
		}
	})
}

func isAsset(path string) bool {
	return !strings.HasPrefix(path, "/api/")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Flush keeps SSE working through the recorder.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func requestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}
