package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestGetLogger(t *testing.T) {
	originalCfg := config.Cfg
	originalLog := Log // Save original global logger
	// Preserve the original PathPrefix from config.Cfg to restore it.
	// store.GetPathPrefix() uses config.Cfg.PathPrefix.
	originalStorePathPrefix := config.Cfg.PathPrefix


	t.Cleanup(func() {
		config.Cfg = originalCfg
		Log = originalLog
		// Restore the original PathPrefix in the global config.Cfg
		config.Cfg.PathPrefix = originalStorePathPrefix
	})

	// Define a base temporary path for default logs when LogPath is empty
	defaultTestLogBaseDir, err := os.MkdirTemp("", "default_log_base_")
	if err != nil {
		t.Fatalf("Failed to create temp base dir for default logs: %v", err)
	}
	// Cleanup this base directory for default logs after all tests in this function run.
	// Individual tests creating subdirectories under this might also have their own cleanup.
	t.Cleanup(func() {
		os.RemoveAll(defaultTestLogBaseDir)
	})


	testCases := []struct {
		name               string
		setupCfg           func() (effectiveLogDir string, cleanup func()) // Returns directory where jiotv_go.log should be, and a cleanup function
		logToStdout        bool
		debug              bool
		expectedFlags      int
		expectedPrefix     string
		expectFileLog      bool // Whether we expect a lumberjack.Logger to be configured
		expectStdoutLog    bool // Whether we expect os.Stdout to be configured
		checkSpecificPath  string // If non-empty, this exact directory should be created by GetLogger
		expectedLumberjackFilename string // Expected filename for lumberjack
	}{
		{
			name: "Stdout and Default File, No Debug",
			setupCfg: func() (string, func()) {
				// Set PathPrefix for predictable default log location under defaultTestLogBaseDir
				config.Cfg.PathPrefix = filepath.Join(defaultTestLogBaseDir, "apphome1") 
				// GetPathPrefix() in utils.go will use this. Default log is inside ".jiotv_go" under this.
				return filepath.Join(config.Cfg.PathPrefix, ".jiotv_go"), func() { 
					// os.RemoveAll(config.Cfg.PathPrefix) // Clean up the specific apphome used for this test
				}
			},
			logToStdout:     true,
			debug:           false,
			expectedFlags:   log.Ldate | log.Ltime,
			expectedPrefix:  "[INFO] ",
			expectFileLog:   true,
			expectStdoutLog: true,
		},
		{
			name: "File only, Custom Path, No Debug",
			setupCfg: func() (string, func()) {
				customLogDir, _ := os.MkdirTemp("", "custom_log_dir1_")
				config.Cfg.LogPath = customLogDir // GetLogger should use this path
				return customLogDir, func() { os.RemoveAll(customLogDir) }
			},
			logToStdout:     false,
			debug:           false,
			expectedFlags:   log.Ldate | log.Ltime,
			expectedPrefix:  "[INFO] ",
			expectFileLog:   true,
			expectStdoutLog: false,
			checkSpecificPath: "custom_log_dir1_", // Indicate that this path (or similar, from MkdirTemp) should be created
		},
		{
			name: "Stdout and File, Custom Path, Debug Enabled",
			setupCfg: func() (string, func()) {
				customLogDir, _ := os.MkdirTemp("", "custom_log_dir2_")
				config.Cfg.LogPath = customLogDir
				return customLogDir, func() { os.RemoveAll(customLogDir) }
			},
			logToStdout:     true,
			debug:           true,
			expectedFlags:   log.Ldate | log.Ltime | log.Lshortfile,
			expectedPrefix:  "[DEBUG] ",
			expectFileLog:   true,
			expectStdoutLog: true,
			checkSpecificPath: "custom_log_dir2_",
		},
		{
			name: "Default File only, No Stdout, Debug Enabled",
			setupCfg: func() (string, func()) {
				config.Cfg.PathPrefix = filepath.Join(defaultTestLogBaseDir, "apphome2")
				return filepath.Join(config.Cfg.PathPrefix, ".jiotv_go"), func() { 
					// os.RemoveAll(config.Cfg.PathPrefix) 
				}
			},
			logToStdout:     false,
			debug:           true,
			expectedFlags:   log.Ldate | log.Ltime | log.Lshortfile,
			expectedPrefix:  "[DEBUG] ",
			expectFileLog:   true,
			expectStdoutLog: false,
		},
		{
			name: "Directory Creation for Nested Custom LogPath",
			setupCfg: func() (string, func()) {
				customPathParent, _ := os.MkdirTemp("", "custom_log_parent_")
				nestedPath := filepath.Join(customPathParent, "subdir1", "subdir2") // This path doesn't exist yet
				config.Cfg.LogPath = nestedPath                                   // GetLogger should create it
				return nestedPath, func() { os.RemoveAll(customPathParent) }
			},
			logToStdout:     false,
			debug:           false,
			expectedFlags:   log.Ldate | log.Ltime,
			expectedPrefix:  "[INFO] ",
			expectFileLog:   true,
			expectStdoutLog: false,
			checkSpecificPath: "custom_log_parent_", // The base of the nested path for check
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset parts of global config.Cfg that are modified by test cases
			// PathPrefix is reset before setupCfg if setupCfg needs to set it.
			// LogPath, LogToStdout, Debug are set per test case.
			currentOriginalCfg := config.Cfg // Save current state of global Cfg for this subtest
			config.Cfg.LogPath = "" // Reset LogPath before setup
			config.Cfg.PathPrefix = originalStorePathPrefix // Ensure PathPrefix is reset for setupCfg

			effectiveLogDir, testSpecificCleanup := tc.setupCfg()
			if testSpecificCleanup != nil {
				t.Cleanup(testSpecificCleanup)
			}
			
			config.Cfg.LogToStdout = tc.logToStdout
			config.Cfg.Debug = tc.debug
			
			// Store the actual LogPath that was set in config.Cfg by setupCfg for later checks
			configuredLogPath := config.Cfg.LogPath 

			// Reset the global Log variable before calling GetLogger
			// This ensures we are testing GetLogger's effect on the global Log variable as well.
			Log = nil

			logger := GetLogger() // This will also set the global utils.Log

			if logger.Flags() != tc.expectedFlags {
				t.Errorf("Expected flags %d, got %d", tc.expectedFlags, logger.Flags())
			}
			// Prefix check is tricky as it might contain date/time if not reset.
			// The GetLogger implementation sets prefix and flags AFTER creating the logger.
			// So, logger.Prefix() should be exactly what we expect.
			if logger.Prefix() != tc.expectedPrefix {
				t.Errorf("Expected prefix %q, got %q", tc.expectedPrefix, logger.Prefix())
			}

			// Verify directory creation for custom log paths
			if tc.checkSpecificPath != "" { // Indicates a custom path where directory creation should be checked
				// effectiveLogDir is the path that GetLogger is expected to create or ensure exists.
				if _, err := os.Stat(effectiveLogDir); os.IsNotExist(err) {
					t.Errorf("Expected log directory %s to be created, but it was not", effectiveLogDir)
				}
			}
			
			// Attempt to check the lumberjack logger's filename if possible.
			// This is difficult because the writer is wrapped in io.MultiWriter.
			// However, if expectFileLog is true AND expectStdoutLog is false, 
			// the writer *might* be the lumberjack logger directly.
			writer := logger.Writer()
			lumberjackLogger, isLumberjack := writer.(*lumberjack.Logger)

			if tc.expectFileLog {
				if !isLumberjack {
					// If it's not directly a lumberjack logger, it's likely a MultiWriter.
					// We can't easily inspect its components without reflection.
					// However, we can check if the directory was created (done above for custom paths).
					// For default paths, GetLogger also ensures the directory exists.
					if configuredLogPath == "" { // Default path case
						// effectiveLogDir is $PathPrefix/.jiotv_go
						if _, err := os.Stat(effectiveLogDir); os.IsNotExist(err) {
							t.Errorf("Expected default log directory %s to be created, but it was not", effectiveLogDir)
						}
					}
					t.Logf("File logging is expected, but writer is not directly a lumberjack.Logger (likely MultiWriter). Directory creation is checked.")
				} else {
					// It is directly a lumberjack logger
					expectedLogFile := filepath.Join(effectiveLogDir, "jiotv_go.log")
					if lumberjackLogger.Filename != expectedLogFile {
						t.Errorf("Expected lumberjack filename %s, got %s", expectedLogFile, lumberjackLogger.Filename)
					}
				}
			}

			if !tc.expectFileLog && isLumberjack {
				t.Errorf("Did not expect lumberjack.Logger, but found one with file: %s", lumberjackLogger.Filename)
			}

			// Verifying os.Stdout is part of a MultiWriter is hard without output capture or reflection.
			// We rely on the GetLogger's internal logic being correct based on config.Cfg.LogToStdout.
			if tc.expectStdoutLog {
				// This is an implicit check. If GetLogger is correct, os.Stdout was included.
				// If !tc.expectFileLog, then logger.Writer() would be os.Stdout.
				if !tc.expectFileLog && writer != os.Stdout {
					t.Errorf("Expected os.Stdout writer, but got %T", writer)
				}
			} else { // Not expecting stdout
				if writer == os.Stdout {
					t.Errorf("Did not expect os.Stdout writer, but found it directly")
				}
			}
			
			// Ensure the global Log is also set
			if Log != logger {
				t.Error("Global utils.Log was not set by GetLogger, or was not set to the returned logger")
			}

			// Restore config state for the next iteration of the loop (PathPrefix, LogPath)
			// The main t.Cleanup handles the overall originalCfg restoration.
			config.Cfg = currentOriginalCfg
		})
	}
}
