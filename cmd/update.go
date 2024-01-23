package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

func Update(currentVersion string) error {
	fmt.Println("Updating JioTV Go...")

	// Determine the architecture and operating system
	const arch = runtime.GOARCH
	const os_name = runtime.GOOS

	fmt.Println("System detected:", os_name, arch)

	// Fetch the latest release information from GitHub
	release, err := getLatestRelease()
	if err != nil {
		return err
	}

	latestVersion := release.TagName
	fmt.Printf("Latest version: %s\n", latestVersion)

	// Compare versions
	if compareVersions(currentVersion, latestVersion) >= 0 {
		fmt.Println("You are already using the latest version. No update needed.")
		return nil
	}

	fmt.Println("Newer version available. Updating...")

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
	// skipcq GSC-G302
	if err := os.Chmod(tempBinaryPath, 0600); err != nil {
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

// Asset Define the structures to assets from GitHub API
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release Define the structures to release from GitHub API
type Release struct {
	Assets  []Asset `json:"assets"`
	TagName string  `json:"tag_name"`
}

func downloadBinary(url, outputPath string) error {
	statusCode, body, err := fasthttp.Get(nil, url)
	if err != nil {
		return err
	}

	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("failed to download binary. Status code: %d", statusCode)
	}

	return os.WriteFile(outputPath, body, 0644)
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

func compareVersions(currentVersion, latestVersion string) int {
	// Implement version comparison logic based on your versioning scheme
	// This is a simplified example assuming semantic versioning (major.minor.patch)

	// Split versions into components
	currentComponents := strings.Split(currentVersion, ".")
	latestComponents := strings.Split(latestVersion, ".")

	// Compare major version
	currentMajor := atoiOrZero(currentComponents[0])
	latestMajor := atoiOrZero(latestComponents[0])
	if currentMajor != latestMajor {
		return currentMajor - latestMajor
	}

	// Compare minor version
	currentMinor := atoiOrZero(currentComponents[1])
	latestMinor := atoiOrZero(latestComponents[1])
	if currentMinor != latestMinor {
		return currentMinor - latestMinor
	}

	// Compare patch version
	currentPatch := atoiOrZero(currentComponents[2])
	latestPatch := atoiOrZero(latestComponents[2])
	return currentPatch - latestPatch
}

func atoiOrZero(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
