package epg

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
)

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	ID      int      `xml:"id,attr"`
	Display string   `xml:"display-name"`
}

type Icon struct {
	XMLName xml.Name `xml:"icon"`
	Src     string   `xml:"src,attr"`
}

type Title struct {
	XMLName xml.Name `xml:"title"`
	Value   string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr"`
}

type Desc struct {
	XMLName xml.Name `xml:"desc"`
	Value   string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr"`
}

type Programme struct {
	XMLName xml.Name `xml:"programme"`
	Channel string   `xml:"channel,attr"`
	Start   string   `xml:"start,attr"`
	Stop    string   `xml:"stop,attr"`
	Title   Title    `xml:"title"`
	Desc    Desc     `xml:"desc"`
	Icon    Icon     `xml:"icon"`
}

type EPG struct {
	XMLName     xml.Name    `xml:"tv"`
	XMLVersion  string      `xml:"version,attr"`
	XMLEncoding string      `xml:"encoding,attr"`
	Channel     []Channel   `xml:"channel"`
	Programme   []Programme `xml:"programme"`
}

type ChannelObject struct {
	ChannelID   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	LogoURL     string `json:"logoUrl"`
}

type ChannelsResponse struct {
	Channels []ChannelObject `json:"result"`
	Code     int             `json:"code"`
	Message  string          `json:"message"`
}

type EPGObject struct {
	StartEpoch   EpochString `json:"startEpoch"`
	EndEpoch     EpochString `json:"endEpoch"`
	ChannelID    uint16      `json:"channel_id"`
	ChannelName  string      `json:"channel_name"`
	ShowCategory string      `json:"showCategory"`
	Description  string      `json:"description"`
	Title        string      `json:"showname"`
	Thumbnail    string      `json:"episodeThumbnail"`
	Poster       string      `json:"episodePoster"`
}

type EPGResponse struct {
	EPG []EPGObject `json:"epg"`
}

type EpochString string

func (id *EpochString) UnmarshalJSON(data []byte) error {
	var intValue int
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

func (id *EpochString) String() string {
	return string(*id)
}
