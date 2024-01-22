# Get Started with JioTV Go

## Installation

### Pre-Built Binaries

You can download the pre-built binaries for your platform from the [releases](https://github.com/rabilrbl/jiotv-go/releases) page.

> Android Termux users, please read the [Note for Termux Users](#note-for-termuxandroid-users) section below.

<!-- Generate a detail note for android users, to have termux downloaded from fdroid and playstore version is outdated. And to prevent DNS loopup error in termux Install proot and execute -->


The following table lists the binaries available for download:

| OS Name                      | Architecture  | Binary Name                        |
| ---------------------------- | ------------- | ----------------------------------- |
| Linux / Android                        | 386 (x86, i686)           | [jiotv_go-linux-386](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-386)                  |
| Linux / Android                        | amd64 (x86_64)        | [jiotv_go-linux-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-amd64)                |
| Linux / Android                        | arm           | [jiotv_go-linux-arm](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm)                  |
| Linux / Android                        | arm64 (aarch64)        | [jiotv_go-linux-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm64)                |
| Windows                      | 386 (x86, i686)  | [jiotv_go-windows-386.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-386.exe)            |
| Windows                      | amd64 (x86_64)| [jiotv_go-windows-amd64.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-amd64.exe)          |
| Windows                      | arm64 (aarch64)         | [jiotv_go-windows-arm64.exe](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-windows-arm64.exe)          |
| Darwin (macOS)               | amd64 (x86_64)       | [jiotv_go-darwin-amd64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-darwin-amd64)               |
| Darwin (macOS)               | arm64 (aarch64)        | [jiotv_go-darwin-arm64](https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-darwin-arm64)               |


#### Identifying your OS and Architecture

If you are unsure about your OS and Architecture, you can use the following commands to identify them:

##### Linux/macOS

```bash
uname -a
```

##### Windows

```powershell
systeminfo
```

### Note for Termux(Android) Users

You must have latest version of Termux installed from [F-Droid](https://f-droid.org/en/packages/com.termux/) or [GitHub Release](https://github.com/termux/termux-app/releases/latest).

You will need to install `proot` to prevent DNS Lookup errors.

```bash
pkg install proot
```

Then execute the binary with `proot`:

```bash
proot -b "$PREFIX/etc/resolv.conf:/etc/resolv.conf" ./jiotv_go-linux-{arch} [commands]
```