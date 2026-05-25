package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	youtube "github.com/kkdai/youtube/v2"
)

var playlistVideoIDPattern = regexp.MustCompile(`"videoId":"([a-zA-Z0-9_-]{11})"`)

func (c *mediaClient) fetchPlaylistVideos(ctx context.Context, targetURL string) ([]Video, error) {
	playlist, err := c.client.GetPlaylistContext(ctx, targetURL)
	if err == nil {
		videos := buildVideosFromPlaylistEntries(playlist.Videos)
		if len(videos) > 0 {
			return videos, nil
		}
	}

	fallbackVideos, fallbackErr := c.fetchPlaylistVideosFromPage(ctx, targetURL)
	if fallbackErr == nil && len(fallbackVideos) > 0 {
		return fallbackVideos, nil
	}

	if err != nil {
		return nil, fmt.Errorf("load playlist: %s", compactExternalError(err))
	}
	return nil, fmt.Errorf("no downloadable videos found in playlist")
}

func buildVideosFromPlaylistEntries(entries []*youtube.PlaylistEntry) []Video {
	videos := make([]Video, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.ID == "" {
			continue
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = entry.ID
		}
		videos = append(videos, Video{
			ID:    entry.ID,
			Title: title,
			URL:   "https://www.youtube.com/watch?v=" + entry.ID,
		})
	}
	return videos
}

func (c *mediaClient) fetchPlaylistVideosFromPage(ctx context.Context, targetURL string) ([]Video, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch playlist page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read playlist page: %w", err)
	}

	ids := extractPlaylistVideoIDs(string(body))
	if len(ids) == 0 {
		return nil, fmt.Errorf("no video IDs found in playlist page")
	}

	videos := make([]Video, 0, len(ids))
	for _, id := range ids {
		videos = append(videos, Video{
			ID:    id,
			Title: id,
			URL:   "https://www.youtube.com/watch?v=" + id,
		})
	}
	return videos, nil
}

func extractPlaylistVideoIDs(body string) []string {
	matches := playlistVideoIDPattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		id := match[1]
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func compactExternalError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if idx := strings.Index(msg, "\ngoroutine "); idx > 0 {
		msg = msg[:idx]
	}
	if idx := strings.Index(msg, "\n\t"); idx > 0 {
		msg = msg[:idx]
	}
	return strings.TrimSpace(msg)
}
