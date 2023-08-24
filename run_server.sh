#!/bin/bash

ARCH=$(uname -m)  # Automatically detect system architecture
if [ "$ARCH" != "amd64" ] && [ "$ARCH" != "arm64" ]; then
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Function to get the latest Go version from the website
get_latest_go_version() {
    local latest_version_url="https://golang.org/dl/?mode=json"
    local latest_version=$(curl -s $latest_version_url | grep -o '"version":"[^"]*' | cut -d'"' -f4)
    echo $latest_version
}

# Check if 'go' is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing..."

    # Get the latest Go version
    latest_version=$(get_latest_go_version)
    echo "Latest Go version: $latest_version"

    # Download and install Go
    wget -q https://dl.google.com/go/$latest_version.linux-$ARCH.tar.gz
    tar -C /usr/local -xzf $latest_version.linux-$ARCH.tar.gz
    rm $latest_version.linux-$ARCH.tar.gz

    # Add Go binary directory to PATH
    export PATH=$PATH:/usr/local/go/bin
fi

# Run the Go program
echo "Running Go program..."
go mod tidy
go run .

echo "Go program completed."
