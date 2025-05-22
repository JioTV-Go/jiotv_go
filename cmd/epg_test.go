package cmd

import "testing"

func TestGenEPG(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenEPG(); (err != nil) != tt.wantErr {
				t.Errorf("GenEPG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteEPG(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteEPG(); (err != nil) != tt.wantErr {
				t.Errorf("DeleteEPG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
