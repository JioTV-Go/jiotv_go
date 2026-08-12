package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	internalUtils "github.com/jiotv-go/jiotv_go/v3/internal/utils"
	"github.com/jiotv-go/jiotv_go/v3/pkg/epg"
	pkgUtils "github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

const (
	EPG_POSTER_URL = urls.EPGPosterURLSlash
	// The getepg API stamps serverDate with IST (UTC+05:30) day boundaries;
	// real-day indices are computed in that same timezone.
	istOffsetMS = 5.5 * 60 * 60 * 1000
	dayMS       = 24 * 60 * 60 * 1000
)

// webEPGURLFormat is the upstream getepg URL pattern. It is a variable rather
// than epg.EPG_URL directly so tests can point it at a local server.
var webEPGURLFormat = epg.EPG_URL

// EPGHandler handles EPG requests
func EPGHandler(c *fiber.Ctx) error {
	epgFilePath := pkgUtils.GetPathPrefix() + "epg.xml.gz"
	// if epg.xml.gz exists, return it
	if _, err := os.Stat(epgFilePath); err == nil {
		return c.SendFile(epgFilePath, true)
	} else {
		err_message := "EPG not found. Please restart the server after setting the environment variable JIOTV_EPG to true."
		pkgUtils.Log.Println(err_message) // Changed from fmt.Println
		return internalUtils.NotFoundError(c, err_message)
	}
}

// WebEPGHandler responds to requests for EPG data for individual channels.
func WebEPGHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")

	if strings.HasPrefix(channelID, "sl") {
		channelID = channelID[2:]
	}

	channelIntID, err := strconv.Atoi(channelID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid channel ID")
	}

	// Get offset from URL
	offset, err := strconv.Atoi(c.Params("offset"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid offset")
	}

	body, contentType, statusCode, err := webEPGWithCorrectedDay(channelIntID, offset)
	if err != nil {
		return err
	}

	c.Status(statusCode)
	if contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	} else {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return c.Send(body)
}

// webEPGWithCorrectedDay fetches the schedule for the requested day and, if
// the response's serverDate shows the API's "today" lags real time, re-fetches
// with the corrected offset so the returned schedule covers the present moment
// for any client. It falls back to the original response if the corrected
// fetch fails.
func webEPGWithCorrectedDay(channelID, offset int) ([]byte, string, int, error) {
	body, contentType, statusCode, err := fetchWebEPG(channelID, offset)
	if err != nil {
		return nil, "", 0, err
	}
	if statusCode == fiber.StatusOK {
		if dayOffset := webEPGDayOffset(body); dayOffset > 0 {
			if adjusted, adjType, adjStatus, adjErr := fetchWebEPG(channelID, offset+dayOffset); adjErr == nil && adjStatus == fiber.StatusOK {
				return adjusted, adjType, adjStatus, nil
			}
		}
	}
	return body, contentType, statusCode, nil
}

// fetchWebEPG fetches the EPG schedule for a channel from the JioTV API and
// returns the response body, content type, and status code.
func fetchWebEPG(channelID, offset int) ([]byte, string, int, error) {
	url := fmt.Sprintf(webEPGURLFormat, offset, channelID)
	resp, err := pkgUtils.MakeHTTPRequest(pkgUtils.HTTPRequestConfig{
		URL:    url,
		Method: "GET",
	}, TV.Client)
	if err != nil {
		return nil, "", 0, err
	}
	defer fasthttp.ReleaseResponse(resp)

	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())
	return body, string(resp.Header.ContentType()), resp.StatusCode(), nil
}

// webEPGDayOffset returns how many days the EPG API's "today" lags the real
// date, derived from the serverDate the API stamps on its response. It returns
// 0 when the offset cannot be determined, meaning nothing to correct.
func webEPGDayOffset(body []byte) int {
	var resp struct {
		EPG []struct {
			ServerDate string `json:"serverDate"`
		} `json:"epg"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || len(resp.EPG) == 0 {
		return 0
	}
	if resp.EPG[0].ServerDate == "" {
		return 0
	}
	serverDayStart, err := time.Parse(time.RFC3339Nano, resp.EPG[0].ServerDate)
	if err != nil {
		return 0
	}
	epgDay := (serverDayStart.UnixMilli() + istOffsetMS) / dayMS
	today := (time.Now().UnixMilli() + istOffsetMS) / dayMS
	if delta := int(today - epgDay); delta > 0 {
		return delta
	}
	return 0
}

// PosterHandler loads image from JioTV server
func PosterHandler(c *fiber.Ctx) error {
	// catch all params
	url := EPG_POSTER_URL + c.Params("date") + "/" + c.Params("file")
	return internalUtils.ProxyRequest(c, url, TV.Client, "")
}
