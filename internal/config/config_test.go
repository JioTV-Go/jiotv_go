package config

import (
	"os"
	"reflect"
	"testing"
)

func TestJioTVConfig_Load(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "jiotv_go_test_*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := []byte("epg: true\ndebug: true\ntitle: TestTitle\n")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name    string
		c       *JioTVConfig
		args    struct{ filename string }
		wantErr bool
	}{
		{
			name:    "Load from valid file",
			c:       &JioTVConfig{},
			args:    struct{ filename string }{filename: tmpFile.Name()},
			wantErr: false,
		},
		{
			name:    "Load from non-existent file",
			c:       &JioTVConfig{},
			args:    struct{ filename string }{filename: "nonexistent.yml"},
			wantErr: true,
		},
		{
			name:    "Load from environment variables",
			c:       &JioTVConfig{},
			args:    struct{ filename string }{filename: ""},
			wantErr: false,
		},
	}
	os.Setenv("JIOTV_EPG", "true")
	defer os.Unsetenv("JIOTV_EPG")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Load(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("JioTVConfig.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.name == "Load from valid file" && (tt.c.EPG != true || tt.c.Debug != true || tt.c.Title != "TestTitle") {
				t.Errorf("JioTVConfig.Load() did not load values correctly from file: %+v", tt.c)
			}
			if tt.name == "Load from environment variables" && tt.c.EPG != true {
				t.Errorf("JioTVConfig.Load() did not load EPG from env: %+v", tt.c)
			}
		})
	}
}

func TestJioTVConfig_Get(t *testing.T) {
	// Set the global Cfg for Get to work as intended
	Cfg = JioTVConfig{
		EPG:                 true,
		Debug:               false,
		DisableTSHandler:    true,
		DisableLogout:       false,
		DRM:                 true,
		Title:               "TestTitle",
		DisableURLEncryption: false,
		Proxy:               "http://proxy",
		PathPrefix:          "/tmp/jiotv",
		LogPath:             "/tmp/logs",
		LogToStdout:         true,
	}

	tests := []struct {
		name string
		j    *JioTVConfig
		args struct{ key string }
		want interface{}
	}{
		{
			name: "Get EPG",
			j:    &JioTVConfig{},
			args: struct{ key string }{key: "EPG"},
			want: true,
		},
		{
			name: "Get Debug",
			j:    &JioTVConfig{},
			args: struct{ key string }{key: "Debug"},
			want: false,
		},
		{
			name: "Get Title",
			j:    &JioTVConfig{},
			args: struct{ key string }{key: "Title"},
			want: "TestTitle",
		},
		{
			name: "Get Proxy",
			j:    &JioTVConfig{},
			args: struct{ key string }{key: "Proxy"},
			want: "http://proxy",
		},
		{
			name: "Get invalid key",
			j:    &JioTVConfig{},
			args: struct{ key string }{key: "NonExistent"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.j.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JioTVConfig.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commonFileExists(t *testing.T) {
	// Create a temp file to simulate a config file
	tmpFile, err := os.CreateTemp("", "jiotv_go.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Save current working directory and change to temp dir
	origDir, _ := os.Getwd()
	tmpDir := os.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Rename temp file to match a common config name
	testFile := "jiotv_go.yml"
	os.Rename(tmpFile.Name(), testFile)
	defer os.Remove(testFile)

	tests := []struct {
		name string
		want string
	}{
		{
			name: "File exists",
			want: "jiotv_go.yml",
		},
		{
			name: "File does not exist",
			want: "",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i == 1 {
				os.Remove("jiotv_go.yml")
			}
			got := commonFileExists()
			if got != tt.want {
				t.Errorf("commonFileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJioTVConfig_LoadMultipleSelection(t *testing.T) {
	// Create a temporary config file for testing multiple selection
	tmpFile, err := os.CreateTemp("", "jiotv_go_multi_test_*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	content := []byte(`epg: false
debug: false
title: "Test Multiple Selection"
preferred_categories: [5, 6, 8]
preferred_languages: [1, 6]
`)
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name string
		c    *JioTVConfig
		args struct{ filename string }
		wantCategories []int
		wantLanguages  []int
		wantErr        bool
	}{
		{
			name:           "Load multiple selection config",
			c:              &JioTVConfig{},
			args:           struct{ filename string }{filename: tmpFile.Name()},
			wantCategories: []int{5, 6, 8},
			wantLanguages:  []int{1, 6},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Load(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("JioTVConfig.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !reflect.DeepEqual(tt.c.PreferredCategories, tt.wantCategories) {
				t.Errorf("JioTVConfig.Load() PreferredCategories = %v, want %v", tt.c.PreferredCategories, tt.wantCategories)
			}
			
			if !reflect.DeepEqual(tt.c.PreferredLanguages, tt.wantLanguages) {
				t.Errorf("JioTVConfig.Load() PreferredLanguages = %v, want %v", tt.c.PreferredLanguages, tt.wantLanguages)
			}
			
			if tt.c.Title != "Test Multiple Selection" {
				t.Errorf("JioTVConfig.Load() Title = %v, want %v", tt.c.Title, "Test Multiple Selection")
			}
		})
	}
}
