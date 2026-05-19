package downloader

import (
	"context"
	"fmt"
	"time"
)

type Options struct {
	TargetURL string
	OutputDir string
	YTDLPBin  string
	Mode      DownloadMode
}

type Service struct {
	options Options
	client  *ytDLPClient
}

func NewService(options Options) *Service {
	return &Service{
		options: options,
		client:  newYTDLPClient(options.YTDLPBin),
	}
}

func (s *Service) Run(ctx context.Context) <-chan Event {
	events := make(chan Event, 128)

	go func() {
		defer close(events)

		if err := s.client.validateExecutable(); err != nil {
			events <- Event{
				Type:      EventVideoError,
				Timestamp: time.Now(),
				Err:       err,
				Message:   err.Error(),
			}
			return
		}

		mode := normalizeMode(s.options.Mode)
		videos, err := s.client.fetchTargetVideos(ctx, s.options.TargetURL)
		if err != nil {
			events <- Event{
				Type:      EventVideoError,
				Timestamp: time.Now(),
				Err:       err,
				Message:   err.Error(),
			}
			return
		}

		events <- Event{
			Type:        EventPlaylistReady,
			Timestamp:   time.Now(),
			ChannelURL:  s.options.TargetURL,
			RequestURL:  s.options.TargetURL,
			Mode:        mode,
			Videos:      videos,
			TotalVideos: len(videos),
		}

		completed := 0
		for i, video := range videos {
			select {
			case <-ctx.Done():
				events <- Event{
					Type:      EventVideoError,
					Timestamp: time.Now(),
					Err:       ctx.Err(),
					Message:   ctx.Err().Error(),
				}
				return
			default:
			}

			events <- Event{
				Type:        EventVideoStart,
				Timestamp:   time.Now(),
				Video:       video,
				VideoIndex:  i,
				TotalVideos: len(videos),
			}

			progressCh, errCh := s.client.downloadMedia(ctx, video, s.options.OutputDir, mode)
			for upd := range progressCh {
				if upd.LogLine != "" {
					events <- Event{
						Type:       EventLog,
						Timestamp:  time.Now(),
						Video:      video,
						VideoIndex: i,
						Message:    upd.LogLine,
					}
					continue
				}
				events <- Event{
					Type:            EventVideoProgress,
					Timestamp:       time.Now(),
					Mode:            mode,
					Video:           video,
					VideoIndex:      i,
					TotalVideos:     len(videos),
					DownloadedBytes: upd.DownloadedBytes,
					TotalBytes:      upd.TotalBytes,
					Percent:         upd.Percent,
					Speed:           upd.Speed,
					ETA:             upd.ETA,
				}
			}

			if downloadErr := <-errCh; downloadErr != nil {
				events <- Event{
					Type:       EventVideoError,
					Timestamp:  time.Now(),
					Video:      video,
					VideoIndex: i,
					Err:        downloadErr,
					Message:    fmt.Sprintf("failed video %d/%d (%s): %v", i+1, len(videos), video.Title, downloadErr),
				}
				continue
			}

			completed++
			events <- Event{
				Type:        EventVideoDone,
				Timestamp:   time.Now(),
				Mode:        mode,
				Video:       video,
				VideoIndex:  i,
				TotalVideos: len(videos),
				Message:     video.Title,
			}
		}

		events <- Event{
			Type:        EventFinished,
			Timestamp:   time.Now(),
			Mode:        mode,
			TotalVideos: len(videos),
			VideoIndex:  completed,
		}
	}()

	return events
}

func normalizeMode(mode DownloadMode) DownloadMode {
	switch mode {
	case ModeVideo:
		return ModeVideo
	default:
		return ModeAudio
	}
}
