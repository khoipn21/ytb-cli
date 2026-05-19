package downloader

import (
	"context"
	"fmt"
	"strings"
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

		tasks := buildTasks(videos, mode)
		totalTasks := len(tasks)
		events <- Event{
			Type:        EventPlaylistReady,
			Timestamp:   time.Now(),
			ChannelURL:  s.options.TargetURL,
			RequestURL:  s.options.TargetURL,
			Mode:        mode,
			Videos:      tasks,
			TotalVideos: totalTasks,
		}

		completed := 0
		for i, task := range tasks {
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
				Video:       task,
				ActiveMode:  resolveTaskMode(task, mode),
				VideoIndex:  i,
				TotalVideos: totalTasks,
			}

			activeMode := resolveTaskMode(task, mode)
			progressCh, errCh := s.client.downloadMedia(ctx, task, s.options.OutputDir, activeMode)
			for upd := range progressCh {
				if upd.LogLine != "" {
					events <- Event{
						Type:       EventLog,
						Timestamp:  time.Now(),
						Video:      task,
						ActiveMode: activeMode,
						VideoIndex: i,
						Message:    upd.LogLine,
					}
					continue
				}
				events <- Event{
					Type:            EventVideoProgress,
					Timestamp:       time.Now(),
					Mode:            mode,
					Video:           task,
					ActiveMode:      activeMode,
					VideoIndex:      i,
					TotalVideos:     totalTasks,
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
					Video:      task,
					ActiveMode: activeMode,
					VideoIndex: i,
					Err:        downloadErr,
					Message:    fmt.Sprintf("failed item %d/%d (%s %s): %v", i+1, totalTasks, task.Title, activeMode, downloadErr),
				}
				continue
			}

			completed++
			events <- Event{
				Type:        EventVideoDone,
				Timestamp:   time.Now(),
				Mode:        mode,
				Video:       task,
				ActiveMode:  activeMode,
				VideoIndex:  i,
				TotalVideos: totalTasks,
				Message:     task.Title,
			}
		}

		events <- Event{
			Type:        EventFinished,
			Timestamp:   time.Now(),
			Mode:        mode,
			TotalVideos: totalTasks,
			VideoIndex:  completed,
		}
	}()

	return events
}

func normalizeMode(mode DownloadMode) DownloadMode {
	switch mode {
	case ModeBoth:
		return ModeBoth
	case ModeVideo:
		return ModeVideo
	default:
		return ModeAudio
	}
}

func buildTasks(videos []Video, mode DownloadMode) []Video {
	if mode != ModeBoth {
		return videos
	}
	tasks := make([]Video, 0, len(videos)*2)
	for _, video := range videos {
		audioTask := video
		audioTask.Title = video.Title + " [audio]"
		tasks = append(tasks, audioTask)
		videoTask := video
		videoTask.Title = video.Title + " [video]"
		tasks = append(tasks, videoTask)
	}
	return tasks
}

func resolveTaskMode(video Video, selected DownloadMode) DownloadMode {
	title := strings.ToLower(video.Title)
	if strings.HasSuffix(title, "[video]") {
		return ModeVideo
	}
	if strings.HasSuffix(title, "[audio]") {
		return ModeAudio
	}
	if selected == ModeBoth {
		return ModeAudio
	}
	return selected
}
