package store

import "testing"

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Initialize store",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Verify KVS is not nil after init
			if KVS == nil {
				t.Error("Init() should initialize KVS")
			}
		})
	}
}

func TestGet(t *testing.T) {
	// Initialize store first
	if err := Init(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		setup   func() // Setup function to set values before testing
	}{
		{
			name: "Get non-existent key",
			args: args{key: "nonexistent_key"},
			want: "",
			wantErr: true,
		},
		{
			name: "Get existing key",
			args: args{key: "test_key"},
			want: "test_value",
			wantErr: false,
			setup: func() {
				Set("test_key", "test_value")
			},
		},
		{
			name: "Get empty key",
			args: args{key: ""},
			want: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	// Initialize store first
	if err := Init(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set new key-value pair",
			args: args{key: "new_key", value: "new_value"},
			wantErr: false,
		},
		{
			name: "Set existing key with new value",
			args: args{key: "existing_key", value: "updated_value"},
			wantErr: false,
		},
		{
			name: "Set empty key",
			args: args{key: "", value: "some_value"},
			wantErr: false, // Empty key should be allowed in this implementation
		},
		{
			name: "Set empty value",
			args: args{key: "key_with_empty_value", value: ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If set was successful, verify we can get the value back
			if !tt.wantErr {
				got, err := Get(tt.args.key)
				if err != nil {
					t.Errorf("After Set(), Get() error = %v", err)
					return
				}
				if got != tt.args.value {
					t.Errorf("After Set(), Get() = %v, want %v", got, tt.args.value)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	// Initialize store first
	if err := Init(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func() // Setup function to prepare test data
	}{
		{
			name: "Delete existing key",
			args: args{key: "delete_me"},
			wantErr: false,
			setup: func() {
				Set("delete_me", "value_to_delete")
			},
		},
		{
			name: "Delete non-existent key",
			args: args{key: "non_existent"},
			wantErr: false, // Delete should not error even if key doesn't exist
		},
		{
			name: "Delete empty key",
			args: args{key: ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			
			if err := Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Verify key is deleted by trying to get it
			if !tt.wantErr && tt.args.key != "" {
				_, err := Get(tt.args.key)
				if err == nil {
					t.Errorf("After Delete(), key %s should not exist", tt.args.key)
				}
			}
		})
	}
}

func Test_saveConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// No test cases needed - testing internal function
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := saveConfig(); (err != nil) != tt.wantErr {
				t.Errorf("saveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPathPrefix(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Get path prefix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPathPrefix()
			
			// Should return a non-empty string
			if got == "" {
				t.Errorf("GetPathPrefix() returned empty string")
			}
			
			// Should end with a slash
			if got[len(got)-1] != '/' {
				t.Errorf("GetPathPrefix() should end with '/', got %s", got)
			}
			
			// Should be a valid path (just basic check)
			if len(got) < 2 {
				t.Errorf("GetPathPrefix() returned suspiciously short path: %s", got)
			}
		})
	}
}
