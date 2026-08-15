package urls

// Domain constants
const (
	// JioTV API domains
	JioTVAPIDomain = "jiotvapi.media.jio.com"
	TVMediaDomain  = "tv.media.jio.com"
	JioTVCDNDomain = "jiotvapi.cdn.jio.com"

	// Auth and API domains
	AuthMediaDomain = "auth.media.jio.com"
	APIJioDomain    = "api.jio.com"

	// EPG and data domains
	JioTVDataCDNDomain    = "jiotv.data.cdn.jio.com"
	JioTVCatchupCDNDomain = "jiotv.catchup.cdn.jio.com"
)

// Complete URL endpoints
const (
	// Authentication URLs
	RefreshTokenURL    = "https://auth.media.jio.com/tokenservice/apis/v1/refreshtoken?langId=6"
	RefreshSSOTokenURL = "https://tv.media.jio.com/apis/v2.0/loginotp/refresh?langId=6"

	// Channel listing URLs
	ChannelsAPIURL = "https://jiotvapi.cdn.jio.com/apis/v3.1/getMobileChannelList/get/?langId=6&os=android&devicetype=phone&usertype=JIO&version=315&langId=6"
	ChannelURL     = "https://jiotv.data.cdn.jio.com/apis/v3.1/getMobileChannelList/get/?os=android&devicetype=phone&usertype=tvYR7NSNn7rymo3F"

	// Premium provider URLs
	PlansAPIURL                = "https://tv.media.jio.com/apis/v2.1/plans/get?langId=6"
	ActivePlansAPIURL          = "https://jiotvapi.media.jio.com/userservice/apis/v1/plans"
	ProviderConfigAPIBaseURL   = "https://api-jiotv.media.jio.com"
	ProviderMetadataAPIBaseURL = "https://content-jiotv.media.jio.com"

	// Image base path for premium catalog artwork returned as relative paths
	ImageBaseURL = "https://img.media.jio.com/"

	// EPG URLs
	EPGURL            = "https://jiotv.data.cdn.jio.com/apis/v1.3/getepg/get?offset=%d&channel_id=%d"
	EPGPosterURL      = "https://jiotv.catchup.cdn.jio.com/dare_images/shows"
	EPGPosterURLSlash = "https://jiotv.catchup.cdn.jio.com/dare_images/shows/"
)

// URL path patterns (for string formatting)
const (
	// Playback URL patterns
	PlaybackAPIPath = "/playback/apis/v1.1/geturl?langId=6"
)
