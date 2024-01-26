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

Debug mode enables additional logging and debugging features for developers. It is recommended to disable debug mode if you are not a developer.

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

## Example Configurations

Below are example configuration file for JioTV Go. All fields are optional, and the values shown are the default settings:

You can also specify the path to the configuration file using the `--config` flag.

### Example TOML Configuration

You can save the following configuration in a file named `jiotv_go.toml`. JioTV Go will automatically load the configuration from this file if it is present in the same directory as the binary.

The file is also available at [configs/jiotv_go-config.toml](https://github.com/rabilrbl/jiotv_go/blob/main/configs/jiotv_go-config.toml).

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
```

This example demonstrates how to customize the configuration parameters using TOML syntax. Feel free to modify the values based on your preferences and requirements.

### Example YAML Configuration

You can save the following configuration in a file named `jiotv_go.yaml`. 

The file is also available at [configs/jiotv_go-config.yaml](https://github.com/rabilrbl/jiotv_go/blob/main/configs/jiotv_go-config.yaml).

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
```

### Example JSON Configuration

You can save the following configuration in a file named `jiotv_go.json`.

The file is also available at [configs/jiotv_go-config.json](https://github.com/rabilrbl/jiotv_go/blob/main/configs/jiotv_go-config.json).

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
    "proxy": ""
}
```
