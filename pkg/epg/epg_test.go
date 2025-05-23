package epg

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	// "net/http" // Using fasthttp specific test server
	// "net/http/httptest" // Using fasthttp specific test server
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/pkg/scheduler"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"         // Import for fasthttp status codes
	"github.com/valyala/fasthttp/fasthttptest" // Import for fasthttp test server
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

func TestInit(t *testing.T) {
	oldLog := utils.Log
	utils.Log = utils.GetLogger() 
	defer func() { utils.Log = oldLog }()

	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()

	oldGenXMLGz := GenXMLGz
	var genXMLGzCalled bool
	GenXMLGz = func(filename string) error {
		genXMLGzCalled = true
		return nil
	}
	defer func() { GenXMLGz = oldGenXMLGz }()

	oldSchedulerAdd := scheduler.Add
	var schedulerAddCalled bool
	var scheduledID string
	var scheduledInterval time.Duration
	// var scheduledTask func() error // Task function comparison is tricky

	scheduler.Add = func(id string, interval time.Duration, task func() error) {
		schedulerAddCalled = true
		scheduledID = id
		scheduledInterval = interval
		// scheduledTask = task
	}
	defer func() { scheduler.Add = oldSchedulerAdd }()

	tests := []struct {
		name               string
		setupMock          func()
		expectedGenXMLGz   bool
		expectedScheduler  bool
		expectedScheduleID string
		// expectedInterval   time.Duration // Interval is random, so hard to assert exact value
	}{
		{
			name: "EPG file does not exist",
			setupMock: func() {
				osStat = func(name string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				}
			},
			expectedGenXMLGz:   true,
			expectedScheduler:  true,
			expectedScheduleID: EPG_TASK_ID, 
		},
		{
			name: "EPG file exists and is up-to-date",
			setupMock: func() {
				osStat = func(name string) (os.FileInfo, error) {
					return mockFileInfo{name: "epg.xml.gz", modTime: time.Now()}, nil
				}
			},
			expectedGenXMLGz:   false,
			expectedScheduler:  true,
			expectedScheduleID: EPG_TASK_ID,
		},
		{
			name: "EPG file exists but is outdated",
			setupMock: func() {
				osStat = func(name string) (os.FileInfo, error) {
					return mockFileInfo{name: "epg.xml.gz", modTime: time.Now().Add(-25 * time.Hour)}, nil
				}
			},
			expectedGenXMLGz:   true,
			expectedScheduler:  true,
			expectedScheduleID: EPG_TASK_ID,
		},
		{
			name: "os.Stat returns an error other than ErrNotExist",
			setupMock: func() {
				osStat = func(name string) (os.FileInfo, error) {
					return nil, errors.New("some other stat error")
				}
			},
			expectedGenXMLGz:   true, 
			expectedScheduler:  true,
			expectedScheduleID: EPG_TASK_ID,
		},
	}

	originalEPGFilePath := EPGFilePath 
	EPGFilePath = "test_init_epg.xml.gz" 
	defer func() { EPGFilePath = originalEPGFilePath }() 

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genXMLGzCalled = false 
			schedulerAddCalled = false 
			scheduledID = ""
			scheduledInterval = 0 
			// scheduledTask = nil // Commented out

			tt.setupMock()
			Init()

			if genXMLGzCalled != tt.expectedGenXMLGz {
				t.Errorf("GenXMLGz called = %v, want %v", genXMLGzCalled, tt.expectedGenXMLGz)
			}
			if schedulerAddCalled != tt.expectedScheduler {
				t.Errorf("scheduler.Add called = %v, want %v", schedulerAddCalled, tt.expectedScheduler)
			}
			if schedulerAddCalled && scheduledID != tt.expectedScheduleID {
				t.Errorf("scheduler.Add ID = %q, want %q", scheduledID, tt.expectedScheduleID)
			}
			// We cannot assert the exact interval due to its random nature.
			// We can check if it's positive, which indicates scheduler.Add was likely called correctly.
			if schedulerAddCalled && scheduledInterval <= 0 {
				t.Errorf("scheduler.Add interval = %v, want > 0", scheduledInterval)
			}

			_ = os.Remove(EPGFilePath)
		})
	}
}

