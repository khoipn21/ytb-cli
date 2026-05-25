package downloader

import (
	"context"
	"errors"
	"strings"
)

func userFacingDownloadError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "download canceled"
	}
	msg := compactExternalError(err)
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "unexpected status code: 429") {
		return "YouTube rate-limited requests (HTTP 429). Wait a few minutes, then retry, or switch network/VPN."
	}
	if strings.Contains(lower, "unexpected status code: 403") {
		return "YouTube blocked this request (HTTP 403). Retry later or try another network."
	}
	if strings.Contains(lower, "no downloadable videos found in playlist") {
		return "No downloadable videos found in this playlist."
	}
	if strings.Contains(lower, "no downloadable videos found in url") {
		return "No downloadable videos found at this URL."
	}
	return msg
}
