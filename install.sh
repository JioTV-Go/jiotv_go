#!/bin/bash

# Source: https://github.com/rabilrbl/jiotv_go

# Define the GitHub repository and binary information
REPO_OWNER="rabilrbl"
REPO_NAME="jiotv_go"
BINARY_NAME="jiotv_go"

# Determine the device's architecture
ARCH=$(uname -m)
case $ARCH in
"aarch64")
    ARCH="arm64"
    ;;
"armv7l")
    ARCH="arm"
    ;;
"armv8l")
    ARCH="arm"
    ;;
"x86_64")
    ARCH="amd64"
    ;;
*)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

usage() {
    echo "Usage: $0 {install|update|run}"
    echo "  install: Install JioTV Go for the first time"
    echo "  update: Update JioTV Go to the latest version"
    echo "  run: Run JioTV Go (default)"
    echo ""
    echo "You can optionally specify the \"host:port\" to run JioTV Go as a second argument."
}

release_file_name() {
    # Check if curl is installed
    if ! command -v curl >/dev/null; then
        echo "Error: curl not found. Please install curl."
        return 1
    fi

    # Get the latest release version from GitHub API
    echo "Fetching latest release info from GitHub..."
    RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
    LATEST_VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": "[^"]*' | sed 's/"tag_name": "//')

    file_name="$BINARY_NAME-$LATEST_VERSION-$OS-$ARCH"
}

download_binary() {
    # fetch latest file name
    if ! release_file_name; then
        echo "Failed to fetch latest release info from GitHub."
        return 1
    fi

    # Construct the download URL for the binary
    DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/$file_name"

    # Download and install the binary
    echo "Downloading JioTV Go $LATEST_VERSION for $ARCH..."
    if ! curl -LO --progress-bar --retry 5 --retry-delay 2 "$DOWNLOAD_URL"; then
        echo "Error: Failed to download binary. Please try again later."
        return 1
    fi
    if ! chmod +x "$file_name"; then
        echo "Error: Failed to make binary executable. File missing? Run again!."
        return 1
    fi
}

# Function to install the binary
install_linux() {

    if download_binary; then
        echo "JioTV Go $LATEST_VERSION for $ARCH has been downloaded."
        echo "Execute \"$0 run\" to start JioTV Go."
    else
        echo "Failed to download JioTV Go"
        return 1
    fi
}

# Function to update the binary
update_linux() {
    # fetch existing file name, if multiple files are present, pick the latest one
    existing_file_name=$(ls $BINARY_NAME-*-$ARCH | sort -r | head -n 1) # skipcq: SH-2012

    if [ ! -z "$existing_file_name" ]; then
        echo "Found existing file: $existing_file_name"
        # fetch latest file name
        if ! release_file_name; then
            echo "Failed to fetch latest release info from GitHub."
            return 1
        fi
        # compare existing file name with file name
        # if both are same, need not download again
        if [ "$existing_file_name" == "$file_name" ]; then
            echo "JioTV Go is already up to date."
            return 0
        else
            if download_binary; then
                # Delete the old binary from ls command
                if [ ! -z "$existing_file_name" ]; then
                    echo "Deleting old version $existing_file_name..."
                    rm "$existing_file_name"
                fi
                echo "JioTV Go has been updated to latest $LATEST_VERSION version."
            else
                echo "Failed to update JioTV Go"
            fi
        fi
    else
        echo "Missing existing file name. Is JioTV Go installed?"
        echo "Execute \"$0 install\" to install JioTV Go."
    fi
}

# Function to run the binary
run_linux() {
    # fetch file name from ls command, if multiple files are present, pick the latest one
    file_name=$(ls $BINARY_NAME-*-$ARCH | sort -r | head -n 1) # skipcq: SH-2012

    # Check if the binary exists
    if [ -z "$file_name" ]; then
        echo "Error: Binary '$BINARY_NAME' not found in the current directory."
        echo "Execute \"$0 install\" to install JioTV Go."
        return 1
    fi

    # Add optional second argument as the address to run the binary on
    JIOTV_GO_ADDR="localhost:5001"
    if [ ! -z "$2" ]; then
        JIOTV_GO_ADDR="$2"
    fi

    # Run the Android binary
    echo "Running JioTV Go at $JIOTV_GO_ADDR for $ARCH..."

    /bin/bash -c "./$file_name $JIOTV_GO_ADDR"
}

# Check for the provided argument and perform the corresponding action
case "$1" in
"install")
    install_linux "$@"
    ;;
"update")
    update_linux "$@"
    ;;
"run")
    run_linux "$@"
    ;;
"help")
    usage
    ;;
*)
    usage
    exit 1
    ;;
esac
