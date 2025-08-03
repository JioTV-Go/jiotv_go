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
	Connection     = "Connection"

	// JioTV API authentication headers
	AccessToken  = "accesstoken"
	SsoToken     = "ssotoken"
	SubscriberID = "subscriberId"
	UniqueID     = "uniqueId"
	CRMID        = "crmid"
	ChannelID    = "channelid"

	// JioTV API device and platform headers
	DeviceType   = "devicetype"
	DeviceID     = "deviceId"
	VersionCode  = "versionCode"
	OS           = "os"
	OSVersion    = "osVersion"
	XPlatform    = "x-platform"
	AppName      = "appName"
	UserGroup    = "usergroup"
	SerialNumber = "srno"
	XAPIKey      = "x-api-key"
)

// HTTP Header Values
const (
	// Content types
	ContentTypeJSON            = "application/json"
	ContentTypeJSONCharsetUTF8 = "application/json; charset=utf-8"
	ContentTypeOctetStream     = "application/octet-stream"

	// Connection values
	ConnectionKeepAlive = "keep-alive"

	// Accept values
	AcceptJSON                = "application/json"
	AcceptEncodingGzip        = "gzip"
	AcceptEncodingGzipDeflate = "gzip, deflate"

	// User agents
	UserAgentOkHttp    = "okhttp/4.2.2"
	UserAgentPlayTV    = "plaYtv/7.0.5 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7"
	UserAgentPlayTVNew = "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7"

	// Device info
	DeviceTypePhone = "phone"
	OSAndroid       = "android"
	OSVersion13     = "13"
	VersionCode315  = "315"
	VersionCode330  = "330"

	// JioTV specific values
	AppNameJioTV     = "RJIL_JioTV"
	UserGroupDefault = "tvYR7NSNn7rymo3F"

	// API Key
	APIKeyJio = "l7xx75e822925f184370b2e25170c5d5820a"
)
