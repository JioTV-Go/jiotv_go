# Custom Channels Extension

JioTV Go now supports adding custom channel sources that will be visible on both the web dashboard and IPTV clients.

## Configuration

To enable custom channels, set the `custom_channels_file` option in your configuration file or use the `JIOTV_CUSTOM_CHANNELS_FILE` environment variable.

### Via Configuration File

**YAML (jiotv-config.yml):**
```yaml
custom_channels_file: "./configs/custom-channels.yml"
```

**JSON (jiotv-config.json):**
```json
{
  "custom_channels_file": "./configs/custom-channels.json"
}
```

### Via Environment Variable

```bash
export JIOTV_CUSTOM_CHANNELS_FILE="./configs/custom-channels.json"
```

## Custom Channels File Format

### JSON Format

```json
{
  "channels": [
    {
      "id": "custom_news_1",
      "name": "Sample News Channel",
      "url": "https://example.com/news/playlist.m3u8",
      "logo_url": "https://example.com/logos/news.png",
      "category": 12,
      "language": 6,
      "is_hd": true
    },
    {
      "id": "custom_entertainment_1",
      "name": "Sample Entertainment Channel",
      "url": "https://example.com/entertainment/playlist.m3u8",
      "logo_url": "https://example.com/logos/entertainment.png",
      "category": 5,
      "language": 1,
      "is_hd": false
    }
  ]
}
```

### YAML Format

```yaml
channels:
  - id: custom_news_1
    name: Sample News Channel
    url: https://example.com/news/playlist.m3u8
    logo_url: https://example.com/logos/news.png
    category: 12  # News
    language: 6   # English
    is_hd: true
  - id: custom_entertainment_1
    name: Sample Entertainment Channel
    url: https://example.com/entertainment/playlist.m3u8
    logo_url: https://example.com/logos/entertainment.png
    category: 5   # Entertainment
    language: 1   # Hindi
    is_hd: false
```

## Field Descriptions

- **id**: Unique identifier for the channel (required)
- **name**: Display name of the channel (required)
- **url**: Direct stream URL (M3U8/HLS format recommended) (required)
- **logo_url**: URL to channel logo image (optional)
- **category**: Category ID (see Category IDs below) (required)
- **language**: Language ID (see Language IDs below) (required)
- **is_hd**: Whether the channel is HD quality (boolean) (required)

## Category IDs

- 0: All Categories
- 5: Entertainment
- 6: Movies
- 7: Kids
- 8: Sports
- 9: Lifestyle
- 10: Infotainment
- 12: News
- 13: Music
- 15: Devotional
- 16: Business
- 17: Educational
- 18: Shopping
- 19: JioDarshan

## Language IDs

- 0: All Languages
- 1: Hindi
- 2: Marathi
- 3: Punjabi
- 4: Urdu
- 5: Bengali
- 6: English
- 7: Malayalam
- 8: Tamil
- 9: Gujarati
- 10: Odia
- 11: Telugu
- 12: Bhojpuri
- 13: Kannada
- 14: Assamese
- 15: Nepali
- 16: French
- 18: Other

## Features

- **Web Dashboard Integration**: Custom channels appear alongside JioTV channels in the web interface
- **IPTV M3U Support**: Custom channels are included in generated M3U playlists
- **Filtering Support**: Custom channels work with language and category filters
- **Live Streaming**: Direct streaming support for custom channel URLs
- **Error Handling**: Graceful handling of missing or invalid custom channels files

## Usage Examples

1. **Add a custom news channel:**
   ```json
   {
     "id": "my_news_channel", 
     "name": "My News Channel",
     "url": "https://streaming.example.com/news.m3u8",
     "category": 12,
     "language": 6,
     "is_hd": true
   }
   ```

2. **Access via web interface:** Navigate to `http://localhost:5001` and your custom channels will appear in the channel list

3. **Access via IPTV:** Use `http://localhost:5001/playlist.m3u` to get an M3U playlist including custom channels

4. **Direct stream access:** `http://localhost:5001/live/my_news_channel.m3u8`

## Notes

- Custom channels are loaded at startup. Restart the server after modifying the custom channels file
- Only M3U8/HLS URLs are recommended for streaming compatibility
- Ensure custom channel IDs are unique and don't conflict with existing JioTV channel IDs
- If the custom channels file is not found or contains errors, the server will continue to work with only JioTV channels