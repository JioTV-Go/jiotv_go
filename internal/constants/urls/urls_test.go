package urls

import (
	"strings"
	"testing"
)

func TestDomainConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"JioTVAPIDomain", JioTVAPIDomain, "jiotvapi.media.jio.com"},
		{"TVMediaDomain", TVMediaDomain, "tv.media.jio.com"},
		{"JioTVCDNDomain", JioTVCDNDomain, "jiotvapi.cdn.jio.com"},
		{"AuthMediaDomain", AuthMediaDomain, "auth.media.jio.com"},
		{"APIJioDomain", APIJioDomain, "api.jio.com"},
		{"JioTVDataCDNDomain", JioTVDataCDNDomain, "jiotv.data.cdn.jio.com"},
		{"JioTVCatchupCDNDomain", JioTVCatchupCDNDomain, "jiotv.catchup.cdn.jio.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestURLConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		prefix   string
	}{
		{"RefreshTokenURL", RefreshTokenURL, "https://"},
		{"RefreshSSOTokenURL", RefreshSSOTokenURL, "https://"},
		{"ChannelsAPIURL", ChannelsAPIURL, "https://"},
		{"ChannelURL", ChannelURL, "https://"},
		{"EPGURL", EPGURL, "https://"},
		{"EPGPosterURL", EPGPosterURL, "https://"},
		{"EPGPosterURLSlash", EPGPosterURLSlash, "https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.constant, tt.prefix) {
				t.Errorf("%s = %q, should start with %q", tt.name, tt.constant, tt.prefix)
			}
			if tt.constant == "" {
				t.Errorf("%s constant is empty", tt.name)
			}
		})
	}
}

func TestPathConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		prefix   string
	}{
		{"PlaybackAPIPath", PlaybackAPIPath, "/"},
		{"ChannelURLPath", ChannelURLPath, "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.constant, tt.prefix) {
				t.Errorf("%s = %q, should start with %q", tt.name, tt.constant, tt.prefix)
			}
			if tt.constant == "" {
				t.Errorf("%s constant is empty", tt.name)
			}
		})
	}
}