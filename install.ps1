try {
    # Identify operating system architecture
    $architecture = (Get-WmiObject Win32_OperatingSystem).OSArchitecture
    switch ($architecture) {
        "64-bit" {
            $arch = "x86_64"
            break
        }
        "32-bit" {
            $arch = "x86"
            break
        }
        "ARM64" {
            $arch = "arm64"
            break
        }
        default {
            throw "Unsupported architecture: $architecture"
        }
    }

    Write-Host "Detected architecture: $arch"

    # If the binary already exists, delete it
    if (Test-Path jiotv_go.exe) {
        Write-Host "Deleting existing binary"
        Remove-Item jiotv_go.exe
    }

    # Fetch the latest binary
    # $binaryUrl = "https://api.github.com/repos/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-$arch.exe"
    # for testing
    $binaryUrl = "https://github.com/rabilrbl/jiotv_go/releases/download/dev.2024.01.23.18.48.1706035724/jiotv_go-windows-$arch.exe"
    Write-Host "Fetching the latest binary from $binaryUrl"
    Invoke-WebRequest -Uri $binaryUrl -OutFile jiotv_go.exe -UseBasicParsing

    # Add the directory to PATH
    $binaryPath = Convert-Path ".\"
    [Environment]::SetEnvironmentVariable("Path", "$($env:Path);$binaryPath", [EnvironmentVariableTarget]::Machine)

    # Inform the user
    Write-Host "JioTV Go has successfully downloaded and added to PATH. Start by running jiotv_go help"
}
catch {
    Write-Host "Error: $_"
    exit 1
}
