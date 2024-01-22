package handlers

import (
	"fmt"
	"strconv"

	"github.com/rabilrbl/jiotv_go/v3/pkg/epg"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

const (
	EPG_POSTER_URL = "https://jiotv.catchup.cdn.jio.com/dare_images/shows/"
)

// WebEPGHandler responds to requests for EPG data for individual channels.
func WebEPGHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")

	if channelID[:2] == "sl" {
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

	url := fmt.Sprintf(epg.EPG_URL, offset, channelIntID)
	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}

	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

// PosterHandler loads image from JioTV server
func PosterHandler(c *fiber.Ctx) error {
	// catch all params
	url := EPG_POSTER_URL + c.Params("date") + "/" + c.Params("file")
	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}
