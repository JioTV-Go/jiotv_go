package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestIsTermux(t *testing.T) {
	originalPrefix, present := os.LookupEnv("PREFIX")
	defer func() {
		if present {
			os.Setenv("PREFIX", originalPrefix)
		} else {
			os.Unsetenv("PREFIX")
		}
	}()

	tests := []struct {
		name   string
		prefix string
		setEnv bool
		want   bool
	}{
		{
			name:   "Termux environment (PREFIX is set)",
			prefix: "/data/data/com.termux/files/usr",
			setEnv: true,
			want:   true,
		},
		{
			name:   "Non-Termux environment (PREFIX is empty)",
			prefix: "",
			setEnv: true,
			want:   false,
		},
		{
			name:   "Non-Termux environment (PREFIX is not set)",
			setEnv: false,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv("PREFIX", tt.prefix)
			} else {
				os.Unsetenv("PREFIX")
			}
			if got := isTermux(); got != tt.want {
				t.Errorf("isTermux() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrep(t *testing.T) {
	content := "hello world\nfind me\nanother line"
	tmpfile, err := os.CreateTemp("", "testgrep")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		filename string
		pattern  string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Pattern exists", args{tmpfile.Name(), "find me"}, true, false},
		{"Pattern does not exist", args{tmpfile.Name(), "not here"}, false, false},
		{"File does not exist", args{"no_such_file", "anything"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grep(tt.args.filename, tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("grep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("grep() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_bashrcModifiers tests both addToBashrc and removeFromBashrc.
func TestBashrcModifiers(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testbashrc")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpfile.Name()
	defer os.Remove(filename)
	tmpfile.Close()

	// Test addToBashrc
	lineToAdd := "export TEST_VAR=1"
	if err := addToBashrc(filename, lineToAdd); err != nil {
		t.Fatalf("addToBashrc() failed: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if !strings.Contains(string(content), lineToAdd) {
		t.Errorf("addToBashrc() did not add the line to the file")
	}

	// Add another line to test removal of a specific line
	anotherLine := "export ANOTHER_VAR=2"
	if err := addToBashrc(filename, anotherLine); err != nil {
		t.Fatalf("addToBashrc() failed to add second line: %v", err)
	}

	// Test removeFromBashrc
	if err := removeFromBashrc(filename, lineToAdd); err != nil {
		t.Fatalf("removeFromBashrc() failed: %v", err)
	}

	content, err = os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file after removal: %v", err)
	}
	if strings.Contains(string(content), lineToAdd) {
		t.Errorf("removeFromBashrc() did not remove the line from the file")
	}
	if !strings.Contains(string(content), anotherLine) {
		t.Errorf("removeFromBashrc() removed the wrong line")
	}
}

// TestAutoStart and Test_getConsentFromUser are intentionally left empty
// as they require user interaction or significant mocking of the OS,
// making them unsuitable for simple unit tests. A senior developer would
// likely refactor the production code to be more testable, for example
// by injecting dependencies for filesystem access and user input.
// For now, we will skip adding tests for it.

func TestAutoStart(t *testing.T) {}

func TestGetConsentFromUser(t *testing.T) {}

// The following tests are covered by Test_bashrcModifiers
func TestAddToBashrc(t *testing.T) {}
func TestRemoveFromBashrc(t *testing.T) {}