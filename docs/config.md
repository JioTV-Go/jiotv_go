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

### Custom Channels:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Path to custom channels configuration file. | `custom_channels_file` | `JIOTV_CUSTOM_CHANNELS_FILE` | `""` (empty string) |

This option specifies the path to a JSON or YAML file containing custom channel definitions that will be integrated with JioTV channels. Custom channels will appear in the web interface and IPTV playlists alongside standard JioTV channels. If the file is not found or contains errors, the server will continue to work with only JioTV channels.

For detailed information about custom channels configuration, including file format, field descriptions, and usage examples, please see [Custom Channels Documentation](./CUSTOM_CHANNELS.md).

### Default Categories and Languages:

| Purpose | Config Value | Environment Variable | Default |
| ----- | ------------ | -------------------- | ------- |
| Default categories to display on the web interface when no filters are applied. | `default_categories` | `JIOTV_DEFAULT_CATEGORIES` | `[]` (empty array) |
| Default languages to display on the web interface when no filters are applied. | `default_languages` | `JIOTV_DEFAULT_LANGUAGES` | `[]` (empty array) |

These options allow you to configure which categories and languages should be shown by default on the web interface when users haven't applied any filters. This provides a more curated experience while still allowing users to override these defaults through the filter interface.

**Default Categories**: An array of category IDs to display by default. For example, `[8, 5]` would show only Sports and Entertainment channels by default.

**Default Languages**: An array of language IDs to display by default. For example, `[1, 6]` would show only Hindi and English language channels by default.

**Filtering Logic**: 
- When both arrays are configured, channels must match at least one configured category AND one configured language
- When only one array is configured, channels are filtered by that criteria only
- When both arrays are empty or not configured, all channels are displayed (backward compatible behavior)
- User interactions with filter dropdowns completely override the default configuration

**Example Use Cases**:
- Show only Entertainment and Movies channels in Hindi and English: `default_categories = [5, 6]`, `default_languages = [1, 6]`
- Show all Sports channels regardless of language: `default_categories = [8]`, `default_languages = []`
- Show all Hindi content regardless of category: `default_categories = []`, `default_languages = [1]`
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

# CustomChannelsFile is the path to custom channels configuration file. Default: ""
custom_channels_file = ""

# Default categories to display on the web interface when no filters are applied. Array of category IDs. Default: []
# Example: default_categories = [8, 5] # Sports, Entertainment
default_categories = []

# Default languages to display on the web interface when no filters are applied. Array of language IDs. Default: []
# Example: default_languages = [1, 6] # Hindi, English
default_languages = []
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
custom_channels_file: ""
default_categories: []
default_languages: []
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
    "custom_channels_file": "",
    "default_categories": [],
    "default_languages": []
}
```
