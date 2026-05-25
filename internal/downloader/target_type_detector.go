package downloader

import (
	"net/url"
	"strings"
)

func DetectTargetType(rawURL string) TargetType {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return TargetTypeUnknown
	}
	if !isYouTubeHost(parsed.Host) {
		return TargetTypeUnknown
	}
	if looksLikePlaylistURL(rawURL) {
		return TargetTypePlaylist
	}
	if looksLikeSingleVideoURL(rawURL) {
		return TargetTypeVideo
	}
	if looksLikeChannelURL(rawURL) {
		return TargetTypeChannel
	}
	return TargetTypeUnknown
}

func looksLikePlaylistURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	if parsed.Query().Get("list") == "" {
		return false
	}
	path := strings.ToLower(parsed.Path)
	if strings.Contains(path, "/playlist") || strings.Contains(path, "/watch") {
		return true
	}
	return strings.Contains(strings.ToLower(parsed.Host), "youtu.be")
}

func looksLikeChannelURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	parts := strings.Split(strings.Trim(strings.ToLower(parsed.Path), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return false
	}
	if strings.HasPrefix(parts[0], "@") {
		return true
	}
	if parts[0] == "channel" && len(parts) >= 2 && strings.HasPrefix(strings.ToUpper(parts[1]), "UC") {
		return true
	}
	return parts[0] == "user" || parts[0] == "c"
}

func isYouTubeHost(host string) bool {
	normalized := strings.ToLower(host)
	normalized = strings.TrimPrefix(normalized, "www.")
	return normalized == "youtube.com" || normalized == "youtu.be" || strings.HasSuffix(normalized, ".youtube.com")
}
