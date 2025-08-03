package headers

import "testing"

func TestHeaderConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ContentType", ContentType, "Content-Type"},
		{"Accept", Accept, "Accept"},
		{"UserAgent", UserAgent, "User-Agent"},
		{"ContentTypeJSON", ContentTypeJSON, "application/json"},
		{"UserAgentOkHttp", UserAgentOkHttp, "okhttp/4.2.2"},
		{"DeviceTypePhone", DeviceTypePhone, "phone"},
		{"OSAndroid", OSAndroid, "android"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestHeaderConstantsNotEmpty(t *testing.T) {
	constants := []struct {
		name  string
		value string
	}{
		{"ContentType", ContentType},
		{"Accept", Accept},
		{"AcceptEncoding", AcceptEncoding},
		{"UserAgent", UserAgent},
		{"Authorization", Authorization},
		{"Host", Host},
		{"AccessToken", AccessToken},
		{"DeviceType", DeviceType},
		{"VersionCode", VersionCode},
		{"OS", OS},
		{"XAPIKey", XAPIKey},
		{"ContentTypeJSON", ContentTypeJSON},
		{"ContentTypeJSONCharsetUTF8", ContentTypeJSONCharsetUTF8},
		{"AcceptJSON", AcceptJSON},
		{"AcceptEncodingGzip", AcceptEncodingGzip},
		{"UserAgentOkHttp", UserAgentOkHttp},
		{"UserAgentPlayTV", UserAgentPlayTV},
		{"DeviceTypePhone", DeviceTypePhone},
		{"OSAndroid", OSAndroid},
		{"VersionCode315", VersionCode315},
		{"APIKeyJio", APIKeyJio},
	}

	for _, c := range constants {
		t.Run(c.name+"_not_empty", func(t *testing.T) {
			if c.value == "" {
				t.Errorf("%s constant is empty", c.name)
			}
		})
	}
}