# Custom Channels

JioTV Go supports adding custom channel sources alongside the official JioTV channels. This feature allows you to add your own IPTV streams, local media servers, or other streaming sources.

## Configuration

Custom channels are enabled by default. To disable them, set the `disable_custom_channels` configuration option to `true`:

### Environment Variable
```bash
export JIOTV_DISABLE_CUSTOM_CHANNELS=true
```

### Config File
```json
{
  "disable_custom_channels": true
}
```

## Adding Custom Channels

Create a `custom-channels.json` file in your JioTV Go data directory (usually `~/.jiotv_go/`):

```json
{
  "channels": [
    {
      "id": "my_channel_1",
      "name": "My Custom Channel",
      "url": "https://example.com/stream.m3u8",
      "logo_url": "https://example.com/logo.png",
      "category": 5,
      "language": 6,
      "is_hd": true
    }
  ]
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier for the channel (will be prefixed with `custom_`) |
| `name` | string | Yes | Display name of the channel |
| `url` | string | Yes | Direct stream URL (M3U8, RTMP, etc.) |
| `logo_url` | string | No | URL to channel logo image |
| `category` | integer | No | Channel category (see categories below) |
| `language` | integer | No | Channel language (see languages below) |
| `is_hd` | boolean | No | Whether the channel is HD quality |

### Categories

| ID | Category |
|----|----------|
| 0 | All Categories |
| 5 | Entertainment |
| 6 | Movies |
| 7 | Kids |
| 8 | Sports |
| 9 | Lifestyle |
| 10 | Infotainment |
| 12 | News |
| 13 | Music |
| 15 | Devotional |
| 16 | Business |
| 17 | Educational |
| 18 | Shopping |
| 19 | JioDarshan |

### Languages

| ID | Language |
|----|----------|
| 0 | All Languages |
| 1 | Hindi |
| 2 | Marathi |
| 3 | Punjabi |
| 4 | Urdu |
| 5 | Bengali |
| 6 | English |
| 7 | Malayalam |
| 8 | Tamil |
| 9 | Gujarati |
| 10 | Odia |
| 11 | Telugu |
| 12 | Bhojpuri |
| 13 | Kannada |
| 14 | Assamese |
| 15 | Nepali |
| 16 | French |
| 18 | Other |

## Usage

Once configured, custom channels will appear in:

1. **Web Interface**: Listed alongside JioTV channels in the main dashboard
2. **IPTV Clients**: Included in M3U playlists (`/channels?type=m3u`)
3. **API**: Available through the channels API endpoint (`/channels`)

### Accessing Custom Channels

- **Web Player**: `http://your-server:port/play/custom_my_channel_1`
- **Direct Stream**: `http://your-server:port/live/custom_my_channel_1.m3u8`
- **API**: Channel ID will be `custom_my_channel_1` in all API responses

## File Locations

The system looks for custom channel files in this order:

1. `custom-channels.json` (JSON format)
2. `custom-channels.yml` (YAML format - planned for future implementation)
3. `custom-channels.yaml` (YAML format - planned for future implementation)

All files should be placed in your JioTV Go data directory (configured via `path_prefix` or `JIOTV_PATH_PREFIX`).

## Stream Compatibility

Custom channels support any streaming format that can be played directly by media players:

- **HLS** (`.m3u8`) - Recommended
- **DASH** (`.mpd`)
- **Direct media files** (`.mp4`, `.mkv`, etc.)
- **RTMP streams**
- **Other formats supported by your media player**

## Example Configuration

```json
{
  "channels": [
    {
      "id": "local_news",
      "name": "Local News Channel",
      "url": "https://news-stream.example.com/live.m3u8",
      "logo_url": "https://news-stream.example.com/logo.png",
      "category": 12,
      "language": 6,
      "is_hd": true
    },
    {
      "id": "sports_stream",
      "name": "Sports Stream",
      "url": "https://sports.example.com/stream.m3u8",
      "logo_url": "https://sports.example.com/logo.png",
      "category": 8,
      "language": 1,
      "is_hd": false
    },
    {
      "id": "radio_station",
      "name": "Music Radio",
      "url": "https://radio.example.com/stream.mp3",
      "category": 13,
      "language": 6,
      "is_hd": false
    }
  ]
}
```

## Troubleshooting

### Channels Not Appearing

1. Check if custom channels are enabled in configuration
2. Verify the JSON file is valid (use a JSON validator)
3. Ensure the file is in the correct location
4. Check server logs for any error messages
5. Restart JioTV Go after adding/modifying channels

### Streaming Issues

1. Verify the stream URL is accessible
2. Check if the stream requires special headers or authentication
3. Test the stream URL directly in a media player
4. Some streams may not work due to CORS restrictions

### Logs

Custom channel loading is logged during server startup. Check the logs for messages like:
- `"Loaded X custom channels"`
- `"Custom channels are disabled via configuration"`
- Any error messages related to file parsing or channel validation