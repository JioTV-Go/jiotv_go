# Get Started with JioTV Go

## Installation

### Pre-Built Binaries

You can download the pre-built binaries for your platform from the [releases](https://github.com/rabilrbl/jiotv-go/releases) page or click on `Binary Name` links in the table below.

<div class="warning">
<b>Android Termux</b> users, please read the <a href="#note-for-termuxandroid-users">Note for Termux Users</a> section below.
</div>

<!-- Generate a detail note for android users, to have termux downloaded from fdroid and playstore version is outdated. And to prevent DNS loopup error in termux Install proot and execute -->


#### The following table lists the binaries available for download:

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

### Note for Termux(Android) Users

You must have latest version of Termux installed from [F-Droid](https://f-droid.org/en/packages/com.termux/) or [GitHub Release](https://github.com/termux/termux-app/releases/latest).

Start by updating pkg index and install `wget` and `openssl`

```bash
pkg update && pkg install wget openssl-tool
```

Identify your architecture:

```bash
uname -m
```

Now look for the architecture in the [table above](#the-following-table-lists-the-binaries-available-for-download).

Then download the binary for your architecture, replace `{arch}` with your architecture:

```bash
wget https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-{arch}
```

For example, if your architecture from `uname -m` is `aarch64`, then you will download the binary for `arm64` architecture:

```bash
wget https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm64
```

Make the binary executable:

```bash
chmod +x jiotv_go-linux-{arch}
```

You will need to install `proot` to prevent DNS Lookup errors.

```bash
pkg install proot
```

Then execute the binary with `proot`:

```bash
proot -b "$PREFIX/etc/resolv.conf:/etc/resolv.conf" ./jiotv_go-linux-{arch} [commands]
```

Now to make it easier to execute the binary, you can create an alias:

```bash
echo "alias jiotv_go="proot -b \"$PREFIX/etc/resolv.conf:/etc/resolv.conf\" $PWD/jiotv_go-linux-{arch}" >> $PREFIX/etc/bash.bashrc
source $PREFIX/etc/bash.bashrc
```

Now you can execute the binary easily from anywhere:


```bash
jiotv_go [commands]
```