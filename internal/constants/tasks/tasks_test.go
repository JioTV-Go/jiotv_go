package tasks

import "testing"

func TestTaskIDConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"RefreshTokenTaskID", RefreshTokenTaskID, "jiotv_refresh_token"},
		{"RefreshSSOTokenTaskID", RefreshSSOTokenTaskID, "jiotv_refresh_sso_token"},
		{"HealthCheckTaskID", HealthCheckTaskID, "jiotv_token_health_check"},
		{"EPGTaskID", EPGTaskID, "jiotv_epg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestTaskIDsNotEmpty(t *testing.T) {
	taskIDs := []struct {
		name  string
		value string
	}{
		{"RefreshTokenTaskID", RefreshTokenTaskID},
		{"RefreshSSOTokenTaskID", RefreshSSOTokenTaskID},
		{"HealthCheckTaskID", HealthCheckTaskID},
		{"EPGTaskID", EPGTaskID},
	}

	for _, task := range taskIDs {
		t.Run(task.name+"_not_empty", func(t *testing.T) {
			if task.value == "" {
				t.Errorf("%s constant is empty", task.name)
			}
		})
	}
}