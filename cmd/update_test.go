package cmd

import (
	"reflect"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestUpdate(t *testing.T) {
	type args struct {
		currentVersion string
		customVersion  string
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
			if err := Update(tt.args.currentVersion, tt.args.customVersion); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getLatestRelease(t *testing.T) {
	type args struct {
		customVersion string
	}
	tests := []struct {
		name    string
		args    args
		want    *Release
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLatestRelease(tt.args.customVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLatestRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLatestRelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloadBinary(t *testing.T) {
	type args struct {
		url        string
		outputPath string
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
			if err := downloadBinary(tt.args.url, tt.args.outputPath); (err != nil) != tt.wantErr {
				t.Errorf("downloadBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_replaceBinary(t *testing.T) {
	type args struct {
		newBinaryPath string
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
			if err := replaceBinary(tt.args.newBinaryPath); (err != nil) != tt.wantErr {
				t.Errorf("replaceBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_compareVersions(t *testing.T) {
	type args struct {
		currentVersion string
		latestVersion  string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareVersions(tt.args.currentVersion, tt.args.latestVersion); got != tt.want {
				t.Errorf("compareVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_atoiOrZero(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := atoiOrZero(tt.args.s); got != tt.want {
				t.Errorf("atoiOrZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUpdateAvailable(t *testing.T) {
	type args struct {
		currentVersion string
		customVersion  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUpdateAvailable(tt.args.currentVersion, tt.args.customVersion); got != tt.want {
				t.Errorf("IsUpdateAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintIfUpdateAvailable(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrintIfUpdateAvailable(tt.args.c)
		})
	}
}
