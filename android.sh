#!/bin/bash

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
  "x86_64")
    ARCH="amd64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Function to install the binary
install_android() {
  echo "Installing wget, curl, openssl and proot from Termux repositories..."
  echo "Please be patient, this can take a few minutes to complete..."
  pkg install wget curl openssl proot -y

  # Get the latest release version from GitHub API
  RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
  LATEST_VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": "[^"]*' | sed 's/"tag_name": "//')

  file_name="$BINARY_NAME-$LATEST_VERSION-linux-$ARCH"

  # Construct the download URL for the binary
  DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/$file_name"


  # Download and install the binary
  echo "Downloading JioTV Go $LATEST_VERSION for $ARCH..."
  wget "$DOWNLOAD_URL"
  chmod +x "$file_name"

  echo "$BINARY_NAME $LATEST_VERSION for $ARCH has been downloaded and installed to $INSTALL_DIR"
}

# Function to update the binary
update_android() {
  # Delete the old binary from ls command
  rm -rf $(ls | grep "$BINARY_NAME")

  # Get the latest release version from GitHub API
  RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
  LATEST_VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": "[^"]*' | sed 's/"tag_name": "//')

  file_name="$BINARY_NAME-$LATEST_VERSION-linux-$ARCH"

  # Construct the download URL for the binary
  DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/$file_name"

  # Update the binary
  echo "Updating $BINARY_NAME to $LATEST_VERSION for $ARCH..."
  wget "$DOWNLOAD_URL"
  chmod +x "$file_name"

  echo "JioTV Go has been updated to $LATEST_VERSION for $ARCH"
}

# Function to run the binary
run_android() {
  # fetch file name from ls command
  file_name=$(ls | grep "$BINARY_NAME")

   # Check if the binary exists
  if [ -z "$file_name" ]; then
    echo "Error: Binary '$BINARY_NAME' not found in the current directory. Run "./$0 install" to download the binary."
    return 1
  fi

  # Add optional second argument as the address to run the binary on
  JIOTV_GO_ADDR="localhost:5001"
  if [ ! -z "$2" ]; then
    JIOTV_GO_ADDR="$2"
  fi

  # Run the Android binary
  echo "Running JioTV Go at $JIOTV_GO_ADDR for $ARCH..."

  proot -b $PREFIX/etc/resolv.conf:/etc/resolv.conf ./$file_name $JIOTV_GO_ADDR
  
}

# Check for the provided argument and perform the corresponding action
case "$1" in
  "install")
    install_android
    ;;
  "update")
    update_android
    ;;
  "run")
    run_android "$@"
    ;;
  *)
    echo "Usage: $0 {install|update|run}"
    exit 1
    ;;
esac
