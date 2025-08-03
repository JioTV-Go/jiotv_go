package tasks

// Task IDs for scheduled tasks
const (
	// Authentication-related tasks
	RefreshTokenTaskID    = "jiotv_refresh_token"
	RefreshSSOTokenTaskID = "jiotv_refresh_sso_token"
	HealthCheckTaskID     = "jiotv_token_health_check"
	
	// EPG-related tasks
	EPGTaskID = "jiotv_epg"
)