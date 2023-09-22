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

usage() {
    echo "Usage: $0 {install|update|run|auto-start|help}"
    echo "  install: Install JioTV Go for the first time"
    echo "  update: Update JioTV Go to the latest version"
    echo "  run: Run JioTV Go (default)"
    echo "  auto-start: Add/Remove auto start. To launch JioTV Go automatically on Termux startup"
    echo ""
    echo "You can optionally specify the \"host:port\" to run JioTV Go as a second argument."
}

release_file_name() {
    # Get the latest release version from GitHub API
    echo "Fetching latest release info from GitHub..."
    RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
    LATEST_VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": "[^"]*' | sed 's/"tag_name": "//')

    file_name="$BINARY_NAME-$LATEST_VERSION-linux-$ARCH"
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
install_android() {
    echo "Upgrading termux repositories..."
    echo "Please be patient, this can take a few minutes to complete..."
    DEBIAN_FRONTEND=noninteractive pkg update -y && pkg upgrade -y
    echo "Installing curl, openssl and proot from Termux repositories..."
    pkg install curl openssl proot -y

    if download_binary; then
        echo "JioTV Go $LATEST_VERSION for $ARCH has been downloaded."
        echo "Execute \"$0 run\" to start JioTV Go."
    else
        echo "Failed to download JioTV Go"
        return 1
    fi
}

# Function to update the binary
update_android() {
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
run_android() {
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

    proot -b "$PREFIX/etc/resolv.conf:/etc/resolv.conf" "./$file_name" "$JIOTV_GO_ADDR"
}

auto_start() {
    # add $0 run to bash.bashrc
    if ! grep -q "$0 run" "$PREFIX/etc/bash.bashrc"; then
        echo "Adding auto start to bash.bashrc..."
        echo "$0 run" >>"$PREFIX/etc/bash.bashrc"
    else
        echo "Removing existing auto start from bash.bashrc..."
        grep -v "$0 run" "$PREFIX/etc/bash.bashrc" >"$PREFIX/etc/bash.bashrc.tmp"
        mv "$PREFIX/etc/bash.bashrc.tmp" "$PREFIX/etc/bash.bashrc"
    fi
}

# Check for the provided argument and perform the corresponding action
case "$1" in
"install")
    install_android "$@"
    ;;
"update")
    update_android "$@"
    ;;
"run")
    run_android "$@"
    ;;
"auto-start")
    auto_start "$@"
    ;;
"help")
    usage
    ;;
*)
    usage
    exit 1
    ;;
esac
