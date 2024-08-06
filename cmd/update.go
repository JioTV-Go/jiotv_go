package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/valyala/fasthttp"
)

// Update checks for a newer version of the application
// by calling getLatestRelease() to fetch the latest release
// information from GitHub.

// It compares the latest version to the provided currentVersion
// using compareVersions(). If currentVersion is already
// up-to-date, it returns without updating.

// Otherwise, it finds the appropriate updated binary asset
// for the current OS and architecture, downloads it to a
// temporary location using downloadBinary(), and replaces
// the current binary using replaceBinary().

// Finally, it prints a message that the update was successful
// and the app should be restarted.
func Update(currentVersion, customVersion string) error {
	fmt.Println("Updating JioTV Go...")

	// Determine the architecture and operating system
	const arch = runtime.GOARCH
	const os_name = runtime.GOOS

	fmt.Println("System detected:", os_name, arch)

	// Fetch the latest release information from GitHub
	release, err := getLatestRelease(customVersion)
	if err != nil {
		return err
	}

	latestVersion := release.TagName

	// Compare versions
	if customVersion == "" && compareVersions(currentVersion, latestVersion) >= 0 {
		fmt.Println("You are already using the latest version. No update needed.")
		return nil
	}

	if customVersion == "" {
		fmt.Println("Newer version available. Updating to", latestVersion, "...")
	} else {
		fmt.Println("Updating to custom version", customVersion)
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

	// Replace the old binary with the new one
	if err := replaceBinary(tempBinaryPath); err != nil {
		return err
	}

	fmt.Println("Update successful. Restart the application to apply changes.")

	return nil
}

// getLatestRelease fetches the latest release information from the GitHub API for the given owner and repo.
// It returns a Release struct containing the release details like tag name, assets etc.
func getLatestRelease(customVersion string) (*Release, error) {
	owner := "rabilrbl"
	repo := "jiotv_go"

	var url string
	if customVersion != "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, customVersion)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	}

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

// Asset defines the structure of an asset
// from the GitHub API release response. It contains
// the asset name and browser download URL.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release represents a GitHub release. It contains the release
// tag name and associated assets.
type Release struct {
	Assets  []Asset `json:"assets"`
	TagName string  `json:"tag_name"`
}

// downloadBinary downloads a binary file from the given URL
// and saves it to the specified output path.
// It returns an error if the request fails or the status code is not 200 OK.
// The saved binary file is made executable.
func downloadBinary(url, outputPath string) error {
	initialBufferSize := 8192
	maxBufferSize := 32768

	// Iterate through buffer sizes, starting from initialBufferSize and doubling each time
	for bufferSize := initialBufferSize; bufferSize <= maxBufferSize; bufferSize *= 2 {
		client := &fasthttp.Client{
			ReadBufferSize: bufferSize, // Set the read buffer size for the client
		}

		// Perform an HTTP GET request
		statusCode, body, err := client.Get(nil, url)
		if err != nil {
			// Check if the error is due to a small read buffer and if we can increase the buffer size
			if strings.Contains(err.Error(), "small read buffer") && bufferSize < maxBufferSize {
				fmt.Println("Increasing buffer size and retrying...")
				continue // Retry with a larger buffer size
			}
			return err // Return the error if it's not related to buffer size or max buffer size is reached
		}

		if statusCode != fasthttp.StatusOK {
			return fmt.Errorf("failed to download binary. Status code: %d", statusCode)
		}

		// Write the downloaded binary to the specified output path with executable permissions
		// skipcq: GSC-G302 - We want executable permissions on the binary
		return os.WriteFile(outputPath, body, 0744)
	}

	// Return an error if the binary could not be downloaded after increasing the buffer size
	return fmt.Errorf("failed to download binary after increasing buffer size")
}

// replaceBinary replaces the current executable binary with a new binary.
// It renames the current binary to a .old file, copies the new binary to the current
// binary path, and attempts to roll back if there are any errors. Finally it removes
// the old .old binary.
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

// compareVersions compares two semantic version strings and returns an integer
// indicating whether the current version is less than, equal to, or greater
// than the latest version.

// It splits the version strings into major, minor and patch numeric components
// and compares them sequentially. The first non-equal set of components
// determines the return value. Returns -1 if currentVersion < latestVersion,
// 0 if currentVersion == latestVersion, and 1 if currentVersion > latestVersion.
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

// atoiOrZero converts a string to an integer, returning 0 if the
// conversion fails.
func atoiOrZero(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// IsUpdateAvailable checks if a newer version of the application is available
// by calling getLatestRelease() to fetch the latest release
// information from GitHub.
func IsUpdateAvailable(currentVersion, customVersion string) string {
	release, err := getLatestRelease(customVersion)
	if err != nil {
		return ""
	}

	latestVersion := release.TagName

	// Compare versions
	if customVersion == "" && compareVersions(currentVersion, latestVersion) >= 0 {
		return ""
	}

	return latestVersion
}

// PrintIfUpdateAvailable checks if a newer version of the application is available
func PrintIfUpdateAvailable(c *cli.Context) {
	isUpdateAvailableVersion := IsUpdateAvailable(c.App.Version, "")
	if isUpdateAvailableVersion != "" {
		fmt.Printf("Newer version %s available. Run `jiotv_go update` to update.\n", isUpdateAvailableVersion)
	}
}
