# JioTV Go Server URL Paths

This section provides information about the various web paths that JioTV Go offers. These paths allow you to interact with and access different features of the application.

## Web Paths

### Index

- **Path**: `/`

- **Description**: The gateway to the Home Page, where your JioTV Go adventure begins.

### Player Page

- **Path**: `/play/:channel_id`

- **Description**: Dive into the world of specific channels with the provided `channel_id`.

### FlowPlayer IFrame Player

- **Path**: `/player/:channel_id`

- **Description**: Immerse yourself with the default player (Flowplayer) for the specified `channel_id`.

### Clapper IFrame Player

- **Path**: `/clappr/:channel_id`

- **Description**: Experience the magic of the Clappr player for the specified `channel_id`.

# JioTV Go API Endpoints

This section provides information about the API endpoints that JioTV Go offers. These endpoints allow you to interact with and access different features of the application.

## API Endpoints

### `/login/sendOTP`

- **Description**: Request an OTP to log in to JioTV.

### `/login/verifyOTP`

- **Description**: Verify the OTP and log in to JioTV.

### `/login`

- **Description**: Log in to JioTV with password authentication. Either pass the `username` and `password` as query parameters or as JSON in the post request body.

### `/channels`

- **Description**: Discover the complete list of available channels in JSON format.
  

## TV Endpoints

### M3U Playlist Alias

- **Path**: `/playlist.m3u`

- **Description**: Instantly obtain an M3U playlist for IPTV. 
 
(Redirects to `/channels?type=m3u` for your convenience.) 
  
You can add `?q=<level>` where `<level>` should be replaced with `low`, `medium`, `high`, or `l`, `m`, `h` to set the quality of the stream. The default quality is `auto`.

### M3U Playlist

- **Path**: `/channels?type=m3u`

- **Description**: Effortlessly acquire an M3U playlist for IPTV. You can the set `q` query parameter as described above.

### M3U8 URL

- **Path**: `/live/:channel_id`

- **Description**: M3U8 stream file for the specified `channel_id`.

### M3U8 URL with Quality

- **Path**: `/live/:quality/:channel_id`

- **Description**: M3U8 stream file for the specified `channel_id` with the specified `quality`. The `quality` can be `low`, `medium`, `high`, or `l`, `m`, `h`.


Explore these paths and endpoints to access the features and content offered by JioTV Go. They provide the foundation for interacting with the application and enjoying the available channels and streams.