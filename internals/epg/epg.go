package epg

import (
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"

	"os"
	"sync"
	"time"

	"github.com/rabilrbl/jiotv_go/internals/utils"
	"github.com/valyala/fasthttp"
)

const (
	CHANNEL_URL = "https://jiotv.data.cdn.jio.com/apis/v3.0/getMobileChannelList/get/?os=android&devicetype=phone&usertype=tvYR7NSNn7rymo3F"
	EPG_URL     = "https://jiotv.data.cdn.jio.com/apis/v1.3/getepg/get/?offset=%d&channel_id=%d"
)

func Init() {
	epgFile := "epg.xml.gz"
	var lastModTime time.Time
	flag := false
	utils.Log.Println("Checking EPG file")
	if stat, err := os.Stat(epgFile); err == nil {
		lastModTime = stat.ModTime()
		if time.Since(lastModTime).Hours() < 24 {
			utils.Log.Println("EPG file is up to date")
		} else {
			utils.Log.Println("EPG file is older than 24 hours")
			flag = true
		}
	} else {
		utils.Log.Println("EPG file doesn't exist")
		flag = true
	}

	genepg := func() {
		// schedule to run at 5:30 AM IST
		fmt.Println("Generating new EPG file... Please wait, this may take a while")
		if err := GenXMLGz(epgFile); err != nil {
			utils.Log.Fatal(err)
		}
	}

	if flag {
		genepg()
	}
	time_now := time.Now()
	schedule_time := time.Date(time_now.Year(), time_now.Month(), time_now.Day()+1, -5, -30, 0, 0, time.UTC)
	utils.Log.Println("Scheduled EPG generation in", time.Until(schedule_time).Truncate(time.Second))
	go utils.ScheduleFunctionCall(genepg, schedule_time)
}

func NewChannel(id int, displayName string) Channel {
	return Channel{
		ID:      id,
		Display: displayName,
	}
}

func NewProgramme(channelID int, start, stop, title, desc, iconSrc string) Programme {
	iconURL := fmt.Sprintf("https://jiotv.catchup.cdn.jio.com/dare_images/shows/%s", iconSrc)
	return Programme{
		Channel: fmt.Sprint(channelID),
		Start:   start,
		Stop:    stop,
		Title: Title{
			Value: title,
			Lang:  "en",
		},
		Desc: Desc{
			Value: desc,
			Lang:  "en",
		},
		Icon: Icon{
			Src: iconURL,
		},
	}
}

func NewEPG(channels []Channel, programmes []Programme) EPG {
	return EPG{
		Channel:   channels,
		Programme: programmes,
	}
}

func genXML() ([]byte, error) {
	// Create a reusable fasthttp client with common headers
	client := utils.GetRequestClient()

	// Create channels and programmes slices with initial capacity
	var channels []Channel
	var programmes []Programme

	// Define a worker function for fetching EPG data
	fetchEPG := func(channel Channel) {
		req := fasthttp.AcquireRequest()
		req.Header.SetUserAgent("okhttp/4.2.2")
		defer fasthttp.ReleaseRequest(req)

		resp := fasthttp.AcquireResponse()

		for offset := -1; offset < 2; offset++ {
			reqUrl := fmt.Sprintf(EPG_URL, offset, channel.ID)
			req.SetRequestURI(reqUrl)

			if err := client.Do(req, resp); err != nil {
				// Handle error
				utils.Log.Printf("Error fetching EPG for channel %d, offset %d: %v", channel.ID, offset, err)
				continue
			}

			var epgResponse EPGResponse
			if err := json.Unmarshal(resp.Body(), &epgResponse); err != nil {
				// Handle error
				utils.Log.Printf("Error unmarshaling EPG response for channel %d, offset %d: %v", channel.ID, offset, err)
				continue
			}

			for _, programme := range epgResponse.EPG {
				start, err := strconv.ParseInt(programme.StartEpoch.String(), 10, 64)
				if err != nil {
					utils.Log.Printf("Error parsing start epoch for channel %d, offset %d: %v", channel.ID, offset, err)
					continue
				}
				end, err := strconv.ParseInt(programme.EndEpoch.String(), 10, 64)
				if err != nil {
					utils.Log.Printf("Error parsing end epoch for channel %d, offset %d: %v", channel.ID, offset, err)
					continue
				}
				startTime := formatTime(time.Unix(start, 0))
				endTime := formatTime(time.Unix(end, 0))
				programmes = append(programmes, NewProgramme(channel.ID, startTime, endTime, programme.Title, programme.Description, programme.Poster))
			}
		}

		fasthttp.ReleaseResponse(resp)
	}

	// Fetch channels data
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(CHANNEL_URL)
	resp := fasthttp.AcquireResponse()

	utils.Log.Println("Fetching channels")
	if err := client.Do(req, resp); err != nil {
		utils.Log.Fatal(err)
		return nil, err
	}
	defer fasthttp.ReleaseResponse(resp)

	var channelsResponse ChannelsResponse
	if err := json.Unmarshal(resp.Body(), &channelsResponse); err != nil {
		utils.Log.Fatal(err)
		return nil, err
	}

	for _, channel := range channelsResponse.Channels {
		channels = append(channels, NewChannel(channel.ChannelID, channel.ChannelName))
	}
	utils.Log.Println("Fetched", len(channels), "channels")
	// Use a worker pool to fetch EPG data concurrently
	const numWorkers = 20 // Adjust the number of workers based on your needs
	channelQueue := make(chan Channel, len(channels))
	var wg sync.WaitGroup

	utils.Log.Println("Fetching EPG for channels")
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for channel := range channelQueue {
				// utils.Log.Println("Fetching EPG for channel", channel.ID)
				fetchEPG(channel)
			}
		}()
	}
	// Queue channels for processing
	for _, channel := range channels {
		channelQueue <- channel
	}
	close(channelQueue)
	wg.Wait()

	utils.Log.Println("Fetched programmes")
	// Create EPG and marshal it to XML
	epg := NewEPG(channels, programmes)
	xml, err := xml.MarshalIndent(epg, "", "  ")
	if err != nil {
		return nil, err
	}
	return xml, nil
}

func formatTime(t time.Time) string {
	return t.Format("20060102150405 -0700")
}

func GenXMLGz(filename string) error {
	utils.Log.Println("Generating XML")
	xml, err := genXML()
	if err != nil {
		return err
	}
	// write to file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close() // skipcq: GO-S2307

	utils.Log.Println("Writing XML to gzip file")
	gz := gzip.NewWriter(f)
	defer gz.Close()

	if _, err := gz.Write(xml); err != nil {
		return err
	}
	return nil
}
