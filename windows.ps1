# Configuration
$REPO_OWNER = "rabilrbl"
$REPO_NAME = "jiotv_go"
$BINARY_NAME = "jiotv_go"

# Determine the device's architecture
$ARCH = (Get-WmiObject -Class Win32_ComputerSystem).SystemType

switch -Wildcard ($ARCH) {
    "*64*" { $ARCH = "amd64" }
    # "*32*" { $ARCH = "386" }
    # "*ARM*" { $ARCH = "arm" }
    default {
        Write-Host "Unsupported architecture: $ARCH"
        exit 1
    }
}

function usage {
    Write-Host "Usage: $MyInvocation.InvocationName.ps1 {install|update|run}"
    Write-Host "  install: Install JioTV Go for the first time"
    Write-Host "  update: Update JioTV Go to the latest version"
    Write-Host "  run: Run JioTV Go (default)"
    Write-Host ""
    Write-Host "You can optionally specify the \"host:port\" to run JioTV Go as a second argument."
}

function release_file_name {
    # Check if curl is installed
    if (-not (Test-Path -Path "C:\Program Files\curl\curl.exe" -PathType Leaf)) {
        Write-Host "Error: curl not found. Please install curl."
        return 1
    }

    # Get the latest release version from GitHub API
    Write-Host "Fetching latest release info from GitHub..."
    $RELEASE_INFO = Invoke-WebRequest -Uri "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | ConvertFrom-Json
    $LATEST_VERSION = $RELEASE_INFO.tag_name

    $file_name = "$BINARY_NAME-$LATEST_VERSION-windows-$ARCH.exe"
}

function download_binary {
    # fetch latest file name
    if (-not (release_file_name)) {
        Write-Host "Failed to fetch latest release info from GitHub."
        return 1
    }

    # Construct the download URL for the binary
    $DOWNLOAD_URL = "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/$file_name"

    # Download and install the binary
    Write-Host "Downloading JioTV Go $LATEST_VERSION for $ARCH..."
    if (-not (Test-Path -Path $file_name -PathType Leaf)) {
        Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $file_name -UseBasicParsing
    }

    if ((Test-Path -Path $file_name -PathType Leaf) -and (-not (Test-Path -Path $BINARY_NAME -PathType Leaf))) {
        Move-Item -Path $file_name -Destination $BINARY_NAME -Force
        Write-Host "JioTV Go $LATEST_VERSION for $ARCH has been downloaded."
        Write-Host "Execute \"$MyInvocation.InvocationName.ps1 run\" to start JioTV Go."
    } else {
        Write-Host "Error: Failed to download binary. Please try again later."
        return 1
    }
}

# Function to install the binary
function install_windows {
    if (download_binary) {
        Write-Host "JioTV Go $LATEST_VERSION for $ARCH has been downloaded."
        Write-Host "Execute \"$MyInvocation.InvocationName.ps1 run\" to start JioTV Go."
    } else {
        Write-Host "Failed to download JioTV Go"
        return 1
    }
}

# Function to update the binary
function update_windows {
    # fetch existing file name, if multiple files are present, pick the latest one
    $existing_file_name = Get-ChildItem | Where-Object { $_.Name -match "$BINARY_NAME-.*-$ARCH.exe" } | Sort-Object CreationTime -Descending | Select-Object -First 1

    if ($existing_file_name) {
        Write-Host "Found existing file: $($existing_file_name.Name)"
        # fetch latest file name
        if (-not (release_file_name)) {
            Write-Host "Failed to fetch latest release info from GitHub."
            return 1
        }

        # compare existing file name with file name
        # if both are same, need not download again
        if ($existing_file_name.Name -eq $file_name) {
            Write-Host "JioTV Go is already up to date."
            return 0
        } else {
            if (download_binary) {
                # Delete the old binary
                Write-Host "Deleting old version $($existing_file_name.Name)..."
                Remove-Item -Path $existing_file_name.FullName -Force
                Write-Host "JioTV Go has been updated to latest $LATEST_VERSION version."
            } else {
                Write-Host "Failed to update JioTV Go"
            }
        }
    } else {
        Write-Host "Missing existing file name. Is JioTV Go installed?"
        Write-Host "Execute \"$MyInvocation.InvocationName.ps1 install\" to install JioTV Go."
    }
}

# Function to run the binary
function run_windows {
    # fetch file name from ls command, if multiple files are present, pick the latest one
    $file_name = Get-ChildItem | Where-Object { $_.Name -match "$BINARY_NAME-.*-$ARCH.exe" } | Sort-Object CreationTime -Descending | Select-Object -First 1

    # Check if the binary exists
    if (-not $file_name) {
        Write-Host "Error: Binary '$BINARY_NAME' not found in the current directory."
        Write-Host "Execute \"$MyInvocation.InvocationName.ps1 install\" to install JioTV Go."
        return 1
    }

    # Add optional second argument as the address to run the binary on
    $JIOTV_GO_ADDR = "localhost:5001"
    if ($args.Length -ge 2) {
        $JIOTV_GO_ADDR = $args[1]
    }

    # Run the Windows binary
    Write-Host "Running JioTV Go at $JIOTV_GO_ADDR for $ARCH..."
    Start-Process -Wait -FilePath $file_name -ArgumentList $JIOTV_GO_ADDR
}

# Check for the provided argument and perform the corresponding action
switch ($args[0]) {
    "install" { install_windows }
    "update" { update_windows }
    "run" { run_windows }
    "help" { usage }
    default {
        usage
        exit 1
    }
}
