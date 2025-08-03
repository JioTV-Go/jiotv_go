package handlers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	"github.com/jiotv-go/jiotv_go/v3/pkg/epg"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

const (
	EPG_POSTER_URL = urls.EPGPosterURLSlash
)

// EPGHandler handles EPG requests
func EPGHandler(c *fiber.Ctx) error {
	epgFilePath := utils.GetPathPrefix() + "epg.xml.gz"
	// if epg.xml.gz exists, return it
	if _, err := os.Stat(epgFilePath); err == nil {
		return c.SendFile(epgFilePath, true)
	} else {
		err_message := "EPG not found. Please restart the server after setting the environment variable JIOTV_EPG to true."
		utils.Log.Println(err_message) // Changed from fmt.Println
		return c.Status(fiber.StatusNotFound).SendString(err_message)
	}
}

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
