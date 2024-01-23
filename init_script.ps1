try {
    # Identify processor architecture
    $architecture = (Get-WmiObject Win32_Processor).Architecture
    switch ($architecture) {
        9 {
            $arch = "arm64"
            break
        }
        0 {
            $arch = "386"
            break
        }
        5 {
            $arch = "x86_64"
            break
        }
        default {
            throw "Unsupported architecture: $architecture"
        }
    }

    Write-Host "Detected architecture: $arch"

    # Fetch the latest binary
    $binaryUrl = "https://api.github.com/repos/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-$arch.exe"
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
