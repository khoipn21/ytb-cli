package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	youtube "github.com/kkdai/youtube/v2"
)

func (c *mediaClient) downloadMedia(ctx context.Context, video Video, outDir string, mode DownloadMode) (<-chan progressUpdate, <-chan error) {
	progressCh := make(chan progressUpdate, 64)
	errCh := make(chan error, 1)
	go func() {
		defer close(progressCh)
		defer close(errCh)

		fullVideo, err := c.client.GetVideoContext(ctx, video.URL)
		if err != nil {
			errCh <- fmt.Errorf("load video metadata %s: %w", video.ID, err)
			return
		}
		format, ext, err := selectFormat(fullVideo, mode)
		if err != nil {
			errCh <- err
			return
		}
		stream, size, err := c.client.GetStreamContext(ctx, fullVideo, format)
		if err != nil {
			errCh <- fmt.Errorf("open media stream %s: %w", video.ID, err)
			return
		}
		defer stream.Close()

		filePath := filepath.Join(outDir, buildFileName(video.Title, video.ID, ext))
		file, err := os.Create(filePath)
		if err != nil {
			errCh <- fmt.Errorf("create output file: %w", err)
			return
		}
		defer file.Close()

		if err := copyWithProgress(ctx, file, stream, size, progressCh); err != nil {
			errCh <- err
			return
		}
		progressCh <- progressUpdate{DownloadedBytes: size, TotalBytes: size, Percent: 1, LogLine: "saved: " + filepath.Base(filePath)}
		errCh <- nil
	}()
	return progressCh, errCh
}

func copyWithProgress(ctx context.Context, dst *os.File, src io.Reader, total int64, progressCh chan<- progressUpdate) error {
	buf := make([]byte, 64*1024)
	var downloaded int64
	start := time.Now()
	lastTick := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			now := time.Now()
			if lastTick.IsZero() || now.Sub(lastTick) >= 250*time.Millisecond {
				seconds := now.Sub(start).Seconds()
				speed := float64(downloaded) / maxFloat(seconds, 0.1)
				progressCh <- progressUpdate{
					DownloadedBytes: downloaded,
					TotalBytes:      total,
					Percent:         percent(downloaded, total),
					Speed:           humanBytes(int64(speed)) + "/s",
					ETA:             estimateETA(downloaded, total, speed),
				}
				lastTick = now
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func selectFormat(video *youtube.Video, mode DownloadMode) (*youtube.Format, string, error) {
	var chosen *youtube.Format
	for i := range video.Formats {
		f := &video.Formats[i]
		if mode == ModeAudio && (f.AudioChannels == 0 || f.QualityLabel != "") {
			continue
		}
		if mode == ModeVideo && (f.AudioChannels == 0 || f.QualityLabel == "") {
			continue
		}
		if chosen == nil || f.Bitrate > chosen.Bitrate {
			chosen = f
		}
	}
	if chosen == nil {
		return nil, "", fmt.Errorf("no stream format found for mode %s", mode)
	}
	return chosen, extensionFromMime(chosen.MimeType), nil
}

func extensionFromMime(mime string) string {
	parts := strings.Split(strings.TrimSpace(strings.Split(mime, ";")[0]), "/")
	if len(parts) == 2 && parts[1] != "" {
		return parts[1]
	}
	return "bin"
}
