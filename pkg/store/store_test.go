package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/jiotv-go/jiotv_go/v3/internal/config"
)

// mockFileInfo is a helper struct to mock os.FileInfo
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (mfi mockFileInfo) Name() string       { return mfi.name }
func (mfi mockFileInfo) Size() int64        { return mfi.size }
func (mfi mockFileInfo) Mode() os.FileMode  { return mfi.mode }
func (mfi mockFileInfo) ModTime() time.Time { return mfi.modTime }
func (mfi mockFileInfo) IsDir() bool        { return mfi.isDir }
func (mfi mockFileInfo) Sys() interface{}   { return nil }

// testTracker is a global map to help track mock calls within specific subtests.
var testTracker map[string]*struct {
	mkdirCalled *bool
	mkdirPath   string
}

func TestGetPathPrefix(t *testing.T) {
	originalOsUserHomeDir := osUserHomeDir
	originalOsStat := osStat
	originalOsMkdir := osMkdir
	originalPathPrefixConfig := config.Cfg.PathPrefix

	defer func() {
		osUserHomeDir = originalOsUserHomeDir
		osStat = originalOsStat
		osMkdir = originalOsMkdir
		config.Cfg.PathPrefix = originalPathPrefixConfig
	}()

	tests := []struct {
		name            string
		setupMocks      func()
		pathPrefixCfg   string
		expectedPath    string
		expectPanic     bool
		mkdirCalled     *bool // Pointer to check if mkdir was called
		mkdirPath       string // Expected path for mkdir
	}{
		{
			name: "Default path, .jiotv_go does not exist",
			setupMocks: func() {
				osUserHomeDir = func() (string, error) { return "/home/user", nil }
				osStat = func(path string) (os.FileInfo, error) {
					if path == filepath.Join("/home/user", PATH_PREFIX) {
						return nil, os.ErrNotExist
					}
					return mockFileInfo{name: filepath.Base(path), isDir: true}, nil
				}
				osMkdir = func(path string, perm os.FileMode) error {
					if tt, ok := testTracker["Default path, .jiotv_go does not exist"]; ok {
						if path == tt.mkdirPath {
							*tt.mkdirCalled = true
						}
					}
					return nil
				}
			},
			expectedPath:    filepath.Join("/home/user", PATH_PREFIX) + string(filepath.Separator),
			expectPanic:     false,
			mkdirPath:       filepath.Join("/home/user", PATH_PREFIX),
		},
		{
			name: "Default path, .jiotv_go exists",
			setupMocks: func() {
				osUserHomeDir = func() (string, error) { return "/home/user", nil }
				osStat = func(path string) (os.FileInfo, error) {
					return mockFileInfo{name: filepath.Base(path), isDir: true}, nil
				}
				osMkdir = func(path string, perm os.FileMode) error {
					if tt, ok := testTracker["Default path, .jiotv_go exists"]; ok {
						*tt.mkdirCalled = true 
					}
					return nil
				}
			},
			expectedPath:    filepath.Join("/home/user", PATH_PREFIX) + string(filepath.Separator),
			expectPanic:     false,
		},
		{
			name: "Config.Cfg.PathPrefix is set",
			setupMocks: func() {
				osStat = func(path string) (os.FileInfo, error) { return mockFileInfo{name: filepath.Base(path), isDir: true}, nil } 
				osMkdir = func(path string, perm os.FileMode) error { return nil }
			},
			pathPrefixCfg:   "/custom/path",
			expectedPath:    "/custom/path" + string(filepath.Separator),
			expectPanic:     false,
		},
		{
			name: "Config.Cfg.PathPrefix needs trailing slash",
			setupMocks: func() {
				osStat = func(path string) (os.FileInfo, error) { return mockFileInfo{name: filepath.Base(path), isDir: true}, nil }
				osMkdir = func(path string, perm os.FileMode) error { return nil }
			},
			pathPrefixCfg:   "/custom/path/no_slash",
			expectedPath:    "/custom/path/no_slash" + string(filepath.Separator),
			expectPanic:     false,
		},
		{
			name: "Error getting user home directory",
			setupMocks: func() {
				osUserHomeDir = func() (string, error) { return "", errors.New("home dir error") }
			},
			expectPanic:     true,
		},
		{
			name: "Error creating directory",
			setupMocks: func() {
				osUserHomeDir = func() (string, error) { return "/home/user", nil }
				osStat = func(path string) (os.FileInfo, error) { 
					if path == filepath.Join("/home/user", PATH_PREFIX) {
						return nil, os.ErrNotExist
					}
					return mockFileInfo{name: filepath.Base(path), isDir: true}, nil
				}
				osMkdir = func(path string, perm os.FileMode) error { return errors.New("mkdir error") }
			},
			expectPanic:     true,
			mkdirPath:       filepath.Join("/home/user", PATH_PREFIX),
		},
	}

	testTracker = make(map[string]*struct{ mkdirCalled *bool; mkdirPath string })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Cfg.PathPrefix = tt.pathPrefixCfg
			
			var mkdirCalled bool
			if tt.mkdirPath != "" { 
				testTracker[tt.name] = &struct{ mkdirCalled *bool; mkdirPath string }{&mkdirCalled, tt.mkdirPath}
			}
			
			tt.setupMocks()

			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("GetPathPrefix() panic = %v, wantPanic %v", r, tt.expectPanic)
					}
				} else if tt.expectPanic {
					t.Errorf("GetPathPrefix() expected panic but did not get one")
				}
				delete(testTracker, tt.name) 
			}()

			got := GetPathPrefix() 

			if !tt.expectPanic {
				normalizedGot := strings.TrimRight(got, string(filepath.Separator)) + string(filepath.Separator)
				normalizedWant := strings.TrimRight(tt.expectedPath, string(filepath.Separator)) + string(filepath.Separator)
				if normalizedGot != normalizedWant {
					t.Errorf("GetPathPrefix() = %q, want %q", normalizedGot, normalizedWant)
				}
				if tt.mkdirPath != "" { 
					if tt.name == "Default path, .jiotv_go does not exist" && !mkdirCalled {
						t.Errorf("osMkdir was not called when .jiotv_go directory did not exist")
					}
					if tt.name == "Default path, .jiotv_go exists" && mkdirCalled {
						t.Errorf("osMkdir was called when .jiotv_go directory already existed")
					}
				}
			}
		})
	}
}


