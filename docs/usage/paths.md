# JioTV Go Server URL Paths

This section provides information about the various web paths that JioTV Go offers. These paths allow you to interact with and access different features of the application.

## Web Paths

### Index

- **Path**: `/`

The gateway to the Home Page, where your JioTV Go adventure begins.

### Player Page

- **Path**: `/play/:channel_id`

Dive into the world of specific channels with the provided `channel_id`.

### FlowPlayer IFrame Player

- **Path**: `/player/:channel_id`

Immerse yourself with the default player (Flowplayer) for the specified `channel_id`.

### Clapper IFrame Player

- **Path**: `/clappr/:channel_id`

Experience the magic of the Clappr player for the specified `channel_id`.

# JioTV Go API Endpoints

This section provides information about the API endpoints that JioTV Go offers. These endpoints allow you to interact with and access different features of the application.

## API Endpoints

### Send OTP

- **Path**: `/login/sendOTP`
  Request an OTP to log in to JioTV.

### Verify OTP

- **Path**: `/login/verifyOTP`
  Verify the OTP and log in to JioTV.

### Get Channels data

- **Path**: `/channels`
  Discover the complete list of available channels in JSON format. DRM channels include `channel_url` for the MPD manifest and `key_url` for the `/live/key/:channel_id` license path.

## TV Endpoints

### M3U Playlist Alias

- **Path**: `/playlist.m3u`

  Instantly obtain an M3U playlist for IPTV.

  (Redirects to /channels?type=m3u for your convenience.)

You can append `?q=<level>` to the path where `<level>` should be replaced with `low`, `medium`, `high`, or `l`, `m`, `h` to set the quality of the stream. The default quality is `auto`.

You can also append `&c=split` to the path to have categories based on both language and genre. Example categories: `Hindi - Entertainment`, `English - News`, `Tamil - Sports`, etc.

You can also append `&sg=<genre_list>` to the path in order to skip specific genres. Here replace `<genre_list>` with comma(,) seperated list of genres.
Valid genres: `Entertainment`, `Movies`, `Kids`, `Sports`, `Lifestyle`, `Infotainment`, `News`, `Music`, `Devotional`, `Business`, `Educational`, `Shopping`, `JioDarshan`

You can also append `&sub=hide` to the path to leave out channels that require a separate subscription, or `&sub=only` to get a playlist of just those channels. Any other value, including omitting `sub`, returns every channel. This is the playlist equivalent of the "Hide channels requiring a separate subscription" checkbox on the home page.

### M3U Playlist

- **Path**: `/channels?type=m3u`

The actual path for the M3U playlist. You can append `&q=<level>` to the path as [above](#m3u-playlist-alias). You can also append `&c=split` or `&sub=hide` to the path as [above](#m3u-playlist-alias).

### M3U8 URL

- **Path**: `/live/:channel_id`

M3U8 stream file for the specified `channel_id`.

### M3U8 URL with Quality

- **Path**: `/live/:quality/:channel_id`

M3U8 stream file for the specified `channel_id` with the specified `quality`. The `quality` can be `low`, `medium`, `high`, or `l`, `m`, `h`.

### DRM MPD Manifest

- **Path**: `/live/mpd/:channel_id`

MPD manifest for DRM protected channels. You can also append `?q=<level>` to request a specific quality level.

### DRM License Key

- **Path**: `/live/key/:channel_id`

License key endpoint to authorize playback for Widevine DRM protected streams. This is used internally by supported players (like Kodi) to decrypt the MPD streams.

Explore these paths and endpoints to access the features and content offered by JioTV Go. They provide the foundation for interacting with the application and enjoying the available channels and streams.
