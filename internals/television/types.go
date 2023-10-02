package television

import (
	"github.com/valyala/fasthttp"
)

type Television struct {
	accessToken string
	ssoToken    string
	crm         string
	uniqueID    string
	headers     map[string]string
	Client      *fasthttp.Client
}

type Channel struct {
	ID       int    `json:"channel_id"`
	Name     string `json:"channel_name"`
	URL      string `json:"channel_url"`
	LogoURL  string `json:"logoUrl"`
	Category int    `json:"channelCategoryId"`
	Language int    `json:"channelLanguageId"`
	IsHD     bool   `json:"isHD"`
}

type APIResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Result  []Channel `json:"result"`
}

type Bitrates struct {
	Auto   string `json:"auto"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Medium string `json:"medium"`
}

type LiveURLOutput struct {
	Bitrates Bitrates `json:"bitrates"`
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Result   string   `json:"result"`
}

var CategoryMap = map[int]string{
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

var LanguageMap = map[int]string{
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
