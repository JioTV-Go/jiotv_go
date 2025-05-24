# Config

The `config` package provides configuration settings for JioTV Go, a server that serves JioTV API content. This documentation outlines the available configuration.

<div class="warning">
    You can set following configuration options using either config file (toml, yaml and json) or environment variables. We recommend using toml config file as it is easier to manage. See <a href="#example-configurations">Example Configuration</a> for more details.
</div>

## Configuration Options

### EPG (Electronic Program Guide):

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable EPG generation. | `epg` | `JIOTV_EPG` | `false` |

An EPG is an electronic program guide, an interactive on-screen menu that displays broadcast programming television programs schedules for each channel. It is generated from the JioTV API.

### Debug Mode:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable debug mode. | `debug` | `JIOTV_DEBUG` | `false` |

When `debug: true`, logging becomes more verbose, including file and line numbers for log messages, and the log prefix is set to `[DEBUG]`. This option works in conjunction with `log_to_stdout` and `log_path` to control the overall logging behavior. It is recommended to disable debug mode for regular use unless you are troubleshooting issues.

### TS Handler:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable TS Handler. | `disable_ts_handler` | `JIOTV_DISABLE_TS_HANDLER` | `false` |

TS Files are the video files that are streamed by the JioTV API.

By setting `disable_ts_handler` to `true`, server takes less load.

If `disable_ts_handler` is `true`, then TS files will be served directly from Jio API.

Otherwise the request is sent through the server as an intermediary.

### Logout Feature:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable the logout feature. | `disable_logout` | `JIOTV_DISABLE_LOGOUT` | `false` |

Simply put, the logout feature allows you to log out of your JioTV account in the web interface. Disabling this feature will make the logout button in the web interface non-functional.

### DRM (Digital Rights Management):

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable DRM. | `drm` | `JIOTV_DRM` | `false` |

DRM is a method of restricting access to copyrighted. The latest version of JioTV App uses DRM.
For future compatibility, I have added this feature.

Currently, the DRM is only supported by the web interface. It is not supported by the IPTV playlist.
For more detailed information about the DRM feature, including setup and limitations, please see [DRM Documentation](./drm.md).

### Title:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Title of the webpage. | `title` | `JIOTV_TITLE` | `JioTV Go` |

The title is displayed in the browser tab and the web interface.

### URL Encryption:

Enable or disable URL encryption.

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable URL encryption. | `disable_url_encryption` | `JIOTV_DISABLE_URL_ENCRYPTION` | `false` |

URL encryption prevents hackers from injecting URLs into the server. If you think it is unnecessary, you can disable it. But it is recommended to enable it.

### Path Prefix:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Folder path for all JioTV Go related files. | `path_prefix` | `JIOTV_PATH_PREFIX` | `$HOME/.jiotv_go` |

All JioTV Go related files are stored in this folder. This includes the IPTV playlist, the EPG, and the credentials file.

### Proxy:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Proxy URL. | `proxy` | `JIOTV_PROXY` | `""` |

Useful for bypassing geo-restrictions and IP restrictions for JioTV API.

If you want to use a proxy, set this config before you run the binary. Examples include
For Socks5 Proxy: value will be `socks5://user:pass@host:port`.
All Other Proxy (user:pass@host:port).

If your proxy does not require authentication, you can omit the `user:pass@` part.

### Log Path:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Directory for storing log files. | `log_path` | `JIOTV_LOG_PATH` | `""` (empty string) |

This option specifies the directory where the `jiotv_go.log` file will be stored.
If `log_path` is not set (i.e., an empty string), log files will be stored in a default location, which is typically under the directory specified by `path_prefix` (if set), or `$HOME/.jiotv_go/` if `path_prefix` is also not set.
If a custom path is provided, the application will attempt to create the directory if it doesn't exist.

### Log to Stdout:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Enable or disable logging to standard output. | `log_to_stdout` | `JIOTV_LOG_TO_STDOUT` | `false` |

This option controls whether log messages are also output to the standard output (the console).
Set to `true` to see logs in your terminal, or `false` to suppress console logging. The default value is `false` when specified in a configuration file.

### Custom Channels Path:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Path to custom channels JSON file. | `custom_channels_path` | `JIOTV_CUSTOM_CHANNELS_PATH` | `""` (empty string) |

This option specifies the path to a JSON file containing custom channel definitions. When set, JioTV Go will load these custom channels alongside the official JioTV channels, making them available in both the web interface and IPTV playlists.

If the path is empty or the file doesn't exist, custom channels will be disabled. Custom channels appear with a `custom_` prefix in their IDs to avoid conflicts with official channels.

