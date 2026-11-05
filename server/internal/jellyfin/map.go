package jellyfin

import (
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"marquee/server/internal/domain"
)

const (
	ticksPerSecond = 10_000_000
	// thumb width requested from Jellyfin; the client crops to 16:9 in CSS.
	ThumbWidth = 600
)

// tailscaleDefault is the CGNAT range Tailscale uses by default.
var tailscaleDefault = netip.MustParsePrefix("100.64.0.0/10")

var privatePrefixes = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("::1/128"),
}

// Mapper turns Jellyfin DTOs into the shape the client renders. It is pure apart
// from logging, so it is trivially table-testable.
type Mapper struct {
	Log *slog.Logger
	// Tailscale holds the address ranges treated as tailscale. Defaults to the
	// CGNAT /10 when empty.
	Tailscale []netip.Prefix
}

// Map converts a polling result. Idle sessions (no NowPlayingItem) are dropped.
func (m *Mapper) Map(in []Session) []domain.Session {
	out := make([]domain.Session, 0, len(in))

	for _, s := range in {
		if s.NowPlayingItem == nil {
			continue
		}
		sess := m.convert(s)
		out = append(out, sess)

		if m.Log != nil {
			m.Log.Debug("session mapped",
				"session_id", sess.ID,
				"kind", sess.Kind,
				"method", sess.Stream.Method,
				"hw", sess.Stream.HW,
				"network", sess.Network)
		}
	}
	return out
}

func (m *Mapper) convert(s Session) domain.Session {
	item := s.NowPlayingItem

	out := domain.Session{
		ID:      s.Id,
		User:    deref(s.UserName),
		Title:   derefOr(item.Name, "Untitled"),
		IP:      hostOf(deref(s.RemoteEndPoint)),
		Network: m.network(deref(s.RemoteEndPoint)),
		Client:  m.client(s),
		State:   "playing",
	}

	var subtitle string
	out.Kind, subtitle = classify(item)
	if out.Kind == "episode" {
		subtitle = episodeLabel(item)
	}
	out.Subtitle = subtitle
	out.Year = int(int32Or(item.ProductionYear, 0))
	out.Image = imageOf(item)

	var position, runtime int64
	if s.PlayState != nil {
		if s.PlayState.PositionTicks != nil {
			position = *s.PlayState.PositionTicks
		}
		if s.PlayState.IsPaused != nil && *s.PlayState.IsPaused {
			out.State = "paused"
		}
	}
	if item.RunTimeTicks != nil {
		runtime = *item.RunTimeTicks
	}
	if runtime > 0 {
		out.Duration = runtime / ticksPerSecond
		p := float64(position) / float64(runtime)
		if p > 0 && p <= 1 {
			out.Progress = p
		}
	}

	out.Stream = streamOf(s)
	return out
}

// ---------- classification ----------

func classify(item *Item) (kind, subtitle string) {
	year := ""
	if item.ProductionYear != nil && *item.ProductionYear > 0 {
		year = strconv.Itoa(int(*item.ProductionYear))
	}

	switch deref(item.Type) {
	case "Movie":
		return "movie", year
	case "Episode":
		return "episode", ""
	case "Audio", "MusicVideo":
		if item.AlbumArtist != nil && *item.AlbumArtist != "" {
			return "track", *item.AlbumArtist
		}
		if len(item.Artists) > 0 {
			return "track", strings.Join(item.Artists, ", ")
		}
		if item.Album != nil && *item.Album != "" {
			return "track", *item.Album
		}
		return "track", year
	case "LiveTvChannel", "LiveTvProgram", "Recording":
		return "live", ""
	default:
		return "other", year
	}
}

// episodeLabel renders "Series · S1E4" (or whichever part is known).
func episodeLabel(item *Item) string {
	series := deref(item.SeriesName)

	var part string
	if item.ParentIndexNumber != nil && item.IndexNumber != nil {
		part = fmt.Sprintf("S%dE%d", *item.ParentIndexNumber, *item.IndexNumber)
	} else if item.IndexNumber != nil {
		part = fmt.Sprintf("E%d", *item.IndexNumber)
	}

	switch {
	case series != "" && part != "":
		return series + " · " + part
	case series != "":
		return series
	default:
		return part
	}
}

