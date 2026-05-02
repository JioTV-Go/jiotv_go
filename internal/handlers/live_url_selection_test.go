package handlers

import (
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
)

func TestSelectBestLiveHLSURL(t *testing.T) {
	tests := []struct {
		name     string
		quality  string
		input    *television.LiveURLOutput
		expected string
	}{
		{
			name:    "returns requested quality when available",
			quality: "high",
			input: &television.LiveURLOutput{
				Bitrates: television.Bitrates{High: "https://cdn.example.com/high.m3u8", Auto: "https://cdn.example.com/auto.m3u8"},
			},
			expected: "https://cdn.example.com/high.m3u8",
		},
		{
			name:    "falls back to another bitrate when requested is missing",
			quality: "high",
			input: &television.LiveURLOutput{
				Bitrates: television.Bitrates{Auto: "https://cdn.example.com/auto.m3u8"},
			},
			expected: "https://cdn.example.com/auto.m3u8",
		},
		{
			name:    "falls back to result hls url when bitrates are empty",
			quality: "high",
			input: &television.LiveURLOutput{
				Result: "https://edge.example.com/channel/master.m3u8?token=abc",
			},
			expected: "https://edge.example.com/channel/master.m3u8?token=abc",
		},
		{
			name:    "falls back to mpd result when it is actually hls",
			quality: "high",
			input: &television.LiveURLOutput{
				Result: "https://edge.example.com/channel/manifest.mpd",
				Mpd: television.MPD{
					Result: "https://edge.example.com/channel/master.m3u8",
				},
			},
			expected: "https://edge.example.com/channel/master.m3u8",
		},
		{
			name:    "returns empty when no hls candidate exists",
			quality: "high",
			input: &television.LiveURLOutput{
				Result: "https://edge.example.com/channel/manifest.mpd",
				Mpd: television.MPD{
					Result: "https://edge.example.com/channel/manifest.mpd",
				},
			},
			expected: "",
		},
		{
			name:     "returns empty for nil input",
			quality:  "high",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectBestLiveHLSURL(tt.input, tt.quality)
			if got != tt.expected {
				t.Fatalf("selectBestLiveHLSURL() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestIsLikelyHLSURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{url: "https://cdn.example.com/index.m3u8", expected: true},
		{url: "https://cdn.example.com/index.M3U8", expected: true},
		{url: "https://cdn.example.com/manifest.mpd", expected: false},
		{url: "", expected: false},
	}

	for _, tt := range tests {
		got := isLikelyHLSURL(tt.url)
		if got != tt.expected {
			t.Fatalf("isLikelyHLSURL(%q) = %v, expected %v", tt.url, got, tt.expected)
		}
	}
}

func TestToAbsoluteStreamURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		live     *television.LiveURLOutput
		expect   string
	}{
		{
			name:     "keeps absolute url unchanged",
			inputURL: "https://cdn.example.com/a/master.m3u8",
			live:     nil,
			expect:   "https://cdn.example.com/a/master.m3u8",
		},
		{
			name:     "converts protocol relative url",
			inputURL: "//cdn.example.com/a/master.m3u8",
			live:     nil,
			expect:   "https://cdn.example.com/a/master.m3u8",
		},
		{
			name:     "uses base host from live result",
			inputURL: "/Disney_Channel/index_3_av.m3u8",
			live: &television.LiveURLOutput{
				Mpd: television.MPD{Result: "https://tv.media.example.com/path/manifest.mpd"},
			},
			expect: "https://tv.media.example.com/Disney_Channel/index_3_av.m3u8",
		},
		{
			name:     "falls back to default jio cdn domain",
			inputURL: "/Disney_Channel/index_3_av.m3u8",
			live:     nil,
			expect:   "https://jiotvapi.cdn.jio.com/Disney_Channel/index_3_av.m3u8",
		},
		{
			name:     "normalizes bare relative path",
			inputURL: "Disney_Channel/index_3_av.m3u8",
			live:     nil,
			expect:   "https://jiotvapi.cdn.jio.com/Disney_Channel/index_3_av.m3u8",
		},
		{
			name:     "adds scheme when host is present without scheme",
			inputURL: "jiotvapi.cdn.jio.com/Disney_Channel/index_3_av.m3u8",
			live:     nil,
			expect:   "https://jiotvapi.cdn.jio.com/Disney_Channel/index_3_av.m3u8",
		},
		{
			name:     "repairs malformed hostless https url",
			inputURL: "https:///Disney_Channel/index_3_av.m3u8",
			live:     nil,
			expect:   "https://jiotvapi.cdn.jio.com/https:///Disney_Channel/index_3_av.m3u8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toAbsoluteStreamURL(tt.inputURL, tt.live)
			if got != tt.expect {
				t.Fatalf("toAbsoluteStreamURL() = %q, expected %q", got, tt.expect)
			}
		})
	}
}