func TestInit(t *testing.T) {
	originalGetPathPrefix := getPathPrefix
	originalOsStat := osStat
	originalTomlDecodeFile := tomlDecodeFile
	originalSaveConfig := saveConfig 

	var saveConfigCalled bool
	var saveConfigData Config
	
	mockedSaveConfig := func() error {
		saveConfigCalled = true
		if KVS != nil { 
			saveConfigData = KVS.config 
		}
		return nil
	}

	defer func() {
		getPathPrefix = originalGetPathPrefix
		osStat = originalOsStat
		tomlDecodeFile = originalTomlDecodeFile
		saveConfig = originalSaveConfig 
		KVS = nil 
	}()

	tempDir, err := os.MkdirTemp("", "store_test_init_global")
	if err != nil {
		t.Fatalf("Failed to create temp dir for TestInit: %v", err)
	}
	defer os.RemoveAll(tempDir) 

	tests := []struct {
		name             string
		setupMocks       func(testFileDir string) 
		wantErr          bool
		expectSaveCalled bool
		expectedData     map[string]string
	}{
		{
			name: "File does not exist - create new",
			setupMocks: func(testFileDir string) {
				getPathPrefix = func() string { return testFileDir + string(filepath.Separator) }
				osStat = func(name string) (os.FileInfo, error) {
					if strings.HasSuffix(name, "store_v4.toml") {
						return nil, os.ErrNotExist 
					}
					return mockFileInfo{name: filepath.Base(name), isDir: true}, nil 
				}
				saveConfig = mockedSaveConfig
			},
			wantErr:          false,
			expectSaveCalled: true,
			expectedData:     map[string]string{},
		},
		{
			name: "File exists and is valid TOML",
			setupMocks: func(testFileDir string) {
				getPathPrefix = func() string { return testFileDir + string(filepath.Separator) }
				dummyTomlFile := filepath.Join(testFileDir, "store_v4.toml")
				content := `[data]
key1 = "value1"
key2 = "value2"
`
				if err := os.WriteFile(dummyTomlFile, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write dummy toml: %v", err)
				}
				osStat = func(name string) (os.FileInfo, error) {
					if name == dummyTomlFile {
						return mockFileInfo{name: "store_v4.toml"}, nil
					}
					return mockFileInfo{name: filepath.Base(name), isDir: true}, nil
				}
				tomlDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
					if fpath != dummyTomlFile { 
						return toml.MetaData{}, fmt.Errorf("tomlDecodeFile called with wrong path: %s, expected %s", fpath, dummyTomlFile)
					}
					return toml.DecodeFile(fpath, v)
				}
				saveConfig = func() error { saveConfigCalled = true; t.Error("saveConfig should not be called when file exists and is valid"); return nil }
			},
			wantErr:          false,
			expectSaveCalled: false,
			expectedData:     map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "File exists but is malformed TOML",
			setupMocks: func(testFileDir string) {
				getPathPrefix = func() string { return testFileDir + string(filepath.Separator)}
				dummyTomlFile := filepath.Join(testFileDir, "store_v4.toml")
				content := `data = { key1 = "value1",, }` // Malformed TOML
				if err := os.WriteFile(dummyTomlFile, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write malformed dummy toml: %v", err)
				}
				osStat = func(name string) (os.FileInfo, error) {
					return mockFileInfo{name: "store_v4.toml"}, nil
				}
				tomlDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
					return toml.DecodeFile(fpath, v) 
				}
				saveConfig = func() error { saveConfigCalled = true; t.Error("saveConfig should not be called on malformed TOML"); return nil }
			},
			wantErr:          true, 
			expectSaveCalled: false,
		},
		{
			name: "Path is a directory, not a file",
			setupMocks: func(testFileDir string) {
				getPathPrefix = func() string { return testFileDir + string(filepath.Separator) }
				storeV4PathAsDir := filepath.Join(testFileDir, "store_v4.toml")
				if err := os.Mkdir(storeV4PathAsDir, 0755); err != nil {
					t.Fatalf("Failed to create dummy dir for store_v4.toml: %v", err)
				}
				osStat = func(name string) (os.FileInfo, error) {
					if name == storeV4PathAsDir { 
						return mockFileInfo{name: "store_v4.toml", isDir: true}, nil
					}
					return mockFileInfo{name: filepath.Base(name), isDir: true}, nil
				}
				tomlDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
					return toml.DecodeFile(fpath, v)
				}
				saveConfig = func() error { saveConfigCalled = true; t.Error("saveConfig should not be called if path is dir"); return nil }
			},
			wantErr:          true, 
			expectSaveCalled: false,
		},
		{
			name: "saveConfig fails during initial creation",
			setupMocks: func(testFileDir string) {
				getPathPrefix = func() string { return testFileDir + string(filepath.Separator) }
				osStat = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
				saveConfig = func() error { saveConfigCalled = true; return errors.New("failed to save config") }
			},
			wantErr:          true, 
			expectSaveCalled: true, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			KVS = nil 
			saveConfigCalled = false
			saveConfigData = Config{}

			subTestFileDir := filepath.Join(tempDir, strings.ReplaceAll(tt.name, " ", "_"))
			if err := os.MkdirAll(subTestFileDir, 0755); err != nil { 
				t.Fatalf("Failed to create sub test file dir: %v", err)
			}
			
			tt.setupMocks(subTestFileDir)

			err := Init()
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expectSaveCalled != saveConfigCalled {
				t.Errorf("saveConfig call expectation failed: got %v, want %v", saveConfigCalled, tt.expectSaveCalled)
			}

			if !tt.wantErr && KVS != nil { 
				if !reflect.DeepEqual(KVS.config.Data, tt.expectedData) {
					t.Errorf("Init() data = %v, want %v", KVS.config.Data, tt.expectedData)
				}
				if tt.expectSaveCalled && err == nil && KVS != nil && !reflect.DeepEqual(saveConfigData.Data, tt.expectedData) {
					t.Errorf("saveConfig saved data = %v, want %v", saveConfigData.Data, tt.expectedData)
				}
			}
		})
	}
}


