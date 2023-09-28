#!/bin/bash

# This script is used to detect os type and download auto install script and execute it.

# If env var TERMUX_VERSION is set, then we are running in termux android app.
# Otherwise, we are running in linux/darwin

throw_install_failed() {
    # Check if optional message is provided
    if [ ! -z "$1" ]; then
        echo "$1"
    else
        echo "Installation failed."
    fi
    echo "Please perform manual installation by visiting https://github.com/rabilrbl/jiotv_go"
    exit 1
}

check_previous_command_success() {
    if [ $? -ne 0 ]; then
        throw_install_failed "$@"
    fi
}

install() {
    OS_ENV=$1
    if [ "$OS_ENV" == "android" ]; then
        file_name="android.sh"
    else
        file_name="install.sh"
    fi
    if [ -f "jiotv_go.sh" ]; then
        echo "Existing jiotv_go.sh found. Removing it"
        rm "jiotv_go.sh"
        check_previous_command_success "Failed to remove old jiotv_go.sh. Do you have permission to remove a file?"
    fi
    echo "Downloading JioTV Go installation script"
    curl --progress-bar -Lo "jiotv_go.sh" "https://raw.githubusercontent.com/rabilrbl/jiotv_go/main/$file_name"
    check_previous_command_success "Failed to download JioTV Go installation script. Do you have internet connection?"
    chmod +x "jiotv_go.sh"
    check_previous_command_success "Failed to make jiotv_go.sh executable. Do you have permission to make a file executable?"
    ./jiotv_go.sh install
    check_previous_command_success
}

# If $HOME == /data/data/com.termux/files/home then we are running in termux android app.
if [ "$HOME" == "/data/data/com.termux/files/home" ]; then
    # android
    install "android"
else
    # linux or darwin
    install "linux"
fi
