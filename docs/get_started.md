# Get Started with JioTV Go

## Installation

### Automatic Installation (Recommended)

We have video tutorials for [Windows](https://youtu.be/BnNTYTSvVBc),  and [Android](https://youtu.be/ejiuml11g8o) users. Please watch them if you are unsure about the installation process.

#### Linux/Android/macOS

Here's a one-liner to download and install the latest version of JioTV Go on Linux/Android/macOS.

Simply copy and paste the following command in your terminal and press `Enter`:

```bash
curl -fsSL https://jiotv_go.rabil.me/install.sh | bash
```

The above command will download the latest version of JioTV Go and install it in your system.

> **Termux** users, if you get an errors, first update pkg index and install `curl`, `openssl` packages by running the following command:
>
> ```bash
> pkg update
> pkg install curl openssl
> ```

<div class="warning">

See the [Docker Setup](#docker-setup) section for Docker installation instructions.

</div>

#### Windows

Here's a one-liner to download and install the latest version of JioTV Go on Windows.

Simply copy and paste the following command in your PowerShell terminal and press `Enter`:

```powershell
iwr -useb https://jiotv_go.rabil.me/install.ps1 | iex
```

### Pre-Built Binaries

You can also download the pre-built binaries for your platform from the [releases](https://github.com/rabilrbl/jiotv-go/releases) page or click on `Binary Name` links in the table below.

#### The following table lists the binaries available for download:

| OS Name        | Architecture (AKA) | Binary Name                                                                                                            |
| -------------- | ------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| Android        | arm64 (aarch64)    | [jiotv_go-android-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-android-arm64)         |
| Android        | amd64 (x86_64)     | [jiotv_go-android-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-android-amd64)         |
| Android        | arm                | [jiotv_go-android-arm](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-android-arm)             |
| Linux          | arm64 (aarch64)    | [jiotv_go-linux-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm64)             |
| Linux          | amd64 (x86_64)     | [jiotv_go-linux-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-amd64)             |
| Linux          | arm                | [jiotv_go-linux-arm](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm)                 |
| Linux          | 386 (x86, i686)    | [jiotv_go-linux-386](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-386)                 |
| Windows        | 386 (x86, i686)    | [jiotv_go-windows-386.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-386.exe)     |
| Windows        | amd64 (x86_64)     | [jiotv_go-windows-amd64.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-amd64.exe) |
| Windows        | arm64 (aarch64)    | [jiotv_go-windows-arm64.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-arm64.exe) |
| Darwin (macOS) | amd64 (x86_64)     | [jiotv_go-darwin-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-darwin-amd64)           |
| Darwin (macOS) | arm64 (aarch64)    | [jiotv_go-darwin-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-darwin-arm64)           |

#### Identifying your OS and Architecture

If you are unsure about your OS and Architecture, you can use the following commands to identify them:

#### Linux / Anroid / macOS

Execute the following command in your terminal:

```bash
uname -m
```

For Android, you can use any terminal emulator app. For example, [Termux](https://f-droid.org/en/packages/com.termux/) or [UserLAnd](https://f-droid.org/en/packages/tech.ula/). We recommend UserLAnd with Alpine as it emulates a Linux environment.

#### Windows (PowerShell)

Run the following command in your PowerShell terminal:

```powershell
(Get-WmiObject Win32_OperatingSystem).OSArchitecture
```

Then, look for your architecture in the [above table](#the-following-table-lists-the-binaries-available-for-download).

<div class="warning">

> Windows users, if you are unsure on the next steps, after downloading the binary, please read the [Using JioTV Go on Windows](./usage/windows.md) page.

</div>

## Build from Source

Refer the guide in [Development](./development.md#build-from-source) page.

## Docker Setup

Make sure you have [Docker](https://docs.docker.com/get-docker/) installed on your system.

### Run JioTV Go with Docker

Run the command:

```sh
docker run -p 5001:5001 -v ./.jiotv_go/secrets:/app/secrets ghcr.io/rabilrbl/jiotv_go
```

Open your web browser and visit [http://localhost:5001/](http://localhost:5001/).

### Using CLI Options with Docker

By default, JioTV Go Docker image runs with `serve --public` command. You can override this by passing the command as an argument to the `docker run` command.

For example, to run JioTV Go with `serve --public --port 8080` command, run:

```sh
docker run -p 8080:8080 -v ./.jiotv_go:/app/.jiotv_go ghcr.io/rabilrbl/jiotv_go serve --public --port 8080
```

### Keep JioTV Go Updated

To update to the latest version, run:

```sh
docker pull ghcr.io/rabilrbl/jiotv_go:latest
```

---

- Also read the [Usage](./usage/usage.md) page for more information.
- See the [Config](./config.md) page for more information about the configuration options.
