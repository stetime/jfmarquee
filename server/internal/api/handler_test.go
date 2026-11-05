package api

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"marquee/server/internal/jellyfin"
)

func TestSnapshotFlow(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Sessions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"Id":"a","NowPlayingItem":{"Id":"1","Name":"Heat","Type":"Movie","RunTimeTicks":10000000000,"MediaSources":[{"Bitrate":3000000}]},"PlayState":{"IsPaused":false},"RemoteEndPoint":"192.168.0.5"},
			{"Id":"b","NowPlayingItem":null},
			{"Id":"c","NowPlayingItem":{"Id":"2","Name":"Heat 2","Type":"Movie"},"TranscodingInfo":{"Bitrate":2000000,"HardwareAccelerationType":"nvenc"}}
		]`))
	}))
	defer upstream.Close()

	log := slog.New(slog.DiscardHandler)
	hub := NewHub(log)
	client := jellyfin.NewClient(upstream.URL, "test-token")
	poller := NewPoller(client, &jellyfin.Mapper{}, hub, time.Hour, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	go poller.Run(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for poller.Latest() == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	latest := poller.Latest()
	if latest == nil {
		t.Fatal("no snapshot published")
	}

	var snap struct {
		Sessions    []map[string]any `json:"sessions"`
		ServingKbps int              `json:"servingKbps"`
		Transcoding int              `json:"transcoding"`
	}
	if err := json.Unmarshal(latest, &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Sessions) != 2 {
		t.Fatalf("idle session not filtered: %d sessions", len(snap.Sessions))
	}
	if snap.ServingKbps != 5000 {
		t.Errorf("servingKbps = %d, want 5000", snap.ServingKbps)
	}
	if snap.Transcoding != 1 {
		t.Errorf("transcoding = %d, want 1", snap.Transcoding)
	}

	router := NewRouter(Deps{Log: log, Hub: hub, Poller: poller, Client: client})
	sse := httptest.NewServer(router)
	defer sse.Close()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, sse.URL+"/api/sessions/stream", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content type: %s", ct)
	}

	// The poll interval is an hour, so nothing new is published: grab the last
	// snapshot through the JSON route to confirm the router wiring instead.
	jres, err := http.Get(sse.URL + "/api/sessions")
	if err != nil {
		t.Fatal(err)
	}
	defer jres.Body.Close()
	if jres.StatusCode != http.StatusOK {
		t.Fatalf("sessions status: %d", jres.StatusCode)
	}
	var body struct {
		ServingKbps int `json:"servingKbps"`
	}
	if err := json.NewDecoder(jres.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ServingKbps != 5000 {
		t.Errorf("json snapshot servingKbps = %d", body.ServingKbps)
	}

	// SSE: expect the connect comment, proving the stream is live.
	r := bufio.NewReader(res.Body)
	line, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("stream ended early: %v", err)
	}
	if line != ": connected\n" {
		t.Fatalf("unexpected first line %q", line)
	}
}

func TestSSEStopPublisher(t *testing.T) {
	log := slog.New(slog.DiscardHandler)
	hub := NewHub(log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	router := NewRouter(Deps{Log: log, Hub: hub, Poller: NewPoller(nil, nil, hub, time.Hour, log)})
	sse := httptest.NewServer(router)
	defer sse.Close()

	res, err := http.Get(sse.URL + "/api/sessions/stream")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	r := bufio.NewReader(res.Body)
	if line, err := r.ReadString('\n'); err != nil || line != ": connected\n" {
		t.Fatalf("connect line: %q %v", line, err)
	}

	hub.Publish([]byte(`{"sessions":[]}`))

	// Skip the blank line after the comment, then expect the event frame.
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(line, "event: ") {
			if line != "event: snapshot\n" {
				t.Fatalf("event: %q", line)
			}
			break
		}
	}
	data, err := r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if data != "data: {\"sessions\":[]}\n" {
		t.Fatalf("data: %q", data)
	}
}
