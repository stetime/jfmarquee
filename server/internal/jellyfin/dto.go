package jellyfin

// The structs in this file are a deliberate subset of Jellyfin's API. Only fields
// Marquee renders are decoded; encoding/json ignores the rest. Everything that can
// be absent, or whose value set may grow between Jellyfin versions, is a pointer or
// a plain string — never an enum. See internal/jellyfin/map.go for the semantics.

type Session struct {
	Id                 string         `json:"Id"`
	UserName           *string        `json:"UserName"`
	DeviceName         *string        `json:"DeviceName"`
	Client             *string        `json:"Client"`
	ApplicationVersion *string        `json:"ApplicationVersion"`
	RemoteEndPoint     *string        `json:"RemoteEndPoint"`
	NowPlayingItem     *Item          `json:"NowPlayingItem"`
	PlayState          *PlayState     `json:"PlayState"`
	TranscodingInfo    *TranscodeInfo `json:"TranscodingInfo"`
}

type Item struct {
	Id                    string            `json:"Id"`
	Name                  *string           `json:"Name"`
	Type                  *string           `json:"Type"`
	RunTimeTicks          *int64            `json:"RunTimeTicks"`
	ProductionYear        *int32            `json:"ProductionYear"`
	ImageTags             map[string]string `json:"ImageTags"`
	SeriesName            *string           `json:"SeriesName"`
	SeriesId              *string           `json:"SeriesId"`
	SeriesPrimaryImageTag *string           `json:"SeriesPrimaryImageTag"`
	ParentIndexNumber     *int32            `json:"ParentIndexNumber"`
	IndexNumber           *int32            `json:"IndexNumber"`
	Album                 *string           `json:"Album"`
	AlbumArtist           *string           `json:"AlbumArtist"`
	Artists               []string          `json:"Artists"`
	MediaSources          []MediaSource     `json:"MediaSources"`
}

type MediaSource struct {
	Bitrate *int32 `json:"Bitrate"`
}

type PlayState struct {
	PositionTicks *int64  `json:"PositionTicks"`
	IsPaused      *bool   `json:"IsPaused"`
	PlayMethod    *string `json:"PlayMethod"`
}

type TranscodeInfo struct {
	Bitrate                  *int32  `json:"Bitrate"`
	VideoCodec               *string `json:"VideoCodec"`
	AudioCodec               *string `json:"AudioCodec"`
	Width                    *int32  `json:"Width"`
	Height                   *int32  `json:"Height"`
	HardwareAccelerationType *string `json:"HardwareAccelerationType"`
}
