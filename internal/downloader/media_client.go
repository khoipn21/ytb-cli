package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	youtube "github.com/kkdai/youtube/v2"
)

var channelIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`"channelId":"(UC[a-zA-Z0-9_-]{22})"`),
	regexp.MustCompile(`"externalId":"(UC[a-zA-Z0-9_-]{22})"`),
	regexp.MustCompile(`"browseId":"(UC[a-zA-Z0-9_-]{22})"`),
}

type mediaClient struct {
	client *youtube.Client
	http   *http.Client
}

type progressUpdate struct {
	DownloadedBytes int64
	TotalBytes      int64
	Percent         float64
	Speed           string
	ETA             string
	LogLine         string
}

func newMediaClient() *mediaClient {
	return &mediaClient{
		client: &youtube.Client{},
		http:   &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *mediaClient) fetchTargetVideos(ctx context.Context, targetURL string) ([]Video, error) {
	switch DetectTargetType(targetURL) {
	case TargetTypeVideo:
		video, err := c.client.GetVideoContext(ctx, targetURL)
		if err != nil {
			return nil, fmt.Errorf("load single video: %w", err)
		}
		return []Video{{ID: video.ID, Title: video.Title, URL: targetURL}}, nil
	case TargetTypePlaylist:
		return c.fetchPlaylistVideos(ctx, targetURL)
	}
	channelID, err := c.resolveChannelID(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	uploadsID := "UU" + channelID[2:]
	playlist, err := c.client.GetPlaylistContext(ctx, "https://www.youtube.com/playlist?list="+uploadsID)
	if err != nil {
		return nil, fmt.Errorf("load channel uploads playlist: %w", err)
	}
	videos := make([]Video, 0, len(playlist.Videos))
	for _, entry := range playlist.Videos {
		if entry == nil || entry.ID == "" {
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

func (c *mediaClient) resolveChannelID(ctx context.Context, targetURL string) (string, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "channel" && strings.HasPrefix(parts[1], "UC") {
		return parts[1], nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch channel page: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read channel page: %w", err)
	}
	channelID := extractChannelID(string(body))
	if channelID == "" {
		return "", fmt.Errorf("cannot resolve channel id from URL")
	}
	return channelID, nil
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

func extractChannelID(body string) string {
	for _, pattern := range channelIDPatterns {
		match := pattern.FindStringSubmatch(body)
		if len(match) == 2 {
			return match[1]
		}
	}
	return ""
}
