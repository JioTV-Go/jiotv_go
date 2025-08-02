package epg

import (
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

var (
	setupOnce sync.Once
)

// Setup function to initialize dependencies for tests
func setupTest() {
	setupOnce.Do(func() {
		// Initialize the Log variable to prevent nil pointer dereference
		if utils.Log == nil {
			utils.Log = log.New(os.Stdout, "", log.LstdFlags)
		}
	})
}

// setupTestWithMockServer initializes test environment with HTTP mocking
// Returns a new mock server instance for each test to prevent test interference
func setupTestWithMockServer() *MockEPGServer {
	setupTest()
	return NewMockEPGServer()
}

// teardownTestWithMockServer cleans up the mock server instance
func teardownTestWithMockServer(mockServer *MockEPGServer) {
	if mockServer != nil {
		mockServer.Close()
	}
}

func TestInit(t *testing.T) {
	// This test requires external API calls which are not suitable for unit testing
	// without changing production code to accept dependency injection.
	// Skipping to avoid external dependencies in unit tests.
	t.Skip("Skipping Init test - requires external API calls")
}

func TestNewProgramme(t *testing.T) {
	type args struct {
		channelID int
		start     string
		stop      string
		title     string
		desc      string
		category  string
		iconSrc   string
	}
	tests := []struct {
		name string
		args args
		want Programme
	}{
		{
			name: "Create new programme",
			args: args{
				channelID: 123,
				start:     "20231225120000 +0530",
				stop:      "20231225130000 +0530",
				title:     "Test Show",
				desc:      "Test Description",
				category:  "Entertainment",
				iconSrc:   "test_icon.jpg",
			},
			want: Programme{
				Channel: "123",
				Start:   "20231225120000 +0530",
				Stop:    "20231225130000 +0530",
				Title: Title{
					Value: "Test Show",
					Lang:  "en",
				},
				Desc: Desc{
					Value: "Test Description",
					Lang:  "en",
				},
				Category: Category{
					Value: "Entertainment",
					Lang:  "en",
				},
				Icon: Icon{
					Src: "https://jiotv.catchup.cdn.jio.com/dare_images/shows/test_icon.jpg",
				},
			},
		},
		{
			name: "Create programme with empty values",
			args: args{
				channelID: 0,
				start:     "",
				stop:      "",
				title:     "",
				desc:      "",
				category:  "",
				iconSrc:   "",
			},
			want: Programme{
				Channel: "0",
				Start:   "",
				Stop:    "",
				Title: Title{
					Value: "",
					Lang:  "en",
				},
				Desc: Desc{
					Value: "",
					Lang:  "en",
				},
				Category: Category{
					Value: "",
					Lang:  "en",
				},
				Icon: Icon{
					Src: "https://jiotv.catchup.cdn.jio.com/dare_images/shows/",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewProgramme(tt.args.channelID, tt.args.start, tt.args.stop, tt.args.title, tt.args.desc, tt.args.category, tt.args.iconSrc)
			if got.Channel != tt.want.Channel {
				t.Errorf("NewProgramme().Channel = %v, want %v", got.Channel, tt.want.Channel)
			}
			if got.Start != tt.want.Start {
				t.Errorf("NewProgramme().Start = %v, want %v", got.Start, tt.want.Start)
			}
			if got.Title.Value != tt.want.Title.Value {
				t.Errorf("NewProgramme().Title.Value = %v, want %v", got.Title.Value, tt.want.Title.Value)
			}
			if got.Icon.Src != tt.want.Icon.Src {
				t.Errorf("NewProgramme().Icon.Src = %v, want %v", got.Icon.Src, tt.want.Icon.Src)
			}
		})
	}
}

func Test_genXML(t *testing.T) {
	// This test requires external API calls which are not suitable for unit testing
	// without changing production code to accept dependency injection.
	// Skipping to avoid external dependencies in unit tests.
	t.Skip("Skipping genXML test - requires external API calls")
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func Test_formatTime(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Format specific time",
			args: args{t: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)},
			want: "20231225153045 +0000",
		},
		{
			name: "Format with different timezone",
			args: args{t: time.Date(2023, 1, 1, 0, 0, 0, 0, time.FixedZone("EST", -5*3600))},
			want: "20230101000000 -0500",
		},
		{
			name: "Format with positive timezone",
			args: args{t: time.Date(2023, 6, 15, 12, 0, 0, 0, time.FixedZone("IST", 5*3600+30*60))},
			want: "20230615120000 +0530",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTime(tt.args.t); got != tt.want {
				t.Errorf("formatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenXMLGz(t *testing.T) {
	// This test requires external API calls which are not suitable for unit testing
	// without changing production code to accept dependency injection.
	// Skipping to avoid external dependencies in unit tests.
	t.Skip("Skipping GenXMLGz test - requires external API calls")
}

func TestEpochString_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    EpochString
		wantErr bool
	}{
		{
			name: "Unmarshal from integer",
			args: args{data: []byte("1609459200123")}, // 13-digit timestamp
			want: EpochString("1609459200"),           // Should be truncated to 10 digits
			wantErr: false,
		},
		{
			name: "Unmarshal from string",
			args: args{data: []byte(`"test_string"`)},
			want: EpochString("test_string"),
			wantErr: false,
		},
		{
			name: "Unmarshal from empty string",
			args: args{data: []byte(`""`)},
			want: EpochString(""),
			wantErr: false,
		},
		{
			name: "Unmarshal invalid JSON",
			args: args{data: []byte("invalid json")},
			want: EpochString(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id EpochString
			err := id.UnmarshalJSON(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("EpochString.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id != tt.want {
				t.Errorf("EpochString.UnmarshalJSON() = %v, want %v", id, tt.want)
			}
		})
	}
}

func TestEpochString_String(t *testing.T) {
	tests := []struct {
		name string
		id   EpochString
		want string
	}{
		{
			name: "String representation of epoch",
			id:   EpochString("1609459200"),
			want: "1609459200",
		},
		{
			name: "String representation of empty epoch",
			id:   EpochString(""),
			want: "",
		},
		{
			name: "String representation of text epoch",
			id:   EpochString("test_string"),
			want: "test_string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.String(); got != tt.want {
				t.Errorf("EpochString.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