For detailed information about custom channels configuration and examples, see the [Custom Channels section](#custom-channels-configuration) below.

## Example Configurations

Below are example configuration file for JioTV Go. All fields are optional, and the values shown are the default settings:

You can also specify the path to the configuration file using the `--config` flag.

### Example TOML Configuration

You can save the following configuration in a file named `jiotv_go.toml`. JioTV Go will automatically load the configuration from this file if it is present in the same directory as the binary.

The file is also available at [configs/jiotv_go-config.toml](https://github.com/jiotv-go/jiotv_go/blob/main/configs/jiotv_go-config.toml).

Omit the lines with `#` as they are comments. They are only for explanation purposes.

```toml
# Example config file for JioTV Go
# All fields mentioned below are optional.

# Enable Or Disable EPG Generation. Default: false
epg = false

# Enable Or Disable Debug Mode. Default: false
debug = false

# Enable Or Disable TS Handler. While TS Handler is enabled, the server will serve the TS files directly from JioTV API. Default: false
disable_ts_handler = false

# Enable Or Disable Logout feature. Default: true
disable_logout = false

# Enable Or Disable DRM. As DRM is not supported by most of the players, it is disabled by default. Default: false
drm = false

# Title of the webpage. Default: JioTV Go
title = ""

# Enable Or Disable URL Encryption. URL Encryption prevents hackers from injecting URLs into the server. Default: true
# If you think it is unnecessary, you can disable it. But it is recommended to enable it.
disable_url_encryption = false

# Folder Path for all JioTV Go related files. Default: "$HOME/.jiotv_go"
path_prefix = ""

# Proxy URL. Proxy is useful to bypass geo-restrictions and ip-restrictions for JioTV API. Default: ""
proxy = ""

# LogPath is the directory for log files. Default: "" (logs to default path like $HOME/.jiotv_go/jiotv_go.log)
log_path = ""

# LogToStdout controls logging to stdout/stderr. Default: false (when set in config)
log_to_stdout = false

# Path to custom channels JSON file. If empty, custom channels are disabled. Default: ""
custom_channels_path = ""
```

This example demonstrates how to customize the configuration parameters using TOML syntax. Feel free to modify the values based on your preferences and requirements.

### Example YAML Configuration

You can save the following configuration in a file named `jiotv_go.yaml`. 

The file is also available at [configs/jiotv_go-config.yaml](https://github.com/jiotv-go/jiotv_go/blob/main/configs/jiotv_go-config.yaml).

```yaml
epg: false
debug: false
disable_ts_handler: false
disable_logout: false
drm: false
title: ""
disable_url_encryption: false
path_prefix: ""
proxy: ""
log_path: ""
log_to_stdout: false
custom_channels_path: ""
```

### Example JSON Configuration

You can save the following configuration in a file named `jiotv_go.json`.

The file is also available at [configs/jiotv_go-config.json](https://github.com/jiotv-go/jiotv_go/blob/main/configs/jiotv_go-config.json).

```json
{
    "epg": false,
    "debug": false,
    "disable_ts_handler": false,
    "disable_logout": false,
    "drm": false,
    "title": "",
    "disable_url_encryption": false,
    "path_prefix": "",
    "proxy": "",
    "log_path": "",
    "log_to_stdout": false,
    "custom_channels_path": ""
}
```

## Custom Channels Configuration

JioTV Go supports adding custom channels alongside the official JioTV channels. Custom channels can be configured using a JSON file specified in the `custom_channels_path` configuration option.

### Custom Channels File Format

The custom channels file should be a JSON file with the following structure:

```json
{
  "channels": [
    {
      "id": "channel_identifier",
      "name": "Channel Display Name",
      "url": "https://example.com/stream.m3u8",
      "logo_url": "https://example.com/logo.png",
      "category": 12,
      "language": 6,
      "is_hd": true
    }
  ]
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | ✅ | Unique identifier for the channel. Should not contain spaces or special characters. |
| `name` | string | ✅ | Display name of the channel as it appears in the interface. |
| `url` | string | ✅ | Direct URL to the channel's HLS stream (usually .m3u8 file). |
| `logo_url` | string | ❌ | URL to the channel's logo image. Can be a full HTTP URL or a filename. |
| `category` | integer | ✅ | Category ID. Refer to the category mapping below. |
| `language` | integer | ✅ | Language ID. Refer to the language mapping below. |
| `is_hd` | boolean | ❌ | Whether the channel is HD quality. Defaults to `false`. |

### Category Mapping

| ID | Category |
|----|----------|
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

### Language Mapping

| ID | Language |
|----|----------|
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

### Example Custom Channels File

Here's a complete example of a custom channels file:

```json
{
  "channels": [
    {
      "id": "my_news",
      "name": "My News Channel",
      "url": "https://example.com/streams/news.m3u8",
      "logo_url": "https://example.com/logos/news.png",
      "category": 12,
      "language": 6,
      "is_hd": true
    },
    {
      "id": "sports_stream",
      "name": "Sports Stream",
      "url": "https://example.com/streams/sports.m3u8",
      "logo_url": "https://example.com/logos/sports.png",
      "category": 8,
      "language": 1,
      "is_hd": false
    },
    {
      "id": "entertainment",
      "name": "Entertainment Plus",
      "url": "https://example.com/streams/entertainment.m3u8",
      "category": 5,
      "language": 6,
      "is_hd": true
    }
  ]
}
```

### How to Use Custom Channels

1. **Create the JSON file**: Create a JSON file with your custom channels using the format described above.

2. **Configure the path**: Set the `custom_channels_path` in your JioTV Go configuration to point to your JSON file:
   ```toml
   custom_channels_path = "/path/to/your/custom-channels.json"
   ```

3. **Restart JioTV Go**: Restart the server to load the custom channels.

4. **Access channels**: Custom channels will appear in:
   - Web interface alongside JioTV channels
   - IPTV playlists with `custom_` prefix (e.g., `custom_my_news`)
   - Channel listings via API

### Important Notes

- **Channel IDs**: Custom channel IDs are automatically prefixed with `custom_` to avoid conflicts with official JioTV channels.
- **Direct streaming**: Custom channels stream directly from their source URLs without processing through JioTV Go servers.
- **Logo URLs**: If you provide full HTTP URLs for logos, they'll be used directly. If you provide just filenames, they'll be served through the JioTV Go image handler.
- **Validation**: Invalid channels (missing required fields) will be skipped with warning messages in the logs.
- **Disabling**: To disable custom channels, set `custom_channels_path` to an empty string or remove the configuration option.

### Troubleshooting

- **Channels not appearing**: Check the server logs for validation errors or file loading issues.
- **Stream not working**: Ensure the stream URLs are accessible and in HLS format (.m3u8).
- **Logo not showing**: Verify the logo URL is accessible and the image format is supported.

An example configuration file is available at `configs/custom-channels.example.json` in the JioTV Go repository.
