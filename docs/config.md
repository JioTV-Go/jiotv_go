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
| Path to the custom channels JSON file. | `custom_channels_path` | `JIOTV_CUSTOM_CHANNELS_PATH` | `""` (empty string) |

This option allows users to define their own custom channels by providing a path to a JSON file. If a valid path is provided, JioTV Go will load these channels alongside the default API channels. If the path is empty or the file is invalid, only API channels will be loaded.

#### Custom Channels JSON File Format

The JSON file specified by `custom_channels_path` should contain a single object with a key named `"channels"`. The value of this key should be an array of custom channel objects. Each object in the array represents a custom channel and can have the following fields:

*   `ID` (string, required): A unique identifier for your custom channel (e.g., `"custom_mycoolchannel"`). It's recommended to prefix with `"custom_"` to avoid potential clashes with API channel IDs.
*   `Name` (string, required): The display name for your channel (e.g., `"My Cool Channel"`).
*   `LogoURL` (string, required): The URL to the channel's logo. This can be an absolute URL (e.g., `"http://example.com/logo.png"`) or a relative path if you plan to serve it locally (though absolute URLs are generally more straightforward for external logos). If relative, it's relative to where your IPTV client might interpret it, or if used in the JioTV Go web UI, it might need specific handling (e.g., placing it in a publicly accessible folder).
*   `Category` (string, required): The category of the channel (e.g., `"Movies"`, `"Sports"`). This should match one of the category names known to JioTV Go. You can usually find the list of available categories in the application's UI or by referring to the `CategoryMap` in the source code (`pkg/television/types.go`). If the category is not recognized, it will default to "All Categories".
*   `Language` (string, required): The language of the channel (e.g., `"English"`, `"Hindi"`). Similar to `Category`, this should match a language name known to JioTV Go (see `LanguageMap` in `pkg/television/types.go`). If the language is not recognized, it will default to "Other".
*   `URL` (string, required): The direct M3U8 or other stream URL for the channel (e.g., `"http://example.com/stream.m3u8"`).
*   `EPGID` (string, optional): An identifier for mapping this channel to an Electronic Program Guide (EPG) source. This is for advanced use and may require further configuration depending on your EPG setup. Defaults to `""` (empty string) if omitted.

**Example `custom_channels.json` file:**

```json
{
  "channels": [
    {
      "ID": "custom_1",
      "Name": "My Custom HD Channel",
      "LogoURL": "http://example.com/logo.png",
      "Category": "Movies",
      "Language": "English",
      "URL": "http://example.com/stream.m3u8",
      "EPGID": "mychannel.epg"
    },
    {
      "ID": "custom_2",
      "Name": "Another Channel (SD)",
      "LogoURL": "http://example.com/another_logo.png",
      "Category": "News",
      "Language": "Hindi",
      "URL": "http://example.com/another_stream.m3u8"
    }
  ]
}
```

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

# Path to custom channels JSON file. Default: ""
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
