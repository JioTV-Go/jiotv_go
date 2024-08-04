package epg

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
)

// Channel XML tag structure for the EPG
type Channel struct {
	XMLName xml.Name `xml:"channel"`      // XML tag name
	ID      int      `xml:"id,attr"`      // ID is attribute of channel tag
	Display string   `xml:"display-name"` // Display name of the channel
}

// Icon XML tag for Programme XML tag in EPG
type Icon struct {
	XMLName xml.Name `xml:"icon"`     // XML tag name
	Src     string   `xml:"src,attr"` // Src is attribute of the icon tag
}

// Title XML tag for Programme XML tag in EPG
// Title is the name of the programme or show being aired on the channel
type Title struct {
	XMLName xml.Name `xml:"title"`
	Value   string   `xml:",chardata"` // Title of the programme
	Lang    string   `xml:"lang,attr"` // Language of the title
}

// Category XML tag for Programme XML tag in EPG
// Category is the type of the programme or show being aired on the channel
type Category struct {
	XMLName xml.Name `xml:"category"`
	Value   string   `xml:",chardata"` // Category of the programme
	Lang    string   `xml:"lang,attr"` // Language of the category
}

// Desc represents Description XML tag for Programme XML tag in EPG
type Desc struct {
	XMLName xml.Name `xml:"desc"`
	Value   string   `xml:",chardata"` // Description of the programme
	Lang    string   `xml:"lang,attr"` // Language of the description
}

// Programme XML tag structure for EPG
// Each programme tag represents a show being aired on a channel
type Programme struct {
	XMLName  xml.Name `xml:"programme"`    // XML tag name
	Channel  string   `xml:"channel,attr"` // Channel is attribute of programme tag
	Start    string   `xml:"start,attr"`   // Start time of the programme
	Stop     string   `xml:"stop,attr"`    // Stop time of the programme
	Title    Title    `xml:"title"`        // Title of the programme
	Desc     Desc     `xml:"desc"`         // Description of the programme
	Category Category `xml:"category"`     // Category of the programme
	Icon     Icon     `xml:"icon"`         // Icon of the programme
}

// EPG XML tag structure
type EPG struct {
	XMLName     xml.Name    `xml:"tv"`            // XML tag name
	XMLVersion  string      `xml:"version,attr"`  // XML version
	XMLEncoding string      `xml:"encoding,attr"` // XML encoding
	Channel     []Channel   `xml:"channel"`       // Channel tags
	Programme   []Programme `xml:"programme"`     // Programme tags
}

// ChannelObject represents Individual channel detail from JioTV API response
type ChannelObject struct {
	ChannelID   int    `json:"channel_id"`   // Channel ID
	ChannelName string `json:"channel_name"` // Channel name
	LogoURL     string `json:"logoUrl"`      // Channel logo URL
}

// ChannelsResponse represents Channel details from JioTV API response
type ChannelsResponse struct {
	Channels []ChannelObject `json:"result"`  // Channels
	Code     int             `json:"code"`    // Response code
	Message  string          `json:"message"` // Response message
}

// EPGObject represents Individual EPG detail from JioTV EPG API response
type EPGObject struct {
	StartEpoch   int64  `json:"startEpoch"`       // Start time of the programme
	EndEpoch     int64  `json:"endEpoch"`         // End time of the programme
	ChannelID    uint16 `json:"channel_id"`       // Channel ID
	ChannelName  string `json:"channel_name"`     // Channel name
	ShowCategory string `json:"showCategory"`     // Category of the show
	Description  string `json:"description"`      // Description of the show
	Title        string `json:"showname"`         // Title of the show
	Thumbnail    string `json:"episodeThumbnail"` // Thumbnail of the show
	Poster       string `json:"episodePoster"`    // Poster of the show
}

// EPGResponse represents EPG details from JioTV EPG API response
type EPGResponse struct {
	EPG []EPGObject `json:"epg"` // EPG details for a channel
}

// EpochString is a custom type for unmarshaling epoch from integers to strings in JioTV EPG API
type EpochString string

// UnmarshalJSON unmarshals epoch integers to strings from JioTV EPG API
func (id *EpochString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as integer
	var intValue int
	// If it fails, unmarshal as string
	if err := json.Unmarshal(data, &intValue); err != nil {
		var stringValue string
		if err := json.Unmarshal(data, &stringValue); err != nil {
			return err
		}
		*id = EpochString(stringValue)
	} else {
		// limit to 10 digits
		*id = EpochString(strconv.Itoa(intValue)[:10])
	}
	return nil
}

// String returns the string representation of the EpochString
func (id *EpochString) String() string {
	return string(*id)
}