func TestNewProgramme(t *testing.T) {
	tests := []struct {
		name      string
		channelID int
		start     string
		stop      string
		title     string
		desc      string
		category  string
		iconSrc   string // In epg.go, NewProgramme expects just the filename part of icon
		want      Programme
	}{
		{
			name:      "Valid inputs",
			channelID: 1,
			start:     "20230101000000 +0000",
			stop:      "20230101010000 +0000",
			title:     "Test Title",
			desc:      "Test Description",
			category:  "Test Category",
			iconSrc:   "icon.png", // Filename for icon
			want: Programme{
				Channel:  "1",
				Start:    "20230101000000 +0000",
				Stop:     "20230101010000 +0000",
				Title:    TextCDATA{Text: "Test Title"},
				Desc:     TextCDATA{Text: "Test Description"},
				Category: TextCDATA{Text: "Test Category"},
				Icon:     Icon{Src: EPG_POSTER_URL + "/icon.png"}, // Expect constructed URL
			},
		},
		{
			name:      "Empty inputs",
			channelID: 0,
			start:     "",
			stop:      "",
			title:     "",
			desc:      "",
			category:  "",
			iconSrc:   "",
			want: Programme{
				Channel:  "0",
				Start:    "",
				Stop:     "",
				Title:    TextCDATA{Text: ""},
				Desc:     TextCDATA{Text: ""},
				Category: TextCDATA{Text: ""},
				Icon:     Icon{Src: EPG_POSTER_URL + "/"}, // Expect base URL with slash
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProgramme(tt.channelID, tt.start, tt.stop, tt.title, tt.desc, tt.category, tt.iconSrc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProgramme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func normalizeXML(data []byte) (string, error) {
	var epgData EPG
	if err := xml.Unmarshal(data, &epgData); err != nil {
		if strings.TrimSpace(string(data)) == "<tv></tv>" { // Handle specific simple case if EPG is empty
			return "<tv></tv>", nil // Return a canonical empty representation
		}
		return "", fmt.Errorf("failed to unmarshal for normalization: %w (data: %s)", err, string(data))
	}
	normalizedBytes, err := xml.MarshalIndent(epgData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal for normalization: %w", err)
	}
	return string(normalizedBytes), nil
}


func Test_genXML(t *testing.T) {
	oldLog := utils.Log
	currentLogger := utils.GetLogger()
	utils.Log = currentLogger
	var fatalCalled bool
	var fatalMsg string
	oldFatal := currentLogger.Fatal // Use the Fatal method of the specific logger instance
	currentLogger.Fatal = func(v ...interface{}) { // Mock Fatal
		fatalCalled = true
		fatalMsg = fmt.Sprint(v...)
		// Do not os.Exit(1)
	}
	defer func() {
		utils.Log = oldLog         // Restore original global logger
		currentLogger.Fatal = oldFatal // Restore original Fatal method on the instance
	}()


	// Store and restore original URL constants
	originalChannelURL := CHANNEL_URL
	originalEpgURL := EPG_URL
	defer func() {
		CHANNEL_URL = originalChannelURL
		EPG_URL = originalEpgURL
	}()
	
	var currentTestServerHandler func(ctx *fasthttp.RequestCtx) 

	server := fasthttptest.NewServer(t, func(ctx *fasthttp.RequestCtx) { 
		if currentTestServerHandler != nil {
			currentTestServerHandler(ctx)
		} else {
			ctx.Error("Handler not set for test", fasthttp.StatusInternalServerError) 
		}
	})
	defer server.Close()
	
	CHANNEL_URL = server.URL + "/channels" 
	EPG_URL = server.URL + "/epg?offset=%d&channel_id=%d" 


	tests := []struct {
		name                  string
		serverHandler         func(ctx *fasthttp.RequestCtx) 
		wantErr               bool
		wantFatal             bool 
		expectedNormalizedXML string 
	}{
		{
			name: "No channels returned",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				if string(ctx.Path()) == "/channels" { 
					fmt.Fprint(ctx, `{"result":[]}`)
				}
			},
			wantErr: false,
			expectedNormalizedXML: "<tv></tv>",
		},
		{
			name:    "Successful fetch with one channel and its EPG",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				if string(ctx.Path()) == "/channels" {
					fmt.Fprint(ctx, `{"result":[{"channel_id":101,"channel_name":"TestChannel1"}]}`)
				} else if string(ctx.Path()) == "/epg" && string(ctx.QueryArgs().Peek("channel_id")) == "101" {
					// Simulating IST for start/stop times in the mock response
					loc, _ := time.LoadLocation("Asia/Kolkata")
					startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, loc).UnixMilli()
					endTime := time.Date(2023, 1, 1, 1, 0, 0, 0, loc).UnixMilli()
					fmt.Fprintf(ctx, `{"epg":[{"title":"ShowFor101","description":"Desc","showCategory":"Cat","startEpoch":%d,"endEpoch":%d,"poster":"poster.png"}]}`, startTime, endTime)
				}
			},
			wantErr: false,
			expectedNormalizedXML: strings.TrimSpace(`
<tv>
  <channel id="101">
    <display-name lang="en">TestChannel1</display-name>
  </channel>
  <programme channel="101" start="20230101000000 +0530" stop="20230101010000 +0530">
    <title lang="en"><![CDATA[ShowFor101]]></title>
    <desc lang="en"><![CDATA[Desc]]></desc>
    <category lang="en"><![CDATA[Cat]]></category>
    <icon src="https://jiotv.catchup.cdn.jio.com/dare_images/shows/poster.png"></icon>
  </programme>
</tv>`),
		},
		{
			name: "Channel API returns HTTP error",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				if string(ctx.Path()) == "/channels" {
					ctx.SetStatusCode(fasthttp.StatusInternalServerError) 
					fmt.Fprint(ctx, "channel API error")
				}
			},
			wantErr:   true,
			wantFatal: true,
		},
		{
			name: "EPG API returns HTTP error for one channel",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				if string(ctx.Path()) == "/channels" {
					fmt.Fprint(ctx, `{"result":[{"channel_id":101,"channel_name":"TestChannel1"},{"channel_id":102,"channel_name":"TestChannel2"}]}`)
				} else if string(ctx.Path()) == "/epg" {
					loc, _ := time.LoadLocation("Asia/Kolkata")
					startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, loc).UnixMilli()
					endTime := time.Date(2023, 1, 1, 1, 0, 0, 0, loc).UnixMilli()
					if string(ctx.QueryArgs().Peek("channel_id")) == "101" {
						ctx.SetStatusCode(fasthttp.StatusInternalServerError)
						fmt.Fprint(ctx, "EPG API error for 101")
					} else if string(ctx.QueryArgs().Peek("channel_id")) == "102" {
						fmt.Fprintf(ctx, `{"epg":[{"title":"ShowFor102","description":"Desc","showCategory":"Cat","startEpoch":%d,"endEpoch":%d,"poster":"poster.png"}]}`, startTime, endTime)
					}
				}
			},
			wantErr: false,
			expectedNormalizedXML: strings.TrimSpace(`
<tv>
  <channel id="101">
    <display-name lang="en">TestChannel1</display-name>
  </channel>
  <channel id="102">
    <display-name lang="en">TestChannel2</display-name>
  </channel>
  <programme channel="102" start="20230101000000 +0530" stop="20230101010000 +0530">
    <title lang="en"><![CDATA[ShowFor102]]></title>
    <desc lang="en"><![CDATA[Desc]]></desc>
    <category lang="en"><![CDATA[Cat]]></category>
    <icon src="https://jiotv.catchup.cdn.jio.com/dare_images/shows/poster.png"></icon>
  </programme>
</tv>`),
		},
		{
			name: "Channel API returns malformed JSON",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				if string(ctx.Path()) == "/channels" {
					fmt.Fprint(ctx, `{"result": malformed`)
				}
			},
			wantErr:   true,
			wantFatal: true,
		},
		{
			name: "EPG API returns malformed JSON for one channel",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				loc, _ := time.LoadLocation("Asia/Kolkata")
				startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, loc).UnixMilli()
				endTime := time.Date(2023, 1, 1, 1, 0, 0, 0, loc).UnixMilli()
				if string(ctx.Path()) == "/channels" {
					fmt.Fprint(ctx, `{"result":[{"channel_id":101,"channel_name":"TestChannel1"},{"channel_id":102,"channel_name":"TestChannel2"}]}`)
				} else if string(ctx.Path()) == "/epg" {
					if string(ctx.QueryArgs().Peek("channel_id")) == "101" {
						fmt.Fprint(ctx, `{"epg": malformed`)
					} else if string(ctx.QueryArgs().Peek("channel_id")) == "102" {
						fmt.Fprintf(ctx, `{"epg":[{"title":"ShowFor102","description":"Desc","showCategory":"Cat","startEpoch":%d,"endEpoch":%d,"poster":"poster.png"}]}`, startTime, endTime)
					}
				}
			},
			wantErr: false,
			expectedNormalizedXML: strings.TrimSpace(`
<tv>
  <channel id="101">
    <display-name lang="en">TestChannel1</display-name>
  </channel>
  <channel id="102">
    <display-name lang="en">TestChannel2</display-name>
  </channel>
  <programme channel="102" start="20230101000000 +0530" stop="20230101010000 +0530">
    <title lang="en"><![CDATA[ShowFor102]]></title>
    <desc lang="en"><![CDATA[Desc]]></desc>
    <category lang="en"><![CDATA[Cat]]></category>
    <icon src="https://jiotv.catchup.cdn.jio.com/dare_images/shows/poster.png"></icon>
  </programme>
</tv>`),
		},
		{
			name: "EPG API returns epg:null for one channel",
			serverHandler: func(ctx *fasthttp.RequestCtx) {
				loc, _ := time.LoadLocation("Asia/Kolkata")
				startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, loc).UnixMilli()
				endTime := time.Date(2023, 1, 1, 1, 0, 0, 0, loc).UnixMilli()
				if string(ctx.Path()) == "/channels" {
					fmt.Fprint(ctx, `{"result":[{"channel_id":101,"channel_name":"TestChannel1"},{"channel_id":102,"channel_name":"TestChannel2"}]}`)
				} else if string(ctx.Path()) == "/epg" {
					if string(ctx.QueryArgs().Peek("channel_id")) == "101" {
						fmt.Fprint(ctx, `{"epg": null}`)
					} else if string(ctx.QueryArgs().Peek("channel_id")) == "102" {
						fmt.Fprintf(ctx, `{"epg":[{"title":"ShowFor102","description":"Desc","showCategory":"Cat","startEpoch":%d,"endEpoch":%d,"poster":"poster.png"}]}`,startTime, endTime)
					}
				}
			},
			wantErr: false,
			expectedNormalizedXML: strings.TrimSpace(`
<tv>
  <channel id="101">
    <display-name lang="en">TestChannel1</display-name>
  </channel>
  <channel id="102">
    <display-name lang="en">TestChannel2</display-name>
  </channel>
  <programme channel="102" start="20230101000000 +0530" stop="20230101010000 +0530">
    <title lang="en"><![CDATA[ShowFor102]]></title>
    <desc lang="en"><![CDATA[Desc]]></desc>
    <category lang="en"><![CDATA[Cat]]></category>
    <icon src="https://jiotv.catchup.cdn.jio.com/dare_images/shows/poster.png"></icon>
  </programme>
</tv>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fatalCalled = false
			fatalMsg = ""
			currentTestServerHandler = tt.serverHandler 

			gotBytes, err := genXML()

			if tt.wantErr {
				if err == nil && !fatalCalled {
					t.Errorf("genXML() expected an error or fatal log, but got none. XML: %s", string(gotBytes))
				}
				return
			}
			if err != nil {
				t.Errorf("genXML() unexpected error = %v. XML: %s", err, string(gotBytes))
				return
			}
			if fatalCalled {
				if !tt.wantFatal {
					t.Errorf("genXML() called utils.Log.Fatal unexpectedly: %s", fatalMsg)
				}
				return
			}

			if tt.expectedNormalizedXML != "" {
				gotNormalized, normErr := normalizeXML(gotBytes)
				if normErr != nil {
					t.Fatalf("Failed to normalize gotBytes: %v\nGot bytes: %s", normErr, string(gotBytes))
				}
				expectedNormalized, normErr := normalizeXML([]byte(tt.expectedNormalizedXML))
				if normErr != nil {
					t.Fatalf("Failed to normalize expectedXML: %v\nExpected XML: %s", normErr, tt.expectedNormalizedXML)
				}
				if gotNormalized != expectedNormalized {
					t.Errorf("genXML() normalized XML mismatch:\nGot:\n%s\nWant:\n%s", gotNormalized, expectedNormalized)
				}
			}
		})
	}
}


func Test_formatTime(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Kolkata") 
	tests := []struct {
		name string
		args time.Time
		want string
	}{
		{
			name: "Format time correctly",
			args: time.Date(2023, 1, 1, 10, 30, 0, 0, loc),
			want: "20230101103000 +0530",
		},
		{
			name: "Format time with UTC",
			args: time.Date(2023, 1, 1, 10, 30, 0, 0, time.UTC),
			want: "20230101103000 +0000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTime(tt.args); got != tt.want {
				t.Errorf("formatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenXMLGz(t *testing.T) {
	oldLog := utils.Log
	utils.Log = utils.GetLogger()
	defer func() { utils.Log = oldLog }()

	originalGenXML := genXML 
	defer func() { genXML = originalGenXML }() 

	tests := []struct {
		name          string
		filename      string
		mockGenXML    func() ([]byte, error)
		wantErr       bool
		checkFile     bool
		expectedBytes []byte 
	}{
		{
			name:     "Successful generation",
			filename: "test_output.xml.gz",
			mockGenXML: func() ([]byte, error) {
				return []byte("<tv><channel id=\"1\"><display-name>Test</display-name></channel></tv>"), nil
			},
			wantErr:       false,
			checkFile:     true,
			expectedBytes: []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n\t<!DOCTYPE tv SYSTEM \"http://www.w3.org/2006/05/tv\"><tv><channel id=\"1\"><display-name>Test</display-name></channel></tv>"),
		},
		{
			name:     "genXML returns error",
			filename: "test_output_err.xml.gz",
			mockGenXML: func() ([]byte, error) {
				return nil, errors.New("genXML failed")
			},
			wantErr:   true,
			checkFile: false,
		},
		{
			name:     "Empty filename",
			filename: "", 
			mockGenXML: func() ([]byte, error) {
				return []byte("<tv></tv>"), nil
			},
			wantErr:   true, 
			checkFile: false,
		},
		{
			name:     "genXML returns empty bytes, no error",
			filename: "test_output_empty.xml.gz",
			mockGenXML: func() ([]byte, error) {
				return []byte(""), nil // empty XML content, but not an error from genXML
			},
			wantErr:       false,
			checkFile:     true,
			expectedBytes: []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n\t<!DOCTYPE tv SYSTEM \"http://www.w3.org/2006/05/tv\">"), // Only header expected
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genXML = tt.mockGenXML 

			targetFile := tt.filename
			// If filename is empty, os.Create will fail, which is covered by wantErr: true.
			// No need to change EPGFilePath for this specific function's test as it takes filename as arg.

			err := GenXMLGz(tt.filename) 
			if (err != nil) != tt.wantErr {
				t.Errorf("GenXMLGz() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFile && !tt.wantErr {
				f, openErr := os.Open(targetFile)
				if openErr != nil {
					t.Fatalf("Failed to open generated file %s: %v", targetFile, openErr)
				}
				defer f.Close()

				gzr, gzipErr := gzip.NewReader(f)
				if gzipErr != nil {
					t.Fatalf("Failed to create gzip reader for %s: %v", targetFile, gzipErr)
				}
				defer gzr.Close()

				fileBytes, readErr := io.ReadAll(gzr)
				if readErr != nil {
					t.Fatalf("Failed to read gzipped content from %s: %v", targetFile, readErr)
				}
				
				trimmedFileBytes := bytes.TrimSpace(fileBytes)
				trimmedExpectedBytes := bytes.TrimSpace(tt.expectedBytes)

				if !bytes.Equal(trimmedFileBytes, trimmedExpectedBytes) {
					t.Errorf("GenXMLGz() file content = \n%s\n, want \n%s", string(trimmedFileBytes), string(trimmedExpectedBytes))
				}
			}

			if targetFile != "" {
				_ = os.Remove(targetFile)
			}
		})
	}
}
```
