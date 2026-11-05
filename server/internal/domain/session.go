package domain

// Image points at a poster on the Marquee thumb proxy.
type Image struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

// Client describes the app and device the stream comes from.
type Client struct {
	Name   string `json:"name"`
	Device string `json:"device,omitempty"`
	// OS is inferred from the device/client strings. "unknown" when nothing matched.
	OS string `json:"os"`
}

// Stream is the delivery method plus codec details.
type Stream struct {
	// Method is one of "direct", "remux", "transcode".
	Method string `json:"method"`
	// HW is the hardware accelerator ("qsv", "nvenc", ...) for transcodes,
	// empty for direct play or software transcoding.
	HW         string `json:"hw,omitempty"`
	VideoCodec string `json:"videoCodec,omitempty"`
	AudioCodec string `json:"audioCodec,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"` // kbps
}

// Session is one active playback, as rendered by the client.
type Session struct {
	ID       string `json:"id"`
	User     string `json:"user,omitempty"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	// Kind is one of "movie", "episode", "track", "live", "other".
	Kind     string  `json:"kind"`
	Year     int     `json:"year,omitempty"`
	Image    *Image  `json:"image,omitempty"`
	State    string  `json:"state"`              // "playing" | "paused"
	Progress float64 `json:"progress,omitempty"` // 0..1
	Duration int64   `json:"duration,omitempty"` // seconds
	IP       string  `json:"ip,omitempty"`
	// Network is one of "lan", "wan", "tailscale", "unknown".
	Network string `json:"network"`
	Client  Client `json:"client"`
	Stream  Stream `json:"stream"`
}

// Snapshot is the whole SSE payload: the current sessions plus a few aggregates.
type Snapshot struct {
	Sessions    []Session `json:"sessions"`
	UpdatedAt   int64     `json:"updatedAt"` // unix millis
	ServingKbps int       `json:"servingKbps"`
	Transcoding int       `json:"transcoding"`
	Paused      int       `json:"paused"`
}

// Summarize fills the aggregate fields of s from s.Sessions.
func (s *Snapshot) Summarize() {
	s.ServingKbps, s.Transcoding, s.Paused = 0, 0, 0
	for _, sess := range s.Sessions {
		s.ServingKbps += sess.Stream.Bitrate
		if sess.Stream.Method == "transcode" {
			s.Transcoding++
		}
		if sess.State == "paused" {
			s.Paused++
		}
	}
}
