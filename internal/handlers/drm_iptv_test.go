package handlers

import (
	"testing"
)

func TestIsDRMChannel(t *testing.T) {
	// Store original EnableDRM value and restore after test
	originalEnableDRM := EnableDRM
	defer func() { EnableDRM = originalEnableDRM }()

	tests := []struct {
		name        string
		channelID   string
		enableDRM   bool
		expectedDRM bool
	}{
		{
			name:        "SONY channel with DRM enabled",
			channelID:   "154", // Sony SAB HD
			enableDRM:   true,
			expectedDRM: true,
		},
		{
			name:        "SONY channel with DRM disabled",
			channelID:   "154", // Sony SAB HD
			enableDRM:   false,
			expectedDRM: false,
		},
		{
			name:        "Non-SONY channel with DRM enabled",
			channelID:   "123",
			enableDRM:   true,
			expectedDRM: false,
		},
		{
			name:        "Non-SONY channel with DRM disabled",
			channelID:   "123",
			enableDRM:   false,
			expectedDRM: false,
		},
		{
			name:        "Another SONY channel (Sony HD)",
			channelID:   "291",
			enableDRM:   true,
			expectedDRM: true,
		},
		{
			name:        "Empty channel ID",
			channelID:   "",
			enableDRM:   true,
			expectedDRM: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableDRM = tt.enableDRM
			result := isDRMChannel(tt.channelID)
			if result != tt.expectedDRM {
				t.Errorf("isDRMChannel(%q) with EnableDRM=%v = %v, want %v", 
					tt.channelID, tt.enableDRM, result, tt.expectedDRM)
			}
		})
	}
}