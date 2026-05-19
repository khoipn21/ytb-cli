# YouTube Downloader TUI (Go)

Interactive CLI/TUI app for downloading YouTube media from either:
- a channel URL (all videos)
- a single video URL

## Requirements

- Go 1.23+
- `yt-dlp` in PATH
- `ffmpeg` in PATH

## Usage

```bash
go run .
```

Flags:

- `-url`: Prefill the URL field (channel or single video URL)
- `-mode`: Prefill mode (`audio` or `video`)
- `-output`: Download output directory (default `./downloads`)
- `-yt-dlp-bin`: yt-dlp executable path (default `yt-dlp`)

## Behavior

- Setup screen:
  - select mode: audio/video
  - enter channel or single video URL
  - set output directory
- Download screen:
  - overall progress bar
  - per-video progress rows
  - speed and ETA when available
- Audio mode outputs MP3.
- Video mode outputs MP4 (merged best video/audio).

Controls:
- `Tab` / `Shift+Tab`: move focus in setup
- `Left` / `Right`: toggle mode
- `Enter`: next/start
- `q`: quit
- `esc` on download screen: cancel and return to setup
