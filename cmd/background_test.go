package cmd

import "testing"

func Test_readPIDPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readPIDPath()
		})
	}
}

func TestRunInBackground(t *testing.T) {
	type args struct {
		args       string
		configPath string
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
			if err := RunInBackground(tt.args.args, tt.args.configPath); (err != nil) != tt.wantErr {
				t.Errorf("RunInBackground() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStopBackground(t *testing.T) {
	type args struct {
		configPath string
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
			if err := StopBackground(tt.args.configPath); (err != nil) != tt.wantErr {
				t.Errorf("StopBackground() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
