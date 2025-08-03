package television

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/jiotv-go/jiotv_go/v3/internal/constants"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// generateDateTime generates a timestamp in the format required by DRM requests
func generateDateTime() string {
	currentTime := time.Now()
	formattedDateTime := fmt.Sprintf("%02d%02d%02d%02d%02d%03d",
		currentTime.Year()%100, currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(),
		currentTime.Nanosecond()/1000000)
	return formattedDateTime
}

// RequestDRMKey makes HTTP requests for DRM key with appropriate headers
func (tv *Television) RequestDRMKey(url, channelID string) ([]byte, int, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(constants.MethodPOST)

	// Set DRM-specific headers
	req.Header.Set(headers.AccessToken, tv.AccessToken)
	req.Header.Set(headers.Connection, headers.ConnectionKeepAlive)
	req.Header.Set(headers.OS, headers.OSAndroid)
	req.Header.Set(headers.AppName, headers.AppNameJioTV)
	req.Header.Set(headers.SubscriberID, tv.Crm)
	req.Header.Set(headers.UserAgent, headers.UserAgentPlayTVNew)
	req.Header.Set(headers.SsoToken, tv.SsoToken)
	req.Header.Set(headers.XPlatform, headers.OSAndroid)
	req.Header.Set(headers.SerialNumber, generateDateTime())
	req.Header.Set(headers.CRMID, tv.Crm)
	req.Header.Set(headers.ChannelID, channelID)
	req.Header.Set(headers.UniqueID, tv.UniqueID)
	req.Header.Set(headers.VersionCode, headers.VersionCode330)
	req.Header.Set(headers.UserGroup, headers.UserGroupDefault)
	req.Header.Set(headers.DeviceType, headers.DeviceTypePhone)
	req.Header.Set(headers.AcceptEncoding, headers.AcceptEncodingGzipDeflate)
	req.Header.Set(headers.OSVersion, headers.OSVersion13)
	req.Header.Set(headers.DeviceID, utils.GetDeviceID())
	req.Header.Set(headers.ContentType, headers.ContentTypeOctetStream)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP request
	if err := tv.Client.Do(req, resp); err != nil {
		return nil, 0, err
	}

	// Copy response body before releasing
	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return body, resp.StatusCode(), nil
}

// RequestMPD makes HTTP requests for MPD manifest files
func (tv *Television) RequestMPD(url string) ([]byte, int, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(constants.MethodGET)
	req.Header.Set(headers.UserAgent, headers.UserAgentPlayTVNew)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP request
	if err := tv.Client.Do(req, resp); err != nil {
		return nil, 0, err
	}

	// Copy response body before releasing
	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return body, resp.StatusCode(), nil
}

// RequestDashSegment makes HTTP requests for DASH segments
func (tv *Television) RequestDashSegment(url string) ([]byte, int, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(constants.MethodGET)
	req.Header.Set(headers.UserAgent, headers.UserAgentPlayTVNew)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP request
	if err := tv.Client.Do(req, resp); err != nil {
		return nil, 0, err
	}

	// Copy response body before releasing
	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return body, resp.StatusCode(), nil
}
