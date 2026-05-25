package downloader

import "time"

type Video struct {
	ID    string
	Title string
	URL   string
}

type DownloadMode string

const (
	ModeAudio DownloadMode = "audio"
	ModeVideo DownloadMode = "video"
	ModeBoth  DownloadMode = "both"
)

type TargetType string

const (
	TargetTypeChannel  TargetType = "channel"
	TargetTypeVideo    TargetType = "video"
	TargetTypePlaylist TargetType = "playlist"
	TargetTypeUnknown  TargetType = "unknown"
)

type EventType string

const (
	EventPlaylistReady EventType = "playlist_ready"
	EventVideoStart    EventType = "video_start"
	EventVideoProgress EventType = "video_progress"
	EventVideoDone     EventType = "video_done"
	EventVideoError    EventType = "video_error"
	EventFinished      EventType = "finished"
	EventLog           EventType = "log"
)

type Event struct {
	Type            EventType
	Timestamp       time.Time
	ChannelURL      string
	RequestURL      string
	Mode            DownloadMode
	Videos          []Video
	Video           Video
	ActiveMode      DownloadMode
	VideoIndex      int
	TotalVideos     int
	DownloadedBytes int64
	TotalBytes      int64
	Percent         float64
	Speed           string
	ETA             string
	Err             error
	Message         string
}
