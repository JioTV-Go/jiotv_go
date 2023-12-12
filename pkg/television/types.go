package television

import (
	"encoding/json"
	"strconv"

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
	ID       string    `json:"channel_id"`
	Name     string `json:"channel_name"`
	URL      string `json:"channel_url"`
	LogoURL  string `json:"logoUrl"`
	Category int    `json:"channelCategoryId"`
	Language int    `json:"channelLanguageId"`
	IsHD     bool   `json:"isHD"`
}

// Override Channel.ID to convert int from json to string
func (c *Channel) UnmarshalJSON(b []byte) error {
	type Alias Channel
	aux := &struct {
		ID int `json:"channel_id"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	c.ID = strconv.Itoa(aux.ID)
	return nil
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
	"sonyhd": "https://dai.google.com/linear/hls/event/dBdwOiGaQvy0TA1zOsjV6w/master.m3u8",
	"sonysabhd": "https://dai.google.com/linear/hls/event/CrTivkDESWqwvUj3zFEYEA/master.m3u8",
	"sonypal": "https://dai.google.com/linear/hls/event/dhPrGRwDRvuMQtmlzppzQQ/master.m3u8",
	"sonypixhd": "https://dai.google.com/linear/hls/event/x7rXWd2ERZ2tvyQWPmO1HA/master.m3u8",
	"sonymaxhd": "https://dai.google.com/linear/hls/event/UcjHNJmCQ1WRlGKlZm73QA/master.m3u8",
	"sonymax2": "https://dai.google.com/linear/hls/event/MdQ5Zy-PSraOccXu8jflCg/master.m3u8",
	"sonywah": "https://dai.google.com/linear/hls/event/gX5rCBf6Q7-D5AWY-sovzQ/master.m3u8",
	"sonyten1hd": "https://dai.google.com/linear/hls/event/wG75n5U8RrOKiFzaWObXbA/master.m3u8",
	"sonyten2hd": "https://dai.google.com/linear/hls/event/V9h-iyOxRiGp41ppQScDSQ/master.m3u8",
	"sonyten3hd": "https://dai.google.com/linear/hls/event/ltsCG7TBSCSDmyq0rQtvSA/master.m3u8",
	"sonyten4hd": "https://dai.google.com/linear/hls/event/smYybI_JToWaHzwoxSE9qA/master.m3u8",
	"sonyten5hd": "https://dai.google.com/linear/hls/event/Sle_TR8rQIuZHWzshEXYjQ/master.m3u8",
	"sonybbcearthhd": "https://dai.google.com/linear/hls/event/6bVWYIKGS0CIa-cOpZZJPQ/master.m3u8",
}

var SONY_JIO_MAP = map[string]string{
	"sl291": "sonyhd",
	"sl154": "sonysabhd",
	"sl474": "sonypal",
	"sl762": "sonypixhd",
	"sl476": "sonymaxhd",
	"sl483": "sonymax2",
	"sl1393": "sonywah",
	"sl162": "sonyten1hd",
	"sl891": "sonyten2hd",
	"sl892": "sonyten3hd",
	"sl1772": "sonyten4hd",
	"sl155": "sonyten5hd",
	"sl852": "sonybbcearthhd",
}

var SONY_CHANNELS_API = []Channel{
	{
		ID:          "sl291",
		Name:        "SL: Sony HD",
		Language:    1,
		Category:    5,
		IsHD:        true,
		LogoURL:    "Sony_HD.png",
	},
	{
		ID:          "sl154",
		Name:        "SL: Sony SAB HD",
		Language:    1,
		Category:    5,
		IsHD:        true,
		LogoURL:    "Sony_SAB_HD.png",
	},
	{
		ID:          "sl474",
		Name:        "SL: Sony PAL",
		Language:    1,
		Category:    5,
		IsHD:        false,
		LogoURL:    "Sony_Pal.png",
	},
	{
		ID:          "sl762",
		Name:        "SL: Sony PIX HD",
		Language:    6,
		Category:    6,
		IsHD:        true,
		LogoURL:    "Sony_Pix_HD.png",
	},
	{
		ID:          "sl476",
		Name:        "SL: Sony MAX HD",
		Language:    1,
		Category:    6,
		IsHD:        true,
		LogoURL:    "Sony_Max_HD.png",
	},
	{
		ID:          "sl483",
		Name:        "SL: Sony MAX 2",
		Language:    1,
		Category:    6,
		IsHD:        false,
		LogoURL:    "Sony_MAX2.png",
	},
	{
		ID:          "sl1393",
		Name:        "SL: Sony WAH",
		Language:    1,
		Category:    5,
		IsHD:        false,
		LogoURL:    "Sony_Wah.png",
	},
	{
		ID:          "sl162",
		Name:        "SL: Sony TEN 1 HD",
		Language:    6,
		Category:    8,
		IsHD:        true,
		LogoURL:    "Ten_HD.png",
	},
	{
		ID:          "sl891",
		Name:        "SL: Sony TEN 2 HD",
		Language:    6,
		Category:    8,
		IsHD:        true,
		LogoURL:    "Ten2_HD.png",
	},
	{
		ID:          "sl892",
		Name:        "SL: Sony TEN 3 HD",
		Language:    1,
		Category:    8,
		IsHD:        true,
		LogoURL:    "Ten3_HD.png",
	},
	{
		ID:          "sl1772",
		Name:        "SL: Sony TEN 4 HD",
		Language:    8,
		Category:    8,
		IsHD:        true,
		LogoURL:    "Ten_4_HD_Tamil.png",
	},
	{
		ID:          "sl155",
		Name:        "SL: Sony TEN 5 HD",
		Language:    6,
		Category:    8,
		IsHD:        true,
		LogoURL:    "Six_HD.png",
	},
	{
		ID:          "sl852",
		Name:        "SL: Sony BBC Earth HD",
		Language:    6,
		Category:    10,
		IsHD:        true,
		LogoURL:    "Sony_BBC_Earth_HD_English.png",
	},
}
