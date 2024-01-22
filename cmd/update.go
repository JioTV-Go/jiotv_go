package cmd

import (
	"os"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/valyala/fasthttp"
)

func Update() error {
	fmt.Println("Updating JioTV Go...")

	// Determine the architecture and operating system
	arch := runtime.GOARCH
	os_name := runtime.GOOS

	fmt.Println("System detected:", os_name, arch)

	// Fetch the latest release information from GitHub
	release, err := getLatestRelease()
	if err != nil {
		return err
	}

	// Choose the appropriate asset based on os and arch
	assetName := fmt.Sprintf("jiotv_go-%s-%s", os_name, arch)
	if os_name == "windows" {
		assetName += ".exe"
	}

	// Find the asset with the chosen name
	var assetURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no suitable release asset found for %s", assetName)
	}

	// Download the new binary to a temporary location
	tempBinaryPath := "jiotv_go_temp"
	if err := downloadBinary(assetURL, tempBinaryPath); err != nil {
		return err
	}

	// Make the binary executable
	if err := os.Chmod(tempBinaryPath, 0755); err != nil {
		return err
	}

	// Replace the old binary with the new one
	if err := replaceBinary(tempBinaryPath); err != nil {
		return err
	}

	fmt.Println("Update successful. Restart the application to apply changes.")

	return nil
}

func getLatestRelease() (*Release, error) {
	owner := "rabilrbl"
	repo := "jiotv_go"

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release. Status code: %d", resp.StatusCode())
	}

	body := resp.Body()
	var release Release
	err := json.Unmarshal(body, &release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

// Define the structures to parse the JSON response
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	Assets []Asset `json:"assets"`
}

func downloadBinary(url, outputPath string) error {
	statusCode, body, err := fasthttp.Get(nil, url)
	if err != nil {
		return err
	}

	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("failed to download binary. Status code: %d", statusCode)
	}

	if err := os.WriteFile(outputPath, body, 0644); err != nil {
		return err
	}

	return nil
}

func replaceBinary(newBinaryPath string) error {
	currentBinaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Rename the old binary
	oldBinaryPath := currentBinaryPath + ".old"
	if err := os.Rename(currentBinaryPath, oldBinaryPath); err != nil {
		return err
	}

	// Rename the new binary to the original binary name
	if err := os.Rename(newBinaryPath, currentBinaryPath); err != nil {
		// If there is an error, attempt to roll back to the old binary
		os.Rename(oldBinaryPath, currentBinaryPath)
		return err
	}

	// Remove the old binary
	os.Remove(oldBinaryPath)

	return nil
}
