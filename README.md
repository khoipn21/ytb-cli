# YouTube Downloader TUI (Go)

Interactive terminal app for downloading YouTube media from:
- a channel URL (all uploads)
- a playlist URL
- a single video URL

## Overview

This project is a pure Go downloader with an interactive Bubble Tea TUI. It supports audio, video, and combined download modes, live progress reporting, status filtering, per-item details, and queueing additional URLs while a session is running.

## Requirements

- Go 1.26+
- Network access to YouTube

## Run

```bash
go run .
```

You can also build a binary:

```bash
go build -o yt-downloader .
./yt-downloader
```

## CLI Options

- `-url`: prefill YouTube channel or video URL
- `-mode`: prefill mode (`audio`, `video`, `both`)
- `-output`: output directory (default `./downloads`)
- `-help`: show CLI help

Example:

```bash
go run . -url "https://www.youtube.com/@SomeChannel/videos" -mode both -output "./downloads"
```

## TUI Flow

1. Setup screen
- select mode (`audio` / `video` / `both`)
- enter a channel, playlist, or video URL
- see the detected link type before starting
- choose output directory
- start download
- the setup logo renders as animated multi-color `YTBCLI` text pixel art with an entrance reveal and scattered idle shimmer

2. Download screen
- overall progress bar
- responsive table with per-item status
- top status filters (`All`, `Active`, `Completed`, `Failed`, `Queued`)
- detail panel for the selected row
- speed and ETA when available
- add more URLs while running and queue them sequentially

## Keybindings

Setup screen:
- `Tab` / `Shift+Tab`: move focus
- `Up` / `Down` / `j` / `k`: move focus
- `Left` / `Right` / `h` / `l`: change mode when mode row is focused
- `Enter`: next field or start
- `?`: toggle setup help
- `q`: quit

Download screen:
- `a`: open add-link composer
- `h` / `l`: switch status filter (top bar)
- `Enter` (main table): open detail for selected row
- `Esc` / `Enter` (detail panel): close detail
- `Left` / `Right` (in composer): switch mode for added URL
- `Enter` (in composer): queue/start the added URL
- `Esc` (in composer): close composer
- `Esc` (main download view): cancel current run and return to setup
- `q` or `Ctrl+C`: quit
- `Up` / `Down` / `j` / `k`: move selected row
- `?`: toggle download help

## Download Behavior

URL handling:
- single video URLs are detected for `youtu.be`, `/watch?v=`, `/shorts/`, and `/live/`
- playlist URLs are detected for `/playlist` and playlist-context `/watch` URLs
- channel URLs are resolved to channel uploads playlists
- unsupported URLs are marked as `unknown` in the setup UI

Mode behavior:
- `audio`: best audio-only stream
- `video`: best progressive video stream with audio
- `both`: each source video is expanded into two tasks (`[audio]` and `[video]`)

File output:
- filenames are sanitized for cross-platform safety
- output format/extension comes from selected stream MIME type
- files are written into the selected output directory

## Architecture Summary

- App layer: parses flags, resolves output path, launches TUI
- TUI layer: setup screen, download dashboard, event-driven state updates, add-link queue
- Downloader layer: resolves targets, selects media formats, streams bytes to disk, emits progress events

Data flow:
1. User configures request in setup UI.
2. Downloader resolves target videos.
3. TUI receives playlist metadata and initializes rows.
4. Downloader emits start/progress/done/error events per task.
5. TUI updates table + overall progress until finished.
6. Optional queued URLs run sequentially in the same session.

## Current Status

Core phases complete:
- CLI contract and dependencies
- pure-Go channel/single-video download engine
- Bubble Tea setup and progress UI
- validation and baseline docs consolidation

Recent delivered changes:
- migration away from external `yt-dlp` shell execution to pure Go media downloading
- combined `both` mode support
- responsive progress table with richer per-item status and logs
- add-link composer to queue additional URLs during active sessions

## Known Limitations and Planned Improvements

- No resume support for interrupted downloads yet
- No retry/backoff policy yet
- Queue currently executes sequentially only (no parallel mode)
- Integration test coverage can be expanded for end-to-end scenarios
