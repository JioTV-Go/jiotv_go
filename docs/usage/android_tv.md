# Android TV

JioTV Go is a web application that can be used on Android TV. This page provides information about how to use JioTV Go on Android TV.

You can use [Termux](#using-termux) or [UserLand](#using-userland) to run JioTV Go on Android TV.

<div class="warning">

> We recommend you use [automatic install script](../get_started.md#automatic-installation-recommended) to install JioTV Go on Android TV.
>
> However, if you want to install manually, refer manual installation guide for [Android](./android.md#step-3-download-jiotv-go-binary) from Step 3.

</div>

After setting up JioTV Go on your Android TV, first login using the web interface at `http://localhost:5001/`, then you can setup a IPTV App on your Android TV to access JioTV Go.

We recommend [TiviMate IPTV Player](https://play.google.com/store/apps/details?id=ar.tvplayer.tv) for this purpose or [OTT Navigator IPTV](https://ott-navigator-beta-for-android-tv-android.en.aptoide.com/app) for this purpose.

## Running JioTV Go on Android Phone and Accessing it on Android TV

You can run JioTV Go on your Android phone and access it on your Android TV. This is useful if you don't have a keyboard or mouse to control JioTV Go on your Android TV.

Once you have installed JioTV Go on your Android phone, to access it on TV, Read the [Accessing JioTV Go from another device](../faq.md#how-can-i-access-jiotv-go-from-another-device-eg-computertvphone-in-my-local-network) section in the [FAQ](../faq.md) page for more information.

Access JioTV Go on your Android TV at `http://{ip_address}:5001/`. Replace `{ip_address}` with the IP address of your Android phone.

## Running JioTV Go on Android TV

### Using UserLand

Download UserLand from [Google Play Store](https://play.google.com/store/apps/details?id=tech.ula) or [F-Droid](https://f-droid.org/en/packages/tech.ula/) and install it on your Android TV.

You can use UserLand to run JioTV Go on your Android TV. Choose the Alpine image when setting up UserLand. Then follow the Linux instructions to setup JioTV Go.

### Using Termux

You can use Termux to run JioTV Go on your Android TV. Follow the Linux instructions to setup JioTV Go at [Android](../usage/android.md).


## Auto Start JioTV Go on Android TV

We recommend you using this app from [Siddharth](https://github.com/siddharthsky) called [Sparkle-TV2](https://github.com/siddharthsky/SparkleTV2-auto-service) to auto start JioTV Go on Android TV. Download the APK from the [releases page](https://github.com/siddharthsky/SparkleTV2-auto-service/releases) and install it on your Android TV.

Other apps that you can use to auto start JioTV Go on Android TV:

- [Termux:Boot](https://play.google.com/store/apps/details?id=com.termux.boot)
- [AutoStart](https://play.google.com/store/apps/details?id=com.autostart)

## Third-party EPG Providers

If you are having issue with EPG in Android TV, please see if you can resolve it by [`egp` command](usage.md#3-epg-command). If you are still having issue, you can use third-party EPG providers.

The current implementation of EPG generation is not optimised for Android TV.

Use a third party EPG as mentioned here
```
https://avkb.short.gy/jioepg.xml.gz
```

The above link is from mitthu786. GitHub Link: https://github.com/mitthu786/tvepg

This will not only be easy on your poor tiny compute machine. It is recommended as you will not get any issues or problems.

EPG generation requires up to thousand http requests in total and 20 http requests per second. So you do the math and plus at last it requires a huge amount of memory and CPU for compression to a 3MB file.

There might be other EPG providers. You can search for them on the internet.

## Using mouse and keyboard on Android TV

For mouse and keyboard control such as using Ctrl key, etc. You can use any application that allows you to use mouse and keyboard on Android TV. 

Here are some applications that you can use from your Android phone to control your Android TV:

- [Bluetooth Mouse and Remote](https://play.google.com/store/apps/details?id=com.app.bluetoothremote)
- [Zank Remote](https://play.google.com/store/apps/details?id=zank.remote)