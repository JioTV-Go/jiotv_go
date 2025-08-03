package constants

// Version variable for the application version
var Version string

// HTTP Methods
const (
	MethodGET  = "GET"
	MethodPOST = "POST"
)

// Common constants
const (
	// Path prefix for JioTV Go files
	PathPrefix = ".jiotv_go"

	// Error messages
	ErrUnsupportedChannelsFormat = "unsupported or invalid custom channels file format. Supported formats: .json, .yml, .yaml, or valid JSON/YAML content"

	// Limits and thresholds
	MaxRecommendedChannels = 1000
)
