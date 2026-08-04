package handlers

import "testing"

func TestHdneaCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		streamURL string
		want      string
	}{
		{
			name:      "live URL uses the bare channel ID",
			channelID: "143",
			streamURL: "https://jiotvbpkmob.cdn.jio.com/bpk-tv/CNBCTV18Prime_MOB/Fallback/index.m3u8",
			want:      "143",
		},
		{
			name:      "catchup URL is namespaced separately",
			channelID: "143",
			streamURL: "https://jiotvcod.cdn.jio.com/bpk-tv/CNBCTV18Prime_MOB/Catchup_Fallback/index.m3u8",
			want:      "143|catchup",
		},
		{
			name:      "matching is case-insensitive",
			channelID: "143",
			streamURL: "https://example.com/CATCHUP_FALLBACK/index.m3u8",
			want:      "143|catchup",
		},
		{
			name:      "empty channel ID stays empty regardless of URL",
			channelID: "",
			streamURL: "https://example.com/catchup/index.m3u8",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hdneaCacheKey(tt.channelID, tt.streamURL); got != tt.want {
				t.Errorf("hdneaCacheKey(%q, %q) = %q, want %q", tt.channelID, tt.streamURL, got, tt.want)
			}
		})
	}

	// A live and a catchup URL for the same channel must never collide, since
	// their tokens carry different ACLs and are not interchangeable.
	liveKey := hdneaCacheKey("143", "https://example.com/Fallback/index.m3u8")
	catchupKey := hdneaCacheKey("143", "https://example.com/Catchup_Fallback/index.m3u8")
	if liveKey == catchupKey {
		t.Errorf("live key (%q) and catchup key (%q) must not be equal", liveKey, catchupKey)
	}
}

func TestMediaURIExtension(t *testing.T) {
	tests := []struct {
		name  string
		match string
		want  string
	}{
		{"plain m3u8", "index.m3u8", ".m3u8"},
		{"plain ts", "segment_001.ts", ".ts"},
		{"plain aac", "audio.aac", ".aac"},
		{
			name:  "catchup variant with a query string is still recognised",
			match: "CNBCTV18Prime_MOB-audio_33635_hin=33600-video=148000.m3u8?vbegin=1785002400&vend=1785004199",
			want:  ".m3u8",
		},
		{
			name:  "catchup segment with a query string is still recognised",
			match: "segment_001.ts?vbegin=1785002400&vend=1785004199",
			want:  ".ts",
		},
		{"unrelated extension", "poster.jpg", ""},
		{"empty input", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mediaURIExtension([]byte(tt.match)); got != tt.want {
				t.Errorf("mediaURIExtension(%q) = %q, want %q", tt.match, got, tt.want)
			}
		})
	}
}
