package cmd

import "testing"

func TestAutoStart(t *testing.T) {
	type args struct {
		extraArgs string
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
			if err := AutoStart(tt.args.extraArgs); (err != nil) != tt.wantErr {
				t.Errorf("AutoStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isTermux(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTermux(); got != tt.want {
				t.Errorf("isTermux() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getConsentFromUser(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getConsentFromUser(); got != tt.want {
				t.Errorf("getConsentFromUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grep(t *testing.T) {
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
		// TODO: Add test cases.
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

func Test_addToBashrc(t *testing.T) {
	type args struct {
		filename string
		line     string
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
			if err := addToBashrc(tt.args.filename, tt.args.line); (err != nil) != tt.wantErr {
				t.Errorf("addToBashrc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeFromBashrc(t *testing.T) {
	type args struct {
		filename string
		line     string
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
			if err := removeFromBashrc(tt.args.filename, tt.args.line); (err != nil) != tt.wantErr {
				t.Errorf("removeFromBashrc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
