#!/bin/bash

# Step 1: Identify the operating system
OS=""
if [[ "$OSTYPE" == "linux-gnu" ]]; then
    OS="linux"
    if [[ -n "$OS_ENV" && "$OS_ENV" == "android" ]]; then
        OS="android"
    fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
else
    echo "Unsupported operating system: $OSTYPE"
    exit 1
fi

echo "Step 1: Identified operating system as $OS"

# Step 2: Identify processor architecture
ARCH=$(uname -m)
case $ARCH in
    "x86_64")
        ARCH="amd64"
        ;;
    "aarch64")
        ARCH="arm64"
        ;;
    "i386" | "i686")
        ARCH="386"
        ;;
    "armv7l")
        ARCH="arm"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Step 2: Identified processor architecture as $ARCH"

# Step 3: Fetch the latest binary
BINARY_URL="https://api.github.com/repos/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-$OS-$ARCH"
echo "Step 3: Fetching the latest binary from $BINARY_URL"
curl -SL --progress-bar --retry 5 --retry-delay 2 -o jiotv_go "$BINARY_URL" || { echo "Failed to download binary"; exit 1; }

# Step 4: Give executable permissions
chmod +x jiotv_go
echo "Step 4: Granted executable permissions to the binary"

# Step 5: Move binary to local bin folder
if [[ "$OS" == "android" && -n "$PREFIX" ]]; then
    BIN_FOLDER="$PREFIX/bin/"
    if [[ ! -d "$BIN_FOLDER" ]]; then
        mkdir -p "$BIN_FOLDER"
    fi
    mv jiotv_go "$BIN_FOLDER"
    echo "Step 5: Moved the binary to $BIN_FOLDER"
else
    if [[ ! -d "$HOME/bin" ]]; then
        mkdir -p "$HOME/bin"
    fi
    mv jiotv_go "$HOME/bin/"
    echo "Step 5: Moved the binary to $HOME/bin/"
fi

# Step 6: Inform the user
echo "JioTV Go has successfully installed. Start by running jiotv_go help"
