# Explore JioTV Go's Paths and Endpoints

This section provides information about the various web paths and API endpoints that JioTV Go offers. These paths and endpoints allow you to interact with and access different features of the application.

## Web Paths

### `/`

- **Description**: The gateway to the Home Page, where your JioTV Go adventure begins.

### `/play/:channel_id`

- **Description**: Dive into the world of specific channels with the provided `channel_id`.

### `/playlist.m3u`

- **Description**: Instantly obtain an M3U playlist for IPTV. (Redirects to `/channels?type=m3u` for your convenience.) You can add `?q=<level>` where `<level>` should be replaced with `low`, `medium`, `high`, or `l`, `m`, `h` to set the quality of the stream. The default quality is `auto`.

### `/player/:channel_id`

- **Description**: Immerse yourself with the default player (Flowplayer) for the specified `channel_id`.

### `/clappr/:channel_id`

- **Description**: Experience the magic of the Clappr player for the specified `channel_id`.

## API Endpoints

### `/login/sendOTP`

- **Description**: Request an OTP to log in to JioTV.

### `/login/verifyOTP`

- **Description**: Verify the OTP and log in to JioTV.

### `/login`

- **Description**: Log in to JioTV with password authentication. Either pass the `username` and `password` as query parameters or as JSON in the post request body.

### `/channels`

- **Description**: Discover the complete list of available channels.

### `/channels?type=m3u`

- **Description**: Effortlessly acquire an M3U playlist for IPTV. You can add `q=<level>` where `<level>` should be replaced with `low`, `medium`, `high`, or `l`, `m`, `h` to set the quality of the stream. The default quality is `auto`.

### `/live/:channel_id`

- **Description**: Tune in to live TV with the specified `channel_id`.

Explore these paths and endpoints to access the features and content offered by JioTV Go. They provide the foundation for interacting with the application and enjoying the available channels and streams.
