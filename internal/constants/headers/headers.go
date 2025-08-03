package headers

// HTTP Header Names
const (
	// Standard HTTP headers
	ContentType     = "Content-Type"
	Accept          = "Accept"
	AcceptEncoding  = "Accept-Encoding"
	UserAgent       = "User-Agent"
	Authorization   = "Authorization"
	Host            = "Host"
	AccessToken     = "accessToken"

	// Custom headers used by JioTV API
	DeviceType    = "devicetype"
	VersionCode   = "versionCode"
	OS            = "os"
	XAPIKey       = "x-api-key"
)

// HTTP Header Values
const (
	// Content types
	ContentTypeJSON = "application/json"
	ContentTypeJSONCharsetUTF8 = "application/json; charset=utf-8"

	// Accept values
	AcceptJSON = "application/json"
	AcceptEncodingGzip = "gzip"

	// User agents
	UserAgentOkHttp = "okhttp/4.2.2"
	UserAgentPlayTV = "plaYtv/7.0.5 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7"

	// Device info
	DeviceTypePhone = "phone"
	OSAndroid = "android"
	VersionCode315 = "315"

	// API Key
	APIKeyJio = "l7xx75e822925f184370b2e25170c5d5820a"
)