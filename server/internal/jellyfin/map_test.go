package jellyfin

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"marquee/server/internal/domain"
)

func load(t *testing.T, name string) []Session {
	t.Helper()
	b, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var s []Session
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return s
}

func mapOne(t *testing.T, name string) domain.Session {
	t.Helper()
	m := &Mapper{}
	out := m.Map(load(t, name))
	if len(out) != 1 {
		t.Fatalf("expected 1 session from %s, got %d", name, len(out))
	}
	return out[0]
}

func TestMovieDirect(t *testing.T) {
	s := mapOne(t, "movie_direct.json")

	if s.Kind != "movie" || s.Title != "Fight Club" || s.Subtitle != "1999" || s.Year != 1999 {
		t.Errorf("unexpected identity: %+v", s)
	}
	if s.Network != "lan" || s.IP != "192.168.1.44" {
		t.Errorf("network: got %s / %s", s.Network, s.IP)
	}
	if s.Stream.Method != "direct" || s.Stream.HW != "" {
		t.Errorf("stream: %+v", s.Stream)
	}
	if s.Stream.Bitrate != 1969 {
		t.Errorf("bitrate: %d", s.Stream.Bitrate)
	}
	if s.State != "playing" {
		t.Errorf("state: %s", s.State)
	}
	if math.Abs(s.Progress-0.5) > 0.01 {
		t.Errorf("progress: %f", s.Progress)
	}
	if s.Duration != 8348 {
		t.Errorf("duration: %d", s.Duration)
	}
	if s.Image == nil || s.Image.ID != "item-movie" || s.Image.Tag != "tag-movie" {
		t.Errorf("image: %+v", s.Image)
	}
	if s.Client.OS != "unknown" {
		t.Errorf("os: %q", s.Client.OS)
	}
}

func TestEpisodeTranscodeHW(t *testing.T) {
	s := mapOne(t, "episode_transcode_qsv.json")

	if s.Kind != "episode" {
		t.Fatalf("kind: %s", s.Kind)
	}
	if want := "It's Always Sunny in Philadelphia · S4E2"; s.Subtitle != want {
		t.Errorf("subtitle: got %q want %q", s.Subtitle, want)
	}
	// Episodes fall back to the series image.
	if s.Image == nil || s.Image.ID != "series-1" || s.Image.Tag != "tag-series" {
		t.Errorf("image: %+v", s.Image)
	}
	if s.Network != "tailscale" {
		t.Errorf("network: %s", s.Network)
	}
	if s.Stream.Method != "transcode" || s.Stream.HW != "qsv" {
		t.Errorf("stream: %+v", s.Stream)
	}
	if s.Stream.Bitrate != 4000 { // transcoder target, not source bitrate
		t.Errorf("bitrate: %d", s.Stream.Bitrate)
	}
	if s.Client.OS != "android" {
		t.Errorf("os: %q", s.Client.OS)
	}
}

func TestTrackPaused(t *testing.T) {
	s := mapOne(t, "track_paused.json")

	if s.Kind != "track" || s.Subtitle != "Aphex Twin" {
		t.Errorf("track identity: %+v", s)
	}
	if s.State != "paused" {
		t.Errorf("state: %s", s.State)
	}
	if s.Network != "tailscale" {
		t.Errorf("network for 100.64.0.9: %s", s.Network)
	}
	if s.Client.OS != "ios" {
		t.Errorf("os: %q", s.Client.OS)
	}
}

func TestLiveTV(t *testing.T) {
	s := mapOne(t, "live_tv.json")

	if s.Kind != "live" || s.Network != "wan" {
		t.Errorf("got kind=%s network=%s", s.Kind, s.Network)
	}
	if s.Duration != 0 || s.Progress != 0 {
		t.Errorf("live should have no runtime: %d %f", s.Duration, s.Progress)
	}
	if s.Stream.Method != "transcode" || s.Stream.HW != "" {
		t.Errorf("software transcode expected: %+v", s.Stream)
	}
	if s.Image != nil {
		t.Errorf("expected no image, got %+v", s.Image)
	}
}

// The whole point of the permissive DTO: weird values and idle sessions must not
// take the snapshot down, and unknown kinds fall through to "other".
func TestEdgeCases(t *testing.T) {
	m := &Mapper{}
	out := m.Map(load(t, "edge_cases.json"))

	if len(out) != 1 {
		t.Fatalf("idle session should be dropped; got %d sessions", len(out))
	}
	s := out[0]
	if s.Kind != "other" {
		t.Errorf("unknown Type: got %s", s.Kind)
	}
	if s.Network != "unknown" {
		t.Errorf("missing endpoint: got %s", s.Network)
	}
	if s.State != "playing" || s.Stream.Method != "direct" {
		t.Errorf("missing PlayState should default to playing/direct, got %s/%s", s.State, s.Stream.Method)
	}
	if s.Client.Name != "Unknown client" {
		t.Errorf("client: %q", s.Client.Name)
	}
	if s.Title != "Dune" {
		t.Errorf("title: %q", s.Title)
	}
}

func TestNetworkClassification(t *testing.T) {
	cases := []struct{ endpoint, want string }{
		{"192.168.1.5", "lan"},
		{"10.0.0.2:8096", "lan"},
		{"172.20.1.1", "lan"},
		{"[fd00::1]:9000", "lan"},
		{"127.0.0.1", "lan"},
		{"100.64.0.1", "tailscale"},
		{"100.127.255.254", "tailscale"},
		{"100.63.0.1", "wan"},
		{"8.8.8.8", "wan"},
		{"[2001:db8::1]:443", "wan"},
		{"", "unknown"},
		{"not-an-ip", "unknown"},
	}
	m := &Mapper{}
	for _, c := range cases {
		if got := m.network(c.endpoint); got != c.want {
			t.Errorf("network(%q) = %q, want %q", c.endpoint, got, c.want)
		}
	}
}

func TestOSInference(t *testing.T) {
	cases := []struct{ device, client, want string }{
		{"Chrome", "Jellyfin Web", "unknown"},
		{"Firefox", "Jellyfin Web", "unknown"},
		{"Windows 11", "Jellyfin Media Player", "windows"},
		{"Macbook Pro", "Jellyfin Media Player", "macos"},
		{"iPhone", "Swiftfin", "ios"},
		{"Apple TV", "Swiftfin", "ios"},
		{"Living Room TV", "Jellyfin Android TV", "android"},
		{"NVIDIA SHIELD", "Jellyfin Android TV", "android"},
		{"Roku Ultra", "Jellyfin", "tv"},
		{"Steam Deck", "Jellyfin", "unknown"},
	}
	for _, c := range cases {
		if got := inferOS(c.device, c.client); got != c.want {
			t.Errorf("inferOS(%q, %q) = %q, want %q", c.device, c.client, got, c.want)
		}
	}
}

func TestHostOf(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.2.3.4", "1.2.3.4"},
		{"1.2.3.4:8096", "1.2.3.4"},
		{"[::1]:8096", "::1"},
		{"[2001:db8::1]", "2001:db8::1"},
		{"", ""},
	}
	for _, c := range cases {
		if got := hostOf(c.in); got != c.want {
			t.Errorf("hostOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