func TestGet(t *testing.T) {
	originalGetPathPrefix := getPathPrefix
	originalOsStat := osStat
	originalTomlDecodeFile := tomlDecodeFile
	originalSaveConfig := saveConfig
	originalKVS := KVS

	tempDir, err := os.MkdirTemp("", "store_test_get")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	getPathPrefix = func() string { return tempDir + string(filepath.Separator) }
	osStat = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist } 
	saveConfig = func() error { return nil } 

	defer func() {
		getPathPrefix = originalGetPathPrefix
		osStat = originalOsStat
		tomlDecodeFile = originalTomlDecodeFile
		saveConfig = originalSaveConfig
		KVS = originalKVS 
	}()

	if err := Init(); err != nil { 
		t.Fatalf("Initial Init() failed: %v", err)
	}

	KVS.mu.Lock()
	if KVS.config.Data == nil { 
		KVS.config.Data = make(map[string]string)
	}
	KVS.config.Data["existingKey"] = "existingValue"
	KVS.config.Data["anotherKey"] = "anotherValue"
	KVS.mu.Unlock()


	tests := []struct {
		name    string
		key     string
		want    string
		wantErr error 
	}{
		{
			name:    "Get existing key",
			key:     "existingKey",
			want:    "existingValue",
			wantErr: nil,
		},
		{
			name:    "Get another existing key",
			key:     "anotherKey",
			want:    "anotherValue",
			wantErr: nil,
		},
		{
			name:    "Get non-existent key",
			key:     "nonExistentKey",
			want:    "",
			wantErr: ErrKeyNotFound,
		},
		{
			name:    "Get empty key string", 
			key:     "",
			want:    "",
			wantErr: ErrKeyNotFound, 
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if KVS == nil {
				t.Fatalf("KVS is nil before Get, Init was expected to be called.")
			}

			got, err := Get(tt.key)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) { 
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Get() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	originalGetPathPrefix := getPathPrefix
	originalOsStat := osStat
	originalSaveConfig := saveConfig
	originalKVS := KVS 

	tempDir, _ := os.MkdirTemp("", "store_test_set")
	defer os.RemoveAll(tempDir)

	getPathPrefix = func() string { return tempDir + string(filepath.Separator) }
	osStat = func(name string) (os.FileInfo, error) { 
		if name == tempDir || name == tempDir+string(filepath.Separator) {
			return mockFileInfo{name: filepath.Base(name), isDir: true}, nil
		}
		return nil, os.ErrNotExist 
	}


	var saveCalled bool
	var lastSavedData Config
	saveConfig = func() error {
		saveCalled = true
		if KVS != nil { 
			lastSavedData = KVS.config
		}
		return nil
	}

	defer func() {
		getPathPrefix = originalGetPathPrefix
		osStat = originalOsStat
		saveConfig = originalSaveConfig
		KVS = originalKVS 
	}()

	tests := []struct {
		name        string
		key         string
		value       string
		setupKVS    func() 
		wantErr     bool
		expectedMap map[string]string
		saveShouldFail bool
	}{
		{
			name:  "Set new key",
			key:   "newKey",
			value: "newValue",
			setupKVS: func() {
				KVS = nil 
				Init()    
			},
			wantErr:     false,
			expectedMap: map[string]string{"newKey": "newValue"},
		},
		{
			name:  "Update existing key",
			key:   "existingKey",
			value: "updatedValue",
			setupKVS: func() {
				Init() 
				KVS.mu.Lock()
				KVS.config.Data["existingKey"] = "initialValue"
				KVS.mu.Unlock()
			},
			wantErr:     false,
			expectedMap: map[string]string{"existingKey": "updatedValue"},
		},
		{
			name:  "Set on nil KVS (should auto-init)",
			key:   "autoInitKey",
			value: "autoInitValue",
			setupKVS: func() {
				KVS = nil 
			},
			wantErr:     false,
			expectedMap: map[string]string{"autoInitKey": "autoInitValue"},
		},
		{
			name:  "Set causes saveConfig error",
			key:   "saveFailKey",
			value: "saveFailValue",
			setupKVS: func() {
				Init()
			},
			wantErr:     true, 
			expectedMap: map[string]string{"saveFailKey": "saveFailValue"}, 
			saveShouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveCalled = false 
			lastSavedData = Config{}
			
			if tt.setupKVS != nil {
				tt.setupKVS()
			}

			if tt.saveShouldFail {
				saveConfig = func() error { 
					saveCalled = true
					return errors.New("simulated save config error")
				}
			} else {
				saveConfig = func() error { 
					saveCalled = true
					if KVS != nil {
						lastSavedData = KVS.config
					}
					return nil
				}
			}


			err := Set(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if KVS == nil || KVS.config.Data == nil {
					t.Fatalf("KVS or KVS.config.Data is nil after Set call")
				}
				if val, ok := KVS.config.Data[tt.key]; !ok || val != tt.value {
					t.Errorf("Set() failed to set key %q to %q in map, got %q", tt.key, tt.value, val)
				}
				if !saveCalled {
					t.Error("Set() did not call saveConfig")
				}
				if !tt.saveShouldFail && !reflect.DeepEqual(lastSavedData.Data, tt.expectedMap) {
					t.Errorf("saveConfig was called with data = %v, want %v", lastSavedData.Data, tt.expectedMap)
				}
			} else if err != nil && tt.saveShouldFail && !saveCalled {
				t.Error("Set() wanted saveConfig error, but saveConfig was not called")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	originalGetPathPrefix := getPathPrefix
	originalOsStat := osStat
	originalSaveConfig := saveConfig
	originalKVS := KVS

	tempDir, _ := os.MkdirTemp("", "store_test_delete")
	defer os.RemoveAll(tempDir)

	getPathPrefix = func() string { return tempDir + string(filepath.Separator) }
	osStat = func(name string) (os.FileInfo, error) { 
		if name == tempDir || name == tempDir+string(filepath.Separator) {
			return mockFileInfo{name: filepath.Base(name), isDir: true}, nil
		}
		return nil, os.ErrNotExist 
	}

	var saveCalled bool
	var lastSavedConfig Config
	saveConfig = func() error {
		saveCalled = true
		if KVS != nil {
			lastSavedConfig = KVS.config
		}
		return nil
	}

	defer func() {
		getPathPrefix = originalGetPathPrefix
		osStat = originalOsStat
		saveConfig = originalSaveConfig
		KVS = originalKVS
	}()


	tests := []struct {
		name        string
		keyToDelete string
		initialData map[string]string
		wantErr     bool
		keyShouldExistAfterDelete bool
		expectedDataAfterDelete map[string]string 
		saveShouldFail bool
	}{
		{
			name:        "Delete existing key",
			keyToDelete: "key1",
			initialData: map[string]string{"key1": "value1", "key2": "value2"},
			wantErr:     false,
			keyShouldExistAfterDelete: false,
			expectedDataAfterDelete: map[string]string{"key2": "value2"},
		},
		{
			name:        "Delete non-existent key",
			keyToDelete: "nonExistent",
			initialData: map[string]string{"key1": "value1"},
			wantErr:     false, 
			keyShouldExistAfterDelete: false, 
			expectedDataAfterDelete: map[string]string{"key1": "value1"},
		},
		{
			name:        "Delete from nil KVS (should auto-init to empty then delete non-existent)",
			keyToDelete: "anyKey",
			initialData: nil, 
			wantErr:     false, 
			keyShouldExistAfterDelete: false,
			expectedDataAfterDelete: map[string]string{},
		},
		{
			name:        "Delete causes saveConfig error",
			keyToDelete: "key1",
			initialData: map[string]string{"key1": "value1"},
			wantErr:     true,
			keyShouldExistAfterDelete: false, 
			expectedDataAfterDelete: map[string]string{},
			saveShouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveCalled = false
			lastSavedConfig = Config{}
			
			if tt.initialData == nil { 
				KVS = nil
			} else {
				if err := Init(); err != nil { 
					t.Fatalf("Init() failed for test setup: %v", err)
				}
				KVS.mu.Lock()
				KVS.config.Data = make(map[string]string) 
				for k, v := range tt.initialData {
					KVS.config.Data[k] = v
				}
				KVS.mu.Unlock()
			}

			if tt.saveShouldFail {
				saveConfig = func() error {
					saveCalled = true
					return errors.New("simulated save error for delete")
				}
			} else {
				saveConfig = func() error {
					saveCalled = true
					if KVS != nil { lastSavedConfig = KVS.config }
					return nil
				}
			}


			err := Delete(tt.keyToDelete)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if KVS == nil || KVS.config.Data == nil {
					t.Fatalf("KVS or KVS.config.Data is nil after Delete call")
				}
				if _, ok := KVS.config.Data[tt.keyToDelete]; ok != tt.keyShouldExistAfterDelete {
					if tt.keyShouldExistAfterDelete { 
						t.Errorf("Delete() did not remove key %q as expected", tt.keyToDelete)
					} else { 
						t.Errorf("Delete() key %q still exists in map, but should have been removed", tt.keyToDelete)
					}
				}
				if !reflect.DeepEqual(KVS.config.Data, tt.expectedDataAfterDelete) {
					t.Errorf("KVS.config.Data after Delete() = %v, want %v", KVS.config.Data, tt.expectedDataAfterDelete)
				}

				if !saveCalled {
					t.Error("Delete() did not call saveConfig")
				}
				if !tt.saveShouldFail && !reflect.DeepEqual(lastSavedConfig.Data, tt.expectedDataAfterDelete) {
					t.Errorf("saveConfig was called with data = %v, want %v", lastSavedConfig.Data, tt.expectedDataAfterDelete)
				}
			} else if err != nil && tt.saveShouldFail && !saveCalled {
				t.Error("Delete() wanted saveConfig error, but saveConfig was not called")
			}
		})
	}
}


func Test_saveConfig(t *testing.T) {
	originalOsCreate := osCreate
	originalTomlEncode := tomlEncode
	originalKVS := KVS 

	// Setup KVS for these tests - needs to be initialized here as it's global
	// and Init() itself is not called directly by saveConfig
	KVS = &TomlStore{
		config: Config{Data: make(map[string]string)},
		mu:     sync.Mutex{},
	}


	defer func() {
		osCreate = originalOsCreate
		tomlEncode = originalTomlEncode
		KVS = originalKVS 
	}()

	tests := []struct {
		name            string
		setupKVSData    map[string]string
		mockOsCreate    func(name string) (*os.File, error)
		mockTomlEncode  func(w io.Writer, v interface{}) error
		wantErr         bool
		expectedContent string // Expected content if save is successful
	}{
		{
			name: "Successful save",
			setupKVSData: map[string]string{"key1": "value1", "key2": "value2"},
			mockOsCreate: func(name string) (*os.File, error) {
				// For this test, we want to capture the output.
				// A bytes.Buffer can act as an io.Writer.
				// However, os.Create returns *os.File, which also needs to be an io.Closer.
				// We'll use a real temp file to simplify this and check its content.
				return os.CreateTemp("", "save_config_success_*.toml")
			},
			mockTomlEncode: func(w io.Writer, v interface{}) error { // Use actual toml.Encode
				return toml.NewEncoder(w).Encode(v)
			},
			wantErr:     false,
			expectedContent: "[data]\n  key1 = \"value1\"\n  key2 = \"value2\"\n",
		},
		{
			name: "os.Create fails",
			setupKVSData: map[string]string{"key1": "value1"},
			mockOsCreate: func(name string) (*os.File, error) {
				return nil, errors.New("os.Create failed")
			},
			mockTomlEncode: tomlEncode, // Should not be called
			wantErr:     true,
		},
		{
			name: "toml.Encode fails",
			setupKVSData: map[string]string{"key1": "value1"},
			mockOsCreate: func(name string) (*os.File, error) {
				// This needs to return a valid *os.File for the tomlEncoder to attempt to write to.
				// For simplicity, we'll use a real temp file here too, and the error will come from tomlEncode.
				return os.CreateTemp("", "save_config_toml_fail_*.toml")
			},
			mockTomlEncode: func(w io.Writer, v interface{}) error {
				return errors.New("toml.Encode failed")
			},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			KVS.mu.Lock()
			KVS.config.Data = tt.setupKVSData
			KVS.mu.Unlock()
			
			var tempFileName string // To store the name of the temp file for cleanup and reading
			
			currentOsCreate := tt.mockOsCreate
			if tt.name == "Successful save" || tt.name == "toml.Encode fails" {
				// For these tests, we want osCreate to actually create a temp file.
				actualOsCreate := os.Create 
				currentOsCreate = func(name string) (*os.File, error) {
					f, err := actualOsCreate(name) // Create a real temp file
					if err == nil {
						tempFileName = name 
						KVS.filename = name // Ensure KVS.filename is this temp file for tomlEncode
					}
					return f, err
				}
			} else {
				// For os.Create fails test, KVS.filename needs to be some non-empty string.
				if KVS.filename == "" { KVS.filename = "dummy.toml"}
			}
			osCreate = currentOsCreate
			tomlEncode = tt.mockTomlEncode


			err := saveConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("saveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.name == "Successful save" && tempFileName != "" {
				content, readErr := os.ReadFile(tempFileName)
				if readErr != nil {
					t.Fatalf("Failed to read test file %s: %v", tempFileName, readErr)
				}
				normalizedContent := strings.ReplaceAll(string(content), "\r\n", "\n")
				normalizedExpectedContent := strings.ReplaceAll(tt.expectedContent, "\r\n", "\n")
				if normalizedContent != normalizedExpectedContent {
					t.Errorf("saveConfig() wrote\n%s\nwant\n%s", normalizedContent, normalizedExpectedContent)
				}
			}
			if tempFileName != "" { 
				os.Remove(tempFileName)
			}
			
			if KVS != nil { KVS.filename = "" } // Reset KVS.filename
		})
	}
}


// testTracker is a global map to help track mock calls within specific subtests.
>>>>>>> REPLACE
