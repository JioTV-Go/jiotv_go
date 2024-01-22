# IPTV Guide

This section provides information about the various ways you can use JioTV Go in your IPTV setup.

## Generate M3U Playlist

JioTV Go provides an M3U playlist endpoint that you can use to generate an M3U playlist for your IPTV setup.

You can directly paste the following URL in your IPTV player:

```
http://localhost:5001/playlist.m3u
```

If you want to enforce a specific quality, you can use the `q` query parameter:

```
http://localhost:5001/playlist.m3u?q=high
```

Where `q` can be `low`, `medium`, `high`, or `l`, `m`, `h`.

## Electronic Program Guide (EPG)

JioTV Go provides an EPG endpoint that you can use to generate an EPG for your IPTV setup.

The EPG is disabled by default. To enable it, you need to set the `epg` config value to `true`. for more information, see the [Config](./config.md#epg-electronic-program-guide) page.

Once you have enabled the EPG, wait for a few minutes for the EPG to be generated. Then you can use the following URL in your IPTV player:

```
http://localhost:5001/epg.xml.gz
```

Once EPG is generated, it will be updated every 24 hours. The duration of the EPG is 2 days.

## Catchup

Currently, JioTV Go does not support catchup. Because I don't know how to implement it. If you know how to implement it, please open a pull request. I will be very grateful.

## Categories

JioTV Go M3U playlist provides the following categories:

- Entertainment
- Movies
- Kids
- Sports
- Lifestyle
- Infotainment
- News
- Music
- Devotional
- Business
- Educational
- Shopping

