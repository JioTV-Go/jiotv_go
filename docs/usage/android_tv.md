# Android TV

JioTV Go is a web application that can be used on Android TV. This page provides information about how to use JioTV Go on Android TV.

You can use [Termux](#using-termux) or [UserLand](#using-userland) to run JioTV Go on Android TV. We recommend [UserLand](#using-userland).

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

You can use Termux to run JioTV Go on your Android TV. Follow the Termux instructions to setup JioTV Go at [Termux](./termux.md).