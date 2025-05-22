package television

import (
	"reflect"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestNew(t *testing.T) {
	type args struct {
		credentials *utils.JIOTV_CREDENTIALS
	}
	tests := []struct {
		name string
		args args
		want *Television
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.credentials); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTelevision_Live(t *testing.T) {
	type args struct {
		channelID string
	}
	tests := []struct {
		name    string
		tv      *Television
		args    args
		want    *LiveURLOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tv.Live(tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Television.Live() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Television.Live() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTelevision_Render(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name  string
		tv    *Television
		args  args
		want  []byte
		want1 int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.tv.Render(tt.args.url)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Television.Render() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Television.Render() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestChannels(t *testing.T) {
	tests := []struct {
		name string
		want ChannelsResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Channels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterChannels(t *testing.T) {
	type args struct {
		channels []Channel
		language int
		category int
	}
	tests := []struct {
		name string
		args args
		want []Channel
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterChannels(tt.args.channels, tt.args.language, tt.args.category); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterChannels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceM3U8(t *testing.T) {
	type args struct {
		baseUrl    []byte
		match      []byte
		params     string
		channel_id string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceM3U8(tt.args.baseUrl, tt.args.match, tt.args.params, tt.args.channel_id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceM3U8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceTS(t *testing.T) {
	type args struct {
		baseUrl []byte
		match   []byte
		params  string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceTS(tt.args.baseUrl, tt.args.match, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceTS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceAAC(t *testing.T) {
	type args struct {
		baseUrl []byte
		match   []byte
		params  string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceAAC(tt.args.baseUrl, tt.args.match, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceAAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceKey(t *testing.T) {
	type args struct {
		match      []byte
		params     string
		channel_id string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceKey(tt.args.match, tt.args.params, tt.args.channel_id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSLChannel(t *testing.T) {
	type args struct {
		channelID string
	}
	tests := []struct {
		name    string
		args    args
		want    *LiveURLOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSLChannel(tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSLChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSLChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}
