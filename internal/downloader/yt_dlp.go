package downloader

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type ytDLPClient struct {
	binPath  string
	baseCmd  string
	baseArgs []string
}

type playlistDump struct {
	Entries []playlistEntry `json:"entries"`
}

type playlistEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type singleVideoDump struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type progressUpdate struct {
	DownloadedBytes int64
	TotalBytes      int64
	Percent         float64
	Speed           string
	ETA             string
	LogLine         string
}

func newYTDLPClient(binPath string) *ytDLPClient {
	if binPath == "" {
		binPath = "yt-dlp"
	}
	return &ytDLPClient{binPath: binPath}
}

func (c *ytDLPClient) validateExecutable() error {
	if path, err := exec.LookPath(c.binPath); err == nil {
		c.baseCmd = path
		c.baseArgs = nil
		return nil
	}
	pythonPath, pyErr := exec.LookPath("python")
	if pyErr == nil {
		check := exec.Command(pythonPath, "-m", "yt_dlp", "--version")
		if runErr := check.Run(); runErr == nil {
			c.baseCmd = pythonPath
			c.baseArgs = []string{"-m", "yt_dlp"}
			return nil
		}
	}
	return fmt.Errorf("yt-dlp not found in PATH and python -m yt_dlp is unavailable")
}

func (c *ytDLPClient) fetchTargetVideos(ctx context.Context, targetURL string) ([]Video, error) {
	if looksLikeSingleVideoURL(targetURL) {
		return c.fetchSingleVideo(ctx, targetURL)
	}
	return c.fetchChannelVideos(ctx, targetURL)
}

func (c *ytDLPClient) fetchChannelVideos(ctx context.Context, channelURL string) ([]Video, error) {
	args := []string{"--flat-playlist", "--dump-single-json", "--playlist-items", "1:99999", "--no-warnings", channelURL}
	cmd := exec.CommandContext(ctx, c.baseCmd, append(c.baseArgs, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate channel videos: %w", err)
	}
	var dump playlistDump
	if err := json.Unmarshal(out, &dump); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp JSON output: %w", err)
	}
	videos := make([]Video, 0, len(dump.Entries))
	for _, entry := range dump.Entries {
		if entry.ID == "" {
			continue
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = entry.ID
		}
		videos = append(videos, Video{ID: entry.ID, Title: title, URL: "https://www.youtube.com/watch?v=" + entry.ID})
	}
	if len(videos) == 0 {
		return nil, fmt.Errorf("no downloadable videos found in URL")
	}
	return videos, nil
}

func (c *ytDLPClient) fetchSingleVideo(ctx context.Context, videoURL string) ([]Video, error) {
	args := []string{"--dump-single-json", "--no-playlist", "--no-warnings", videoURL}
	cmd := exec.CommandContext(ctx, c.baseCmd, append(c.baseArgs, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect single video URL: %w", err)
	}
	var dump singleVideoDump
	if err := json.Unmarshal(out, &dump); err != nil {
		return nil, fmt.Errorf("failed parsing single video metadata: %w", err)
	}
	title := strings.TrimSpace(dump.Title)
	if title == "" {
		title = dump.ID
	}
	return []Video{{ID: dump.ID, Title: title, URL: videoURL}}, nil
}

func (c *ytDLPClient) downloadMedia(ctx context.Context, video Video, outDir string, mode DownloadMode) (<-chan progressUpdate, <-chan error) {
	progressCh := make(chan progressUpdate, 64)
	errCh := make(chan error, 1)
	go func() {
		defer close(progressCh)
		defer close(errCh)
		cmd := exec.CommandContext(ctx, c.baseCmd, append(c.baseArgs, buildDownloadArgs(video.URL, outDir, mode)...)...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errCh <- fmt.Errorf("failed creating stdout pipe: %w", err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			errCh <- fmt.Errorf("failed creating stderr pipe: %w", err)
			return
		}
		if err := cmd.Start(); err != nil {
			errCh <- fmt.Errorf("failed to start yt-dlp: %w", err)
			return
		}
		doneCh := make(chan struct{}, 2)
		readPipe(bufio.NewScanner(stdout), progressCh, doneCh)
		readPipe(bufio.NewScanner(stderr), progressCh, doneCh)
		<-doneCh
		<-doneCh
		if waitErr := cmd.Wait(); waitErr != nil {
			errCh <- fmt.Errorf("yt-dlp failed for %s: %w", video.ID, waitErr)
			return
		}
		errCh <- nil
	}()
	return progressCh, errCh
}

func buildDownloadArgs(videoURL, outDir string, mode DownloadMode) []string {
	base := []string{"--newline", "--no-warnings", "-P", outDir, videoURL}
	if mode == ModeVideo {
		return append([]string{"-f", "bv*+ba/b", "--merge-output-format", "mp4"}, base...)
	}
	return append([]string{"-x", "--audio-format", "mp3", "--audio-quality", "0"}, base...)
}

func readPipe(scanner *bufio.Scanner, progressCh chan<- progressUpdate, doneCh chan<- struct{}) {
	go func() {
		defer func() { doneCh <- struct{}{} }()
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			progressCh <- parseProgressLine(line)
		}
	}()
}

func looksLikeSingleVideoURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	path := strings.ToLower(parsed.Path)
	if strings.Contains(host, "youtu.be") {
		return true
	}
	if strings.Contains(path, "/watch") && parsed.Query().Get("v") != "" {
		return true
	}
	return strings.Contains(path, "/shorts/") || strings.Contains(path, "/live/")
}
