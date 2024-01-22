# Config

The `config` package provides configuration settings for JioTV Go, a server that serves JioTV API content. This documentation outlines the available configuration options and how to customize them using various file formats, including TOML, YAML, and JSON.

## Configuration Options

### EPG (Electronic Program Guide):
  Enable or disable EPG generation. Default: `false`.

### Debug Mode:
  Enable or disable debug mode. Default: `false`.

### TS Handler:
  Enable or disable TS Handler. When enabled, the server serves TS files directly from the JioTV API. Default: `false`.

### Logout Feature:
  Enable or disable the logout feature. Default: `true`.

### DRM (Digital Rights Management):
  Enable or disable DRM. As DRM is not supported by most players, it is disabled by default. Default: `false`.

### Title:
  Title of the webpage. Default: "JioTV Go".

### URL Encryption:
  Enable or disable URL encryption. URL encryption prevents hackers from injecting URLs into the server. Default: `true`.

### Credentials Path:
  Path to the credentials file. Default: "credentials.json".

### Proxy:
  Proxy URL. Useful for bypassing geo-restrictions and IP restrictions for JioTV API. Default: "".

## Example Configuration (TOML)

Here is an example TOML configuration file for JioTV Go. All fields are optional, and the values shown are the default settings:

You can save the following configuration in a file named `jiotv_go.toml`. JioTV Go will automatically load the configuration from this file if it is present in the same directory as the binary. You can also specify the path to the configuration file using the `--config` flag.

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

# Path to the custom credentials file. 
credentials_path = ""

# Proxy URL. Proxy is useful to bypass geo-restrictions and ip-restrictions for JioTV API. Default: ""
proxy = ""
```

This example demonstrates how to customize the configuration parameters using TOML syntax. Feel free to modify the values based on your preferences and requirements.