func imageOf(item *Item) *domain.Image {
	if tag := item.ImageTags["Primary"]; tag != "" {
		return &domain.Image{ID: item.Id, Tag: tag}
	}
	// Episodes usually only carry a series image.
	if item.SeriesId != nil && *item.SeriesId != "" &&
		item.SeriesPrimaryImageTag != nil && *item.SeriesPrimaryImageTag != "" {
		return &domain.Image{ID: *item.SeriesId, Tag: *item.SeriesPrimaryImageTag}
	}
	return nil
}

// ---------- stream ----------

func streamOf(s Session) domain.Stream {
	st := domain.Stream{Method: "direct"}

	// TranscodingInfo is authoritative: Jellyfin sometimes reports DirectStream
	// while a transcode is in fact running.
	if ti := s.TranscodingInfo; ti != nil {
		st.Method = "transcode"
		st.VideoCodec = deref(ti.VideoCodec)
		st.AudioCodec = deref(ti.AudioCodec)
		if ti.HardwareAccelerationType != nil && *ti.HardwareAccelerationType != "" && *ti.HardwareAccelerationType != "none" {
			st.HW = *ti.HardwareAccelerationType
		}
		if ti.Bitrate != nil {
			st.Bitrate = int(*ti.Bitrate) / 1000
		}
	} else if s.PlayState != nil && s.PlayState.PlayMethod != nil {
		switch *s.PlayState.PlayMethod {
		case "DirectPlay":
			st.Method = "direct"
		case "DirectStream":
			st.Method = "remux"
		case "Transcode":
			st.Method = "transcode"
		}
	}

	if st.Bitrate == 0 {
		if s.NowPlayingItem != nil && len(s.NowPlayingItem.MediaSources) > 0 && s.NowPlayingItem.MediaSources[0].Bitrate != nil {
			st.Bitrate = int(*s.NowPlayingItem.MediaSources[0].Bitrate) / 1000
		}
	}
	return st
}

// ---------- network ----------

func (m *Mapper) network(endpoint string) string {
	host := hostOf(endpoint)
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return "unknown"
	}
	addr = addr.Unmap()

	ranges := m.Tailscale
	if len(ranges) == 0 {
		ranges = []netip.Prefix{tailscaleDefault}
	}
	for _, p := range ranges {
		if p.Contains(addr) {
			return "tailscale"
		}
	}
	for _, p := range privatePrefixes {
		if p.Contains(addr) {
			return "lan"
		}
	}
	return "wan"
}

// hostOf strips a port from "ip", "ip:port" or "[v6]:port".
func hostOf(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(endpoint); err == nil {
		return host
	}
	return strings.TrimSuffix(strings.TrimPrefix(endpoint, "["), "]")
}

// ---------- client ----------

func (m *Mapper) client(s Session) domain.Client {
	name := deref(s.Client)
	if name == "" {
		name = "Unknown client"
	}
	return domain.Client{
		Name:   name,
		Device: deref(s.DeviceName),
		OS:     inferOS(deref(s.DeviceName), deref(s.Client)),
	}
}

func inferOS(device, client string) string {
	hay := strings.ToLower(device + " " + client)

	switch {
	case containsAny(hay, "windows"):
		return "windows"
	case containsAny(hay, "macos", "mac os", "osx", "macbook", "imac"):
		return "macos"
	case containsAny(hay, "iphone", "ipad", "ios", "apple tv", "tvos", "infuse"):
		return "ios"
	case containsAny(hay, "android", "fire tv", "firetv", "shield", "chromecast"):
		return "android"
	case containsAny(hay, "roku", "tizen", "webos", "web os", "smart tv", "smarttv", "tv"):
		return "tv"
	case containsAny(hay, "linux", "ubuntu", "debian", "fedora", "arch", "flatpak"):
		return "linux"
	default:
		return "unknown"
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ---------- small helpers ----------

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func derefOr[T ~string](p *T, fallback T) T {
	if p == nil || *p == "" {
		return fallback
	}
	return *p
}

func int32Or(p *int32, fallback int32) int32 {
	if p == nil {
		return fallback
	}
	return *p
}
