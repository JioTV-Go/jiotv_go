package cmd

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a dummy config file
	dummyConfig := `{"debug": true, "port": "8080"}`
	filePath := "dummy_config.json"
	err := os.WriteFile(filePath, []byte(dummyConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy config file: %v", err)
	}
	defer os.Remove(filePath)

	// Test loading the dummy config
	if err := LoadConfig(filePath); err != nil {
		t.Errorf("LoadConfig() failed: %v", err)
	}

	// Test with a non-existent file
	if err := LoadConfig("non_existent_config.json"); err == nil {
		t.Errorf("LoadConfig() should have failed for non_existent_config.json but did not")
	}
}

func TestInitializeLogger(t *testing.T) {
	// This test mainly checks if InitializeLogger panics.
	// A more thorough test would require checking log output,
	// which might be complex for a unit test without refactoring.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitializeLogger() panicked: %v", r)
		}
	}()
	InitializeLogger()
	// Additionally, check if the logger was initialized
	if Logger() == nil {
		t.Errorf("Logger() returned nil after InitializeLogger()")
	}
}
