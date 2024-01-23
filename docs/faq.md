# Frequently Asked Questions

Before proceeding to the FAQ, here are the pages that you should read first:

- [Usage](../usage/usage.md)
- [IPTV Guide](../usage/iptv.md)
- [Config](../config.md)
- [Cloud Hosting](../cloud_hosting.md)

## Does JioTV Go support JioFiber / JioFiber Set-Top Box?

No. JioTV Go does not support JioFiber / JioFiber Set-Top Box. JioTV Go only supports JioTV from mobile application.

## How can I use JioTV Go on my TV?

You can use JioTV Go on your TV using the following methods:

Install JioTV Go on your Android TV using Termux or UserLand. Read the [Android TV](../usage/android_tv.md) page for more information.

## How can I access JioTV Go from another device (e.g. computer/TV/phone) in my local network?

Run the JioTV Go server with `--public` flag:

```sh
jiotv_go-linux-{arch} serve --public
```

Then find the IP address of your device with JioTV Go installed in your local network. You will have many tutorials on the internet on how to find the IP address of your device.

Now, on any device, access JioTV Go at `http://{ip_address}:5001/`. Replace `{ip_address}` with the IP address of your device.

## Can I host JioTV Go on a VPS or a cloud server?

Read the [Cloud Hosting](../cloud_hosting.md) page for more information.

## Can I host JioTV Go on a Home Server / Raspberry Pi?

Yes. Read the [Cloud Hosting](../cloud_hosting.md) page for more information.

## Why do I get error or blank in the player?

This error occurs when you have not logged in to JioTV Go or your session has expired. To fix this error, simply delete the `jiotv_credentials_v2.json` file and restart JioTV Go, then log in again.

## Does JioTV Go support catchup?

No. JioTV Go does not support catchup. Because I don't know how to implement it. If you know how to implement it, please open a pull request. I will be very grateful. See the [IPTV Guide](../usage/iptv.md#catchup) for more information. And [contributing](../contributing.md) page for more information about contributing.

## Why do I get buffering in the IPTV player?

Read the [Buffering issues on IPTV](./usage/iptv.md#buffering-issues-on-iptv-players) guide.

## Why do I see same resolution under quality options?

The two resolutions are same but with different bitrates.

If there are two same resolutions with different bitrates, higher bitrate will be selected based on your internet bandwidth/speed.

## How do I update JioTV Go?

Read the [Update command section](./usage/usage.md#2-update-command) in the [Usage](../usage/usage.md) page for more information.

## How do I update JioTV Go if I have installed it using Docker?

Simply pull the latest image from Docker Hub:

```sh
docker pull ghcr.io/rabilrbl/jiotv_go:latest
```

## How do I update JioTV Go if I have installed it using Termux?

Simply run the following command:

```sh
jiotv_go-linux-{arch} update
```

## How do I stop JioTV Go server?

Press `Ctrl + C` in the terminal where you have started JioTV Go.

## How do I Uninstall JioTV Go completely?

Simply delete all files related to JioTV Go. If you have installed JioTV Go using Docker, then delete the Docker image. If you have installed JioTV Go via pre-built binaries, then delete the binary file. 
