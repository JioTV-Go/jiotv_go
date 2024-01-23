# Termux (Android) Guide

You must have latest version of Termux installed from [F-Droid](https://f-droid.org/en/packages/com.termux/) or [GitHub Release](https://github.com/termux/termux-app/releases/latest).

In any typical Linux system, the nameserver is defined in `/etc/resolv.conf`. But in Termux, it is defined in `$PREFIX/etc/resolv.conf`. So, we need to mount `$PREFIX/etc/resolv.conf` to `/etc/resolv.conf` in the container to prevent DNS Lookup errors.

## Install required packages

Start by updating pkg index and install `wget` and `openssl`

```bash
pkg update && pkg install wget openssl-tool
```

## Identify your architecture

If you are unsure about your architecture, you can use the following command to identify it:

```bash
uname -m
```

Now look for the architecture in the [table](./get_started.md#the-following-table-lists-the-binaries-available-for-download).


## Download binary

Then download the binary for your architecture, replace `{arch}` with your architecture:

```bash
wget "https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-{arch}"
```

For example, if your architecture from `uname -m` is `aarch64`, then you will download the binary for `arm64` architecture:

```bash
wget "https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-arm64"
```

Make the binary executable

```bash
chmod +x jiotv_go-linux-{arch}
```

## Execute the binary

You will need to install `proot` to prevent DNS Lookup errors.

```bash
pkg install proot
```

Then execute the binary with `proot`:

```bash
proot -b "$PREFIX/etc/resolv.conf:/etc/resolv.conf" ./jiotv_go-linux-{arch} [commands]
```

<div class="warning">
don't freak out, we have a shortcut below
</div>

## PRoot Shortcut

I know that the above command is too long to type every time you want to execute the binary. So, here is a shortcut for you.

Paste the following command in your terminal and execute it:

```bash
echo "alias jiotv_go="proot -b \"$PREFIX/etc/resolv.conf:/etc/resolv.conf\" $PWD/jiotv_go-linux-{arch}" >> $PREFIX/etc/bash.bashrc
source $PREFIX/etc/bash.bashrc
```

Now you can execute the binary easily from anywhere:

```bash
jiotv_go [commands]
```
