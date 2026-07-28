package headers

// HTTP Header Names
const (
	// Standard HTTP headers
	ContentType    = "Content-Type"
	Accept         = "Accept"
	AcceptEncoding = "Accept-Encoding"
	UserAgent      = "User-Agent"
	Authorization  = "Authorization"
	Host           = "Host"
	AccessToken    = "accessToken"

	// Custom headers used by JioTV API
	DeviceType  = "devicetype"
	VersionCode = "versionCode"
	OS          = "os"
	XAPIKey     = "x-api-key"

	// Headers required by the premium provider APIs
	XAccessToken = "X-AccessToken"
	XVersionCode = "X-VersionCode"
	XPlatform    = "X-Platform"
)

// HTTP Header Values
const (
	// Content types
	ContentTypeJSON            = "application/json"
	ContentTypeJSONCharsetUTF8 = "application/json; charset=utf-8"

	// Accept values
	AcceptJSON         = "application/json"
	AcceptEncodingGzip = "gzip"

	// User agents
	UserAgentOkHttp = "okhttp/4.2.2"
	UserAgentPlayTV = "plaYtv/7.1.7 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7"

	// Device info
	DeviceTypePhone = "phone"
	OSAndroid       = "android"
	VersionCode315  = "315"
	VersionCode389  = "389"
	// VersionCode413 matches the current JioTV app build (v7.1.7). The premium
	// provider APIs require it: the plans endpoint rejects requests without a
	// current versionCode/X-Platform pair.
	VersionCode413 = "413"

	// Platform values for the X-Platform header used by premium provider APIs
	PlatformAndroid       = "android"
	PlatformAndroidMobile = "android_mobile"

	// API Key
	APIKeyJio = "l7xx75e822925f184370b2e25170c5d5820a"
)
