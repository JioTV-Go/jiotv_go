# Get Started with JioTV Go

## Installation

<div class="warning">

See the [Docker Setup](#docker-setup) section for Docker installation instructions.

</div>

### Pre-Built Binaries

You can download the pre-built binaries for your platform from the [releases](https://github.com/rabilrbl/jiotv-go/releases) page or click on `Binary Name` links in the table below.

<div class="warning">
<b>Android Termux</b> users, please read the <a href="./termux.md">Note for Termux Users</a> page.
</div>

<!-- Generate a detail note for android users, to have termux downloaded from fdroid and playstore version is outdated. And to prevent DNS loopup error in termux Install proot and execute -->


#### The following table lists the binaries available for download:

| OS Name                      | Architecture (AKA)  | Binary Name                        |
| ---------------------------- | ------------- | ----------------------------------- |
| Linux / Android                        | arm64 (aarch64)        | [jiotv_go-linux-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm64)                |
| Linux / Android                        | amd64 (x86_64)        | [jiotv_go-linux-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-amd64)                |
| Linux / Android                        | arm           | [jiotv_go-linux-arm](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm)                  |
| Linux / Android                        | 386 (x86, i686)           | [jiotv_go-linux-386](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-386)                  |
| Windows                      | 386 (x86, i686)  | [jiotv_go-windows-386.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-386.exe)            |
| Windows                      | amd64 (x86_64)| [jiotv_go-windows-amd64.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-amd64.exe)          |
| Windows                      | arm64 (aarch64)         | [jiotv_go-windows-arm64.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-arm64.exe)          |
| Darwin (macOS)               | amd64 (x86_64)       | [jiotv_go-darwin-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-darwin-amd64)               |
| Darwin (macOS)               | arm64 (aarch64)        | [jiotv_go-darwin-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-darwin-arm64)               |


#### Identifying your OS and Architecture

If you are unsure about your OS and Architecture, you can use the following commands to identify them:

#### Linux / Anroid / macOS

Execute the following command in your terminal:

```bash
uname -m
```

For Android, you can use any terminal emulator app. For example, [Termux](https://f-droid.org/en/packages/com.termux/) or [UserLAnd](https://f-droid.org/en/packages/tech.ula/). We recommend UserLAnd with Alpine as it emulates a Linux environment.


#### Windows (PowerShell)

```powershell
systeminfo
```

## Docker Setup

Make sure you have [Docker](https://docs.docker.com/get-docker/) installed on your system.

### Run JioTV Go with Docker
Run the command:

```sh
docker run -p 5001:5001 -v ./.jiotv_go/secrets:/app/secrets ghcr.io/rabilrbl/jiotv_go
```

Open your web browser and visit [http://localhost:5001/](http://localhost:5001/).

### Keep JioTV Go Updated

To update to the latest version, run:

```sh
docker pull ghcr.io/rabilrbl/jiotv_go:latest
```

---

- Also read the [Usage](./usage/usage.md) page for more information.
- See the [Config](./config.md) page for more information about the configuration options.
