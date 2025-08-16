package television

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

var (
	setupOnce sync.Once
)

// Setup function to initialize store for tests
func setupTest() {
	setupOnce.Do(func() {
		// Setup test environment with temporary pathPrefix
		_, err := store.SetupTestPathPrefix()
		if err != nil {
			panic(fmt.Sprintf("Failed to setup test environment: %v", err))
		}
		// Note: cleanup is handled by the temp directory system cleanup

		// Initialize store for testing
		store.Init()
		// Initialize secureurl for URL encryption/decryption
		secureurl.Init()
		// Initialize the Log variable to prevent nil pointer dereference
		if utils.Log == nil {
			utils.Log = log.New(os.Stdout, "", log.LstdFlags)
		}
	})
}

func TestFilterChannels(t *testing.T) {
	// Create test data
	testChannels := []Channel{
		{ID: "1", Name: "Hindi Entertainment", Language: 1, Category: 5}, // Hindi Entertainment
		{ID: "2", Name: "English Movies", Language: 6, Category: 6},      // English Movies
		{ID: "3", Name: "Hindi Movies", Language: 1, Category: 6},        // Hindi Movies
		{ID: "4", Name: "English Sports", Language: 6, Category: 8},      // English Sports
		{ID: "5", Name: "Tamil Entertainment", Language: 8, Category: 5}, // Tamil Entertainment
	}

	type args struct {
		channels []Channel
		language int
		category int
	}
	tests := []struct {
		name string
		args args
		want []Channel
	}{
		{
			name: "Filter by language only (Hindi)",
			args: args{
				channels: testChannels,
				language: 1, // Hindi
				category: 0, // No category filter
			},
			want: []Channel{
				{ID: "1", Name: "Hindi Entertainment", Language: 1, Category: 5},
				{ID: "3", Name: "Hindi Movies", Language: 1, Category: 6},
			},
		},
		{
			name: "Filter by category only (Movies)",
			args: args{
				channels: testChannels,
				language: 0, // No language filter
				category: 6, // Movies
			},
			want: []Channel{
				{ID: "2", Name: "English Movies", Language: 6, Category: 6},
				{ID: "3", Name: "Hindi Movies", Language: 1, Category: 6},
			},
		},
		{
			name: "Filter by both language and category (English Movies)",
			args: args{
				channels: testChannels,
				language: 6, // English
				category: 6, // Movies
			},
			want: []Channel{
				{ID: "2", Name: "English Movies", Language: 6, Category: 6},
			},
		},
		{
			name: "No filters (return all)",
			args: args{
				channels: testChannels,
				language: 0, // No filter
				category: 0, // No filter
			},
			want: testChannels,
		},
		{
			name: "Empty channels slice",
			args: args{
				channels: []Channel{},
				language: 1,
				category: 5,
			},
			want: []Channel{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterChannels(tt.args.channels, tt.args.language, tt.args.category)
			if len(got) != len(tt.want) {
				t.Errorf("FilterChannels() returned %d channels, want %d", len(got), len(tt.want))
				return
			}
			for i, channel := range got {
				if channel.ID != tt.want[i].ID {
					t.Errorf("FilterChannels() channel[%d].ID = %v, want %v", i, channel.ID, tt.want[i].ID)
				}
			}
		})
	}
}

func TestReplaceM3U8(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		baseUrl    []byte
		match      []byte
		params     string
		channel_id string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace M3U8 URL with parameters",
			args: args{
				baseUrl:    []byte("test.m3u8"),
				match:      []byte("test.m3u8"),
				params:     "param1=value1",
				channel_id: "123",
			},
		},
		{
			name: "Replace M3U8 URL with empty params",
			args: args{
				baseUrl:    []byte("example.m3u8"),
				match:      []byte("example.m3u8"),
				params:     "",
				channel_id: "456",
			},
		},
		{
			name: "No match found",
			args: args{
				baseUrl:    []byte("original content"),
				match:      []byte("not_found.m3u8"),
				params:     "param1=value1",
				channel_id: "123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceM3U8(tt.args.baseUrl, tt.args.match, tt.args.params, tt.args.channel_id)
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceM3U8() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceM3U8() should contain /render path, got %s", string(got))
			}
		})
	}
}

func TestReplaceTS(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		baseUrl []byte
		match   []byte
		params  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace TS URL with parameters",
			args: args{
				baseUrl: []byte("segment.ts"),
				match:   []byte("segment.ts"),
				params:  "param1=value1",
			},
		},
		{
			name: "Replace TS URL with empty params",
			args: args{
				baseUrl: []byte("test.ts"),
				match:   []byte("test.ts"),
				params:  "",
			},
		},
		{
			name: "No match found",
			args: args{
				baseUrl: []byte("original content"),
				match:   []byte("not_found.ts"),
				params:  "param1=value1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceTS(tt.args.baseUrl, tt.args.match, tt.args.params)
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceTS() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceTS() should contain /render path, got %s", string(got))
			}
		})
	}
}

func TestReplaceAAC(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		baseUrl []byte
		match   []byte
		params  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace AAC URL with parameters",
			args: args{
				baseUrl: []byte("audio.aac"),
				match:   []byte("audio.aac"),
				params:  "param1=value1",
			},
		},
		{
			name: "Replace AAC URL with empty params",
			args: args{
				baseUrl: []byte("test.aac"),
				match:   []byte("test.aac"),
				params:  "",
			},
		},
		{
			name: "No match found",
			args: args{
				baseUrl: []byte("original content"),
				match:   []byte("not_found.aac"),
				params:  "param1=value1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceAAC(tt.args.baseUrl, tt.args.match, tt.args.params)
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceAAC() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceAAC() should contain /render path, got %s", string(got))
			}
		})
	}
}

func TestReplaceKey(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		match      []byte
		params     string
		channel_id string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace key with parameters",
			args: args{
				match:      []byte("key.bin"),
				params:     "param1=value1",
				channel_id: "123",
			},
		},
		{
			name: "Replace key with empty params",
			args: args{
				match:      []byte("test.key"),
				params:     "",
				channel_id: "456",
			},
		},
		{
			name: "Replace key with empty channel_id",
			args: args{
				match:      []byte("test.key"),
				params:     "param1=value1",
				channel_id: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceKey(tt.args.match, tt.args.params, tt.args.channel_id)
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceKey() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceKey() should contain /render path, got %s", string(got))
			}
		})
	}
}
