package downloader

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var invalidFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

func buildFileName(title, id, ext string) string {
	base := invalidFileChars.ReplaceAllString(strings.TrimSpace(title), "_")
	base = strings.Trim(strings.ReplaceAll(base, "  ", " "), ". ")
	if base == "" {
		base = id
	}
	return fmt.Sprintf("%s-%s.%s", base, id, ext)
}

func humanBytes(n int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	v := float64(n)
	i := 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

func estimateETA(downloaded, total int64, speed float64) string {
	if total <= 0 || speed <= 0 || downloaded >= total {
		return "-"
	}
	remaining := float64(total-downloaded) / speed
	return (time.Duration(remaining) * time.Second).Round(time.Second).String()
}

func percent(downloaded, total int64) float64 {
	if total <= 0 {
		return 0
	}
	p := float64(downloaded) / float64(total)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func maxFloat(v, min float64) float64 {
	if v < min {
		return min
	}
	return v
}
