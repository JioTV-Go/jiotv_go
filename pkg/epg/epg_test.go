package epg

import (
	"reflect"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
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
	tests := []struct {
		name    string
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := genXML()
			if (err != nil) != tt.wantErr {
				t.Errorf("genXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genXML() = %v, want %v", got, tt.want)
			}
		})
	}
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
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenXMLGz(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("GenXMLGz() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
