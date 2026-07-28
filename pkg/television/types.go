package television

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

// Television struct to store credentials and client required for making requests to JioTV API
type Television struct {
	AccessToken string
	SsoToken    string
	Crm         string
	UniqueID    string
	Headers     map[string]string
	Client      *fasthttp.Client
}

// Channel represents Individual channel details from JioTV API
type Channel struct {
	ID                 string `json:"channel_id"`
	Name               string `json:"channel_name"`
	URL                string `json:"channel_url"`
	LogoURL            string `json:"logoUrl"`
	Category           int    `json:"channelCategoryId"`
	Language           int    `json:"channelLanguageId"`
	IsHD               bool   `json:"isHD"`
	IsCatchupAvailable bool   `json:"isCatchupAvailable"`

	// RequiresSubscription marks channels the playback API refuses with
	// "No eligible plans found" unless the account subscribes separately.
	// It is derived from business_type, not from the API's own is_premium flag:
	// is_premium is set on many freely playable channels and mispredicts
	// roughly one in five, while business_type=="premium" matched playability
	// exactly across a sample covering all four business types.
	RequiresSubscription bool `json:"requiresSubscription"`
}

// businessTypePremium is the business_type value marking channels that need a
// separate subscription.
const businessTypePremium = "premium"

// UnmarshalJSON to Override Channel.ID to convert int from json to string
func (c *Channel) UnmarshalJSON(b []byte) error {
	type Alias Channel
	aux := &struct {
		ID           int    `json:"channel_id"`
		BusinessType string `json:"business_type"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	c.ID = strconv.Itoa(aux.ID)
	c.RequiresSubscription = strings.EqualFold(strings.TrimSpace(aux.BusinessType), businessTypePremium)
	return nil
}

// ChannelsResponse is the response body for channels from JioTV API
type ChannelsResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Result  []Channel `json:"result"`
}

// PremiumProvider represents a premium partner available for the logged-in account.
// PremiumProvider is a premium provider available on the account.
//
// Two different identifier namespaces are involved. ID is the entitlement
// identifier from the plans API (for example "Z0177"), while ProviderID is the
// content identifier used by the provider config and catalog APIs (for example
// "200169"). Only ProviderID can be used to browse or play content.
type PremiumProvider struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerId,omitempty"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Image      string `json:"image,omitempty"`

	// matchNames holds the provider taxonomy names reported by the plans API,
	// most authoritative first, used to look the provider up in the provider
	// directory. Not part of the API response.
	matchNames []string
}

// PlansResponse is the response payload for the account plans API
// (tv.media.jio.com/apis/v2.1/plans/get).
//
// PackageInfo holds the plans actually active on the account and is the
// authoritative source for premium entitlements. Result is the older nested
// shape, retained as a fallback.
type PlansResponse struct {
	PackageInfo []ActiveSubscriptionPlan `json:"PackageInfo"`
	Result      PlansResult              `json:"result"`
}

// ActiveSubscriptionPlan represents one plan subscribed on the account.
type ActiveSubscriptionPlan struct {
	PlanID        string                 `json:"planid"`
	IsActive      bool                   `json:"isactive"`
	PackageDetail ActiveSubscriptionPack `json:"packageDetail"`
}

// ActiveSubscriptionPack holds the pack details attached to an active plan.
type ActiveSubscriptionPack struct {
	PackageID   string         `json:"package_id"`
	PackageName string         `json:"package_name"`
	Providers   []PlanProvider `json:"providers"`
}

// PlansResult wraps plans entries returned by plans API.
type PlansResult struct {
	Plans []Plan `json:"plans"`
}

// Plan represents a single plan returned by plans API.
type Plan struct {
	Providers []PlanProvider `json:"providers"`
}

// PlanProvider represents a provider entry attached to a plan. The current API
// returns id/name/image; the older nested shape used providerId/providerName,
// so both spellings are accepted.
type PlanProvider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Image        string `json:"image"`
	Provider     string `json:"provider"`
	ProviderID   string `json:"providerId"`
	ProviderName string `json:"providerName"`
}

// resolvedID returns the provider identifier regardless of response shape.
func (p PlanProvider) resolvedID() string {
	if id := strings.TrimSpace(p.ID); id != "" {
		return id
	}
	return strings.TrimSpace(p.ProviderID)
}

// resolvedName returns the provider display name regardless of response shape.
func (p PlanProvider) resolvedName() string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return name
	}
	return strings.TrimSpace(p.ProviderName)
}

type PremiumProviderFilterResponse struct {
	Data PremiumProviderFilterData `json:"data"`
}

type PremiumProviderFilterData struct {
	Provider []PremiumProviderFilters `json:"provider"`
}

type PremiumProviderFilters struct {
	ProviderID string                  `json:"providerId"`
	Provider   string                  `json:"provider"`
	Filters    []PremiumProviderFilter `json:"filters"`
}

type PremiumProviderFilter struct {
	FilterName string                       `json:"filterName"`
	Values     []PremiumProviderFilterValue `json:"values"`
}

type PremiumProviderFilterValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PremiumProviderCatalogEnvelope struct {
	Code         int                      `json:"Code"`
	CodeLower    int                      `json:"code"`
	Message      string                   `json:"Message"`
	MessageLower string                   `json:"message"`
	Data         []map[string]interface{} `json:"Data"`
	DataLower    []map[string]interface{} `json:"data"`
}

type PremiumProviderCatalogItem struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Subtitle      string `json:"subtitle,omitempty"`
	Description   string `json:"description,omitempty"`
	ImageURL      string `json:"imageUrl,omitempty"`
	StreamType    string `json:"streamType"`
	ProviderID    string `json:"providerId"`
	ChannelID     string `json:"channelId,omitempty"`
	ContentID     string `json:"contentId,omitempty"`
	SubCategoryID string `json:"subCategoryId,omitempty"`
}

type PremiumProviderCatalogResult struct {
	ProviderID string                       `json:"providerId"`
	Code       int                          `json:"code"`
	Message    string                       `json:"message"`
	Result     []PremiumProviderCatalogItem `json:"result"`
}

type PremiumProviderPlayRequest struct {
	StreamType    string `json:"streamType" form:"streamType"`
	ChannelID     string `json:"channelId" form:"channelId"`
	ContentID     string `json:"contentId" form:"contentId"`
	SubCategoryID string `json:"subCategoryId" form:"subCategoryId"`
}

// Bitrates represents Quality levels for live streams for JioTV API
type Bitrates struct {
	Auto   string `json:"auto"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Medium string `json:"medium"`
}

// MPD represents the DASH section of a playback response. Live channels return
// the stream URLs under "bitrates", while premium provider (SVOD) content
// returns a single "auto" URL at this level.
type MPD struct {
	Result   string   `json:"result"`
	Key      string   `json:"key"`
	Bitrates Bitrates `json:"bitrates"`
	Auto     string   `json:"auto"`
	High     string   `json:"high"`
	Low      string   `json:"low"`
	Medium   string   `json:"medium"`
}

// ResolvedBitrates merges the two MPD shapes into one set of stream URLs.
func (m MPD) ResolvedBitrates() Bitrates {
	resolved := m.Bitrates
	if resolved.Auto == "" {
		resolved.Auto = m.Auto
	}
	if resolved.High == "" {
		resolved.High = m.High
	}
	if resolved.Medium == "" {
		resolved.Medium = m.Medium
	}
	if resolved.Low == "" {
		resolved.Low = m.Low
	}
	return resolved
}

// LiveURLOutput represents Response of live stream URL request to JioTV API
type LiveURLOutput struct {
	SonyVodStitchAdsCpCustomerID struct {
		Midroll  string `json:"midroll"`
		Postroll string `json:"postroll"`
		Preroll  string `json:"preroll"`
	} `json:"sonyVodStitchAdsCpCustomerID"`
	VmapURL     string   `json:"vmapUrl"`
	Bitrates    Bitrates `json:"bitrates"`
	Code        int      `json:"code"`
	ContentID   float64  `json:"contentId"`
	CurrentTime float64  `json:"currentTime"`
	EndTime     float64  `json:"endTime"`
	Message     string   `json:"message"`
	Result      string   `json:"result"`
	StartTime   float64  `json:"startTime"`
	VodStitch   bool     `json:"vodStitch"`
	Mpd         MPD      `json:"mpd"`
	M3u8        Bitrates `json:"m3u8"`
	IsDRM       bool     `json:"isDRM"`
	KeyURL      string   `json:"keyUrl"`
	ExtID       string   `json:"extId"`
	AlgoName    string   `json:"algoName"`
	Hdnea       string   `json:"-"` // parsed from URLs in Live response (hdnea query param); may rotate via Set-Cookie (__hdnea__) on m3u8/ts requests
}

// ResolvedLicenseURL returns the DRM license URL. Live channels carry it in
// mpd.key, while premium provider content returns a top-level keyUrl.
func (l LiveURLOutput) ResolvedLicenseURL() string {
	if licenseURL := strings.TrimSpace(l.KeyURL); licenseURL != "" {
		return licenseURL
	}
	return strings.TrimSpace(l.Mpd.Key)
}

// HasDRMStream reports whether the response carries a DASH stream that needs a
// DRM license, regardless of whether the isDRM flag was set.
func (l LiveURLOutput) HasDRMStream() bool {
	return l.Mpd.ResolvedBitrates().Auto != "" && l.ResolvedLicenseURL() != ""
}

// CategoryMap represents Categories for channels
var CategoryMap = map[int]string{
	0:  "All Categories",
	5:  "Entertainment",
	6:  "Movies",
	7:  "Kids",
	8:  "Sports",
	9:  "Lifestyle",
	10: "Infotainment",
	12: "News",
	13: "Music",
	15: "Devotional",
	16: "Business",
	17: "Educational",
	18: "Shopping",
	19: "JioDarshan",
}

// LanguageMap represents Languages for channels
var LanguageMap = map[int]string{
	0:  "All Languages",
	1:  "Hindi",
	2:  "Marathi",
	3:  "Punjabi",
	4:  "Urdu",
	5:  "Bengali",
	6:  "English",
	7:  "Malayalam",
	8:  "Tamil",
	9:  "Gujarati",
	10: "Odia",
	11: "Telugu",
	12: "Bhojpuri",
	13: "Kannada",
	14: "Assamese",
	15: "Nepali",
	16: "French",
	18: "Other",
}

var SONY_CHANNELS = map[string]string{
	"sonyhd":         "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L2RCZHdPaUdhUXZ5MFRBMXpPc2pWNncvbWFzdGVyLm0zdTg=",
	"sonysabhd":      "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L0NyVGl2a0RFU1dxd3ZVajN6RkVZRUEvbWFzdGVyLm0zdTg=",
	"sonypal":        "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L2RoUHJHUndEUnZ1TVF0bWx6cHB6UVEvbWFzdGVyLm0zdTg=",
	"sonypixhd":      "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L3g3clhXZDJFUloydHZ5UVdQbU8xSEEvbWFzdGVyLm0zdTg=",
	"sonymaxhd":      "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L1VjakhOSm1DUTFXUmxHS2xabTczUUEvbWFzdGVyLm0zdTg=",
	"sonymax2":       "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L01kUTVaeS1QU3JhT2NjWHU4amZsQ2cvbWFzdGVyLm0zdTg=",
	"sonywah":        "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L2dYNXJDQmY2UTctRDVBV1ktc292elEvbWFzdGVyLm0zdTg=",
	"sonyten1hd":     "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L3dHNzVuNVU4UnJPS2lGemFXT2JYYkEvbWFzdGVyLm0zdTg=",
	"sonyten2hd":     "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L1Y5aC1peU94UmlHcDQxcHBRU2NEU1EvbWFzdGVyLm0zdTg=",
	"sonyten3hd":     "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L2x0c0NHN1RCU0NTRG15cTByUXR2U0EvbWFzdGVyLm0zdTg=",
	"sonyten4hd":     "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L3NtWXliSV9KVG9XYUh6d294U0U5cUEvbWFzdGVyLm0zdTg=",
	"sonyten5hd":     "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50L1NsZV9UUjhyUUl1WkhXenNoRVhZalEvbWFzdGVyLm0zdTg=",
	"sonybbcearthhd": "aHR0cHM6Ly9kYWkuZ29vZ2xlLmNvbS9saW5lYXIvaGxzL2V2ZW50LzZiVldZSUtHUzBDSWEtY09wWlpKUFEvbWFzdGVyLm0zdTg=",
}

var SONY_JIO_MAP = map[string]string{
	"sl291":  "sonyhd",
	"sl154":  "sonysabhd",
	"sl474":  "sonypal",
	"sl762":  "sonypixhd",
	"sl476":  "sonymaxhd",
	"sl483":  "sonymax2",
	"sl1393": "sonywah",
	"sl162":  "sonyten1hd",
	"sl891":  "sonyten2hd",
	"sl892":  "sonyten3hd",
	"sl1772": "sonyten4hd",
	"sl155":  "sonyten5hd",
	"sl852":  "sonybbcearthhd",
}

// CustomChannel represents a custom channel definition from configuration file
type CustomChannel struct {
	ID       string `json:"id" yaml:"id"`
	Name     string `json:"name" yaml:"name"`
	URL      string `json:"url" yaml:"url"`
	LogoURL  string `json:"logo_url" yaml:"logo_url"`
	Category int    `json:"category" yaml:"category"`
	Language int    `json:"language" yaml:"language"`
	IsHD     bool   `json:"is_hd" yaml:"is_hd"`
}

// CustomChannelsConfig represents the structure of custom channels configuration file
type CustomChannelsConfig struct {
	Channels []CustomChannel `json:"channels" yaml:"channels"`
}

var SONY_CHANNELS_API = []Channel{
	{
		ID:       "sl291",
		Name:     "SL Sony HD",
		Language: 1,
		Category: 5,
		IsHD:     true,
		LogoURL:  "Sony_HD.png",
	},
	{
		ID:       "sl154",
		Name:     "SL Sony SAB HD",
		Language: 1,
		Category: 5,
		IsHD:     true,
		LogoURL:  "Sony_SAB_HD.png",
	},
	{
		ID:       "sl474",
		Name:     "SL Sony PAL",
		Language: 1,
		Category: 5,
		IsHD:     false,
		LogoURL:  "Sony_Pal.png",
	},
	{
		ID:       "sl762",
		Name:     "SL Sony PIX HD",
		Language: 6,
		Category: 6,
		IsHD:     true,
		LogoURL:  "Sony_Pix_HD.png",
	},
	{
		ID:       "sl476",
		Name:     "SL Sony MAX HD",
		Language: 1,
		Category: 6,
		IsHD:     true,
		LogoURL:  "Sony_Max_HD.png",
	},
	{
		ID:       "sl483",
		Name:     "SL Sony MAX 2",
		Language: 1,
		Category: 6,
		IsHD:     false,
		LogoURL:  "Sony_MAX2.png",
	},
	// Disabled as it requires CORS bypass
	// {
	// 	ID:       "sl1393",
	// 	Name:     "SL Sony WAH",
	// 	Language: 1,
	// 	Category: 5,
	// 	IsHD:     false,
	// 	LogoURL:  "Sony_Wah.png",
	// },
	{
		ID:       "sl162",
		Name:     "SL Sony TEN 1 HD",
		Language: 6,
		Category: 8,
		IsHD:     true,
		LogoURL:  "Ten_HD.png",
	},
	{
		ID:       "sl891",
		Name:     "SL Sony TEN 2 HD",
		Language: 6,
		Category: 8,
		IsHD:     true,
		LogoURL:  "Ten2_HD.png",
	},
	{
		ID:       "sl892",
		Name:     "SL Sony TEN 3 HD",
		Language: 1,
		Category: 8,
		IsHD:     true,
		LogoURL:  "Ten3_HD.png",
	},
	{
		ID:       "sl1772",
		Name:     "SL Sony TEN 4 HD",
		Language: 8,
		Category: 8,
		IsHD:     true,
		LogoURL:  "Ten_4_HD_Tamil.png",
	},
	{
		ID:       "sl155",
		Name:     "SL Sony TEN 5 HD",
		Language: 6,
		Category: 8,
		IsHD:     true,
		LogoURL:  "Six_HD.png",
	},
	{
		ID:       "sl852",
		Name:     "SL Sony BBC Earth HD",
		Language: 6,
		Category: 10,
		IsHD:     true,
		LogoURL:  "Sony_BBC_Earth_HD_English.png",
	},
}
