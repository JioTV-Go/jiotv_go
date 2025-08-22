package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/valyala/fasthttp"
)

// HTTPRequestConfig holds configuration for making HTTP requests
type HTTPRequestConfig struct {
	URL         string
	Method      string
	Body        []byte
	Headers     map[string]string
	UserAgent   string
	ContentType string
}

// MakeHTTPRequest creates and executes a fasthttp request with common patterns
func MakeHTTPRequest(config HTTPRequestConfig, client *fasthttp.Client) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(config.URL)
	req.Header.SetMethod(config.Method)

	// Set User-Agent
	if config.UserAgent != "" {
		req.Header.SetUserAgent(config.UserAgent)
	} else {
		req.Header.SetUserAgent(headers.UserAgentOkHttp)
	}

	// Set Content-Type
	if config.ContentType != "" {
		req.Header.SetContentType(config.ContentType)
	}

	// Set custom headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Set body if provided
	if len(config.Body) > 0 {
		req.SetBody(config.Body)
	}

	resp := fasthttp.AcquireResponse()
	
	// Perform the HTTP request
	if err := client.Do(req, resp); err != nil {
		fasthttp.ReleaseResponse(resp)
		return nil, err
	}

	return resp, nil
}

// MakeJSONRequest is a convenience function for making JSON requests
func MakeJSONRequest(url, method string, payload interface{}, requestHeaders map[string]string, client *fasthttp.Client) (*fasthttp.Response, error) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
		}
	}

	config := HTTPRequestConfig{
		URL:         url,
		Method:      method,
		Body:        body,
		Headers:     requestHeaders,
		ContentType: headers.ContentTypeJSON,
	}

	return MakeHTTPRequest(config, client)
}

// BatchStoreOperations holds multiple store operations
type BatchStoreOperations struct {
	Sets    map[string]string
	Deletes []string
}

// ExecuteBatchStoreOperations executes multiple store operations in sequence
func ExecuteBatchStoreOperations(ops BatchStoreOperations) error {
	// Execute all Set operations
	for key, value := range ops.Sets {
		if err := store.Set(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	// Execute all Delete operations
	for _, key := range ops.Deletes {
		if err := store.Delete(key); err != nil {
			return fmt.Errorf("failed to delete %s: %w", key, err)
		}
	}

	return nil
}

// FileOperationResult holds the result of file operations
type FileOperationResult struct {
	Exists bool
	Data   []byte
	Error  error
}

// CheckAndReadFile checks if a file exists and reads it if it does
func CheckAndReadFile(filePath string) FileOperationResult {
	result := FileOperationResult{}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result.Exists = false
		return result
	}

	result.Exists = true

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file %s: %w", filePath, err)
		return result
	}

	result.Data = data
	return result
}

// SetCommonJioTVHeaders sets common headers used across JioTV API requests
func SetCommonJioTVHeaders(req *fasthttp.Request, deviceID, crmID, uniqueID string) {
	req.Header.Set("appkey", "NzNiMDhlYzQyNjJm")
	req.Header.Set("channel_id", "")
	req.Header.Set("crmid", crmID)
	req.Header.Set("userId", crmID)
	req.Header.Set("deviceId", deviceID)
	req.Header.Set("devicetype", "phone")
	req.Header.Set("isott", "false")
	req.Header.Set("languageId", "6")
	req.Header.Set("lbcookie", "1")
	req.Header.Set("os", "android")
	req.Header.Set("osVersion", "13")
	req.Header.Set("subscriberId", crmID)
	req.Header.Set("uniqueId", uniqueID)
	req.Header.SetUserAgent(headers.UserAgentOkHttp)
	req.Header.Set("usergroup", "tvYR7NSNn7rymo3F")
	req.Header.Set("versionCode", headers.VersionCode389)
}

// ParseJSONResponse parses JSON response body into the provided interface
func ParseJSONResponse(resp *fasthttp.Response, target interface{}) error {
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("request failed with status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	if err := json.Unmarshal(resp.Body(), target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}

// LogAndReturnError logs an error and returns it (utility for consistent error handling)
func LogAndReturnError(err error, context string) error {
	SafeLogf("%s: %v", context, err)
	return fmt.Errorf("%s: %w", context, err)
}

// SafeLogf safely logs a formatted message, handling nil logger cases
func SafeLogf(format string, args ...interface{}) {
	if Log != nil {
		Log.Printf(format, args...)
	}
}

// SafeLog safely logs a message, handling nil logger cases
func SafeLog(message string) {
	if Log != nil {
		Log.Println(message)
	}
}