package downloader

import (
	"regexp"
	"strconv"
	"strings"
)

var defaultProgressPattern = regexp.MustCompile(`\[download\]\s+([0-9]+(?:\.[0-9]+)?)%`)
var speedPattern = regexp.MustCompile(`at\s+([^\s]+)`)
var etaPattern = regexp.MustCompile(`ETA\s+([0-9:]+)`)

func parseProgressLine(line string) progressUpdate {
	matches := defaultProgressPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return progressUpdate{LogLine: strings.TrimSpace(line)}
	}
	percentRaw, _ := strconv.ParseFloat(matches[1], 64)
	percent := percentRaw / 100.0
	speed := ""
	eta := ""
	if speedMatch := speedPattern.FindStringSubmatch(line); len(speedMatch) > 1 {
		speed = strings.TrimSpace(speedMatch[1])
	}
	if etaMatch := etaPattern.FindStringSubmatch(line); len(etaMatch) > 1 {
		eta = strings.TrimSpace(etaMatch[1])
	}
	return progressUpdate{
		Percent: clampPercent(percent),
		Speed:   speed,
		ETA:     eta,
	}
}

func clampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
