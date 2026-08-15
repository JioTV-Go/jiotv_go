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
	UserAgentOkHttp = "okhttp/4.12.0"
	UserAgentPlayTV = "plaYtv/7.1.8 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7"

	// Device info
	DeviceTypePhone = "phone"
	OSAndroid       = "android"
	// VersionCode422 mirrors BuildConfig.VERSION_CODE of the current JioTV app
	// (v7.1.8). Several APIs reject stale builds, so keep this in step with the
	// app rather than pinning an older value.
	VersionCode422 = "422"

	// Platform values for the X-Platform header used by premium provider APIs
	PlatformAndroid       = "android"
	PlatformAndroidMobile = "android_mobile"

	// APIKeyJio mirrors BuildConfig.API_KEY of the current JioTV app (v7.1.8).
	APIKeyJio = "l7xx938b6684ee9e4bbe8831a9a682b8e19f"
)
