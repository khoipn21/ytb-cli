package downloader

import "testing"

func TestDetectTargetType(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   TargetType
	}{
		{
			name:   "channel handle",
			rawURL: "https://www.youtube.com/@GoogleDevelopers/videos",
			want:   TargetTypeChannel,
		},
		{
			name:   "channel id path",
			rawURL: "https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw",
			want:   TargetTypeChannel,
		},
		{
			name:   "watch video",
			rawURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:   TargetTypeVideo,
		},
		{
			name:   "short link video",
			rawURL: "https://youtu.be/dQw4w9WgXcQ",
			want:   TargetTypeVideo,
		},
		{
			name:   "playlist",
			rawURL: "https://www.youtube.com/playlist?list=PL590L5WQmH8fJ54F7z8sE5cUuR9m2y5qv",
			want:   TargetTypePlaylist,
		},
		{
			name:   "watch with playlist context",
			rawURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PL590L5WQmH8fJ54F7z8sE5cUuR9m2y5qv",
			want:   TargetTypePlaylist,
		},
		{
			name:   "not youtube",
			rawURL: "https://example.com/watch?v=dQw4w9WgXcQ",
			want:   TargetTypeUnknown,
		},
		{
			name:   "invalid",
			rawURL: "::",
			want:   TargetTypeUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectTargetType(tc.rawURL)
			if got != tc.want {
				t.Fatalf("DetectTargetType(%q) = %q, want %q", tc.rawURL, got, tc.want)
			}
		})
	}
}

func TestExtractPlaylistVideoIDs_DeduplicatesInOrder(t *testing.T) {
	body := `{"videoId":"aaaaabbbbb1"} ... {"videoId":"cccccddddd2"} ... {"videoId":"aaaaabbbbb1"}`
	got := extractPlaylistVideoIDs(body)
	if len(got) != 2 {
		t.Fatalf("expected 2 IDs, got %d: %v", len(got), got)
	}
	if got[0] != "aaaaabbbbb1" || got[1] != "cccccddddd2" {
		t.Fatalf("unexpected IDs order/content: %v", got)
	}
}

func TestCompactExternalError_TrimsStackTrace(t *testing.T) {
	raw := "JSON parsing error: invalid video duration:\ngoroutine 123 [running]:\nruntime/debug.Stack()"
	got := compactExternalError(assertErr(raw))
	want := "JSON parsing error: invalid video duration:"
	if got != want {
		t.Fatalf("compactExternalError mismatch: got %q want %q", got, want)
	}
}

func TestUserFacingDownloadError_429(t *testing.T) {
	raw := "load video metadata x: unexpected status code: 429"
	got := userFacingDownloadError(assertErr(raw))
	if got == raw {
		t.Fatalf("expected friendly message for 429, got raw message: %q", got)
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }

func assertErr(msg string) error { return testErr(msg) }
