package television

import (
	"github.com/valyala/fasthttp"
)

// Television struct to store credentials and client required for making requests to JioTV API
type Television struct {
	AccessToken string
	SsoToken    string
	Crm         string
	UniqueID    string
	headers     map[string]string
	Client      *fasthttp.Client
}

// Channel represents Individual channel details from JioTV API
type Channel struct {
	ID       int    `json:"channel_id"`
	Name     string `json:"channel_name"`
	URL      string `json:"channel_url"`
	LogoURL  string `json:"logoUrl"`
	Category int    `json:"channelCategoryId"`
	Language int    `json:"channelLanguageId"`
	IsHD     bool   `json:"isHD"`
}

// ChannelsResponse is the response body for channels from JioTV API
type ChannelsResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Result  []Channel `json:"result"`
}

// Bitrates represents Quality levels for live streams for JioTV API
type Bitrates struct {
	Auto   string `json:"auto"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Medium string `json:"medium"`
}

type MPD struct {
	Result   string   `json:"result"`
	Key      string   `json:"key"`
	Bitrates Bitrates `json:"bitrates"`
}

// LiveURLOutput represents Response of live stream URL request to JioTV API
type LiveURLOutput struct {
    SonyVodStitchAdsCpCustomerID struct {
        Midroll   string `json:"midroll"`
        Postroll  string `json:"postroll"`
        Preroll   string `json:"preroll"`
    } `json:"sonyVodStitchAdsCpCustomerID"`
    VmapURL string `json:"vmapUrl"`
    Bitrates Bitrates `json:"bitrates"`
    Code        int     `json:"code"`
    ContentID   float64 `json:"contentId"`
    CurrentTime float64 `json:"currentTime"`
    EndTime     float64 `json:"endTime"`
    Message     string  `json:"message"`
    Result      string  `json:"result"`
    StartTime   float64 `json:"startTime"`
    VodStitch   bool    `json:"vodStitch"`
	Mpd      MPD      `json:"mpd"`
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
