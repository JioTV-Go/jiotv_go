package epg

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

var mockEPGServer *MockEPGServer

// Setup function to initialize dependencies for tests
func setupTest() {
	// Initialize the Log variable to prevent nil pointer dereference
	if utils.Log == nil {
		utils.Log = log.New(os.Stdout, "", log.LstdFlags)
	}
}

// setupTestWithMockServer initializes test environment with HTTP mocking
func setupTestWithMockServer() *MockEPGServer {
	setupTest()
	if mockEPGServer == nil {
		mockEPGServer = NewMockEPGServer()
	}
	return mockEPGServer
}

// teardownTestWithMockServer cleans up the mock server
func teardownTestWithMockServer() {
	if mockEPGServer != nil {
		mockEPGServer.Close()
		mockEPGServer = nil
	}
}

func TestInit(t *testing.T) {
	mockServer := setupTestWithMockServer()
	defer teardownTestWithMockServer()
	
	tests := []struct {
		name string
	}{
		{
			name: "Initialize EPG with mock server",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the EPG initialization without external API calls
			InitWithMockServer(mockServer)
			
			// Check if test EPG file was created
			epgFile := utils.GetPathPrefix() + "epg_test.xml.gz"
			if !utils.FileExists(epgFile) {
				t.Errorf("EPG file was not created: %s", epgFile)
			} else {
				// Clean up test file
				os.Remove(epgFile)
			}
		})
	}
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
	mockServer := setupTestWithMockServer()
	defer teardownTestWithMockServer()
	
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Generate XML with mock server",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := genXMLWithMockServer(mockServer)
			if (err != nil) != tt.wantErr {
				t.Errorf("genXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("genXML() returned empty result")
			}
			// Check if the result contains expected XML structure
			if !tt.wantErr && len(got) > 0 {
				xmlString := string(got)
				if !containsString(xmlString, "channel") {
					t.Errorf("genXML() result should contain channel elements")
				}
				if !containsString(xmlString, "programme") {
					t.Errorf("genXML() result should contain programme elements")
				}
			}
		})
	}
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
	mockServer := setupTestWithMockServer()
	defer teardownTestWithMockServer()
	
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "Generate gzipped XML with mock server",
			filename: "/tmp/test_epg.xml.gz",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing test file
			os.Remove(tt.filename)
			
			err := GenXMLGzWithMockServer(tt.filename, mockServer)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenXMLGz() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Check if file was created
				if !utils.FileExists(tt.filename) {
					t.Errorf("GenXMLGz() should create file %s", tt.filename)
				} else {
					// Clean up test file
					os.Remove(tt.filename)
				}
			}
		})
	}
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
