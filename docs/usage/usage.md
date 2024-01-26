# Usage

JioTV Go is a command-line application. It can be used to start the server, update JioTV Go and control certain aspects of the server.

## Command Line Interface

The `jiotv_go` CLI has the following structure:

```shell
jiotv_go command [command options]
```

## Commands

You can always use the `help` command or `-h` / `--help` flag to get help about a command.


## 1. Login Command

The `login` command helps you to login to JioTV Go. Alternatively, you can also login using the web interface at `http://localhost:5001/`.

```
jiotv_go login [command options] [arguments...]
```

#### USAGE

jiotv_go login [command options] [arguments...]

#### DESCRIPTION

The `login` command helps you to login to JioTV Go. It will ask for your JioTV credentials and save it to `jiotv_credentials_v2.json` file.

If you want to change your credentials, you can use the `login` command again. It will overwrite the existing credentials.

#### COMMANDS

- `otp`, `o`: Login with OTP
- `password`, `p`: Login with Password
- `reset`, `logout`, `lo`: Reset credentials. This will delete the existing credentials.
- `help`, `h`: Shows a list of commands or help for one command

### otp (o)

#### USAGE

jiotv_go login otp

#### DESCRIPTION

The `otp` command helps you to login to JioTV Go with OTP. It will ask for your JioTV number and send an OTP to your number. You have to enter the OTP to login.

### password (p)

#### USAGE

jiotv_go login password

#### DESCRIPTION

The `password` command helps you to login to JioTV Go with password. It will ask for your JioTV number and password to login.

### reset (logout, lo)

#### USAGE

jiotv_go login reset

#### DESCRIPTION

The `reset` command helps you to reset your credentials. This will delete the existing credentials. You have to login again to use JioTV Go.


## 2. Serve Command

The `serve` command starts the JioTV Go server.

```shell
jiotv_go serve [command options] [arguments...]
```

**Options:**

- `--config value, -c value`: Path to the configuration file.
  <br>By default, JioTV Go will look for a file named `jiotv_go.(toml|yaml|json)` or `config.(toml|yaml|json)` in the same directory as the binary or `$HOME/.jiotv_go/` directory.
- `--host value, -H value`: Host to listen on (default: "localhost").
- `--port value, -p value`: Port to listen on (default: "5001").
- `--public, -P`: Open the server to the public. This will expose your server outside your local network. Equivalent to passing `--host 0.0.0.0` (default: false).
- `--prefork`: Enable prefork. This will enable preforking the server to multiple processes. This is useful for production deployment (default: false).
- `--skip-update-check`: Skip checking for updates on startup (default: false).
- `--help, -h`: Show help for the `serve` command.

**Example:**

This will start the server on port 8080 and open it to the public.

```shell
jiotv_go serve --port 8080 --public
```

<div class="warning">
Use of the <code>--public</code> flag is not recommended. It exposes your server outside your local network. Use it only if it is necessary for you in some cases where you want to access JioTV Go server in your phone to TV or other devices.
</div>

## 3. Update Command

The `update` command updates JioTV Go to the latest version.

```
jiotv_go update
```

**Options:**

- `--version, -v`: Never use this flag, unless you know what you are doing. This will update JioTV Go to the specified version. This is useful for testing new features before release or to downgrade to a previous version. Supports all JioTV Go version above v3.0.0.

## 4. EPG Command

The `epg` command helps you to manage the EPG feature of JioTV Go.

```shell
jiotv_go epg [command options] [arguments...]
```

#### USAGE

jiotv_go epg command [command options]

#### DESCRIPTION

The `epg` command manages EPG. It can be used to generate EPG, regenerate EPG, and delete EPG.

#### COMMANDS

- `generate`, `gen`, `g`: Generate EPG
- `Delete`, `del`, `d`: Delete EPG
- `help`, `h`: Shows a list of commands or help for one command

### generate (gen, g)

#### USAGE

jiotv_go epg generate [command options] [arguments...]

#### DESCRIPTION

The `generate` command generates EPG by downloading the latest EPG from JioTV, and saving it to epg.xml.gz.

It will delete the existing EPG file if it exists. Once the EPG file is generated, it will be automatically updated by the server. If you want to disable it, use the `epg delete` command.

This is also shortcut method for enabling EPG than setting `epg` to `true` in the configuration file. Read the [EPG Config](../config.md#epg-electronic-program-guide) section for more information.

### delete (del, d)

#### USAGE

jiotv_go epg Delete [command options] [arguments...]

#### DESCRIPTION

The `delete` command deletes the existing EPG file if it exists. This will disable EPG on the server.


## 5. Help Command

The `help` command shows a list of commands or help for a specific command.

```
jiotv_go help [command]
```

**Example:**

```bash
jiotv_go help serve
```

## 6. Autostart Command for Unix

The `autostart` command helps you to setup JioTV Go to start automatically when terminal starts.

This is not recommended for devices other than Android Phone or TV.

```bash
jiotv_go autostart
```

**Options:**

- `-a value, --args value`: Options for the `serve`/`run`/`start` command as mentioned in the [Serve Command](#2-serve-command) section.

If you want to arguments for the `serve`/`run`/`start` command, you can pass `-a` flag enclose all the arguments in quotes.

For example if you want to run at port 8080 and pass a configuration file, you can use the following command:

```bash
jiotv_go autostart -a "--port 8080 --config config.toml"
```

<div class="warning">

> Auto detection of config files will only work if binary is in the same directory as the config file.

</div>

## 7. Background Command

The `background` command allows you to run the JioTV Go server in the background. It provides subcommands for starting and stopping the server in the background.

> Tip: `bg` is an alias for `background`.

#### USAGE

```shell
jiotv_go background [command options] [arguments...]
```

#### DESCRIPTION

The `background` command allows you to run the JioTV Go server in the background. It provides subcommands for starting and stopping the server in the background.

#### COMMANDS

- `start (run, r)`: Run JioTV Go server in the background
  ```shell
  jiotv_go background start [command options] [arguments...]
  ```
  - `--args value, -a value`: String value arguments passed to the `serve/run` command while running in the background as mentioned in the [Serve Command](#2-serve-command) section.

  Description: The `start` command starts the JioTV Go server in the background. It runs the `JioTVServer` function in a separate process.

- `stop (k, kill)`: Stop JioTV Go server running in the background
  ```shell
  jiotv_go background stop
  ```

  Description: The `stop` command stops the JioTV Go server running in the background. It will only work if the server is started using the `background start` command.

### Example:

```shell
jiotv_go background start
```

Example with arguments *(make sure to enclose the arguments in quotes)*:

```shell
jiotv_go background start --args "--port 8080 --config config.toml"
```

### Note:

- Make sure to stop the background server using the `stop` command when it is no longer needed.

## Support and Issues

For any issues or feature requests, please check the [GitHub repository](https://github.com/rabilrbl/jiotv_go) or create a new issue.

**Note:** Ensure that you have the necessary permissions and follow the terms of service when using JioTV Go.
