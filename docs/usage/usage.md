# Usage

JioTV Go is a command-line application. It can be used to start the server, update JioTV Go and control certain aspects of the server.

<div class="warning">

Android users, if you face DNS Lookup errors, please read the [Note for Termux Users](./termux.md) page.

</div>

## Command Line Interface

The `jiotv_go` CLI has the following structure:

```
jiotv_go [global options] command [command options] 
```

### Global Options

- `--help, -h`: Show help for the CLI.
- `--version, -v`: Print the version of the CLI.

### Commands

#### 1. Serve Command

The `serve` command starts the JioTV Go server.

```
jiotv_go serve [command options] [arguments...]
```

**Options:**

- `--config value, -c value`: Path to the configuration file. 
  <br>By default, JioTV Go will look for a file named `jiotv_go.(toml|yaml|json)` or `config.(toml|yaml|json)` in the same directory as the binary.
- `--host value, -H value`: Host to listen on (default: "localhost").
- `--port value, -p value`: Port to listen on (default: "5001").
- `--public, -P`: Open the server to the public. This will expose your server outside your local network. Equivalent to passing `--host 0.0.0.0` (default: false).
- `--prefork`: Enable prefork. This will enable preforking the server to multiple processes. This is useful for production deployment (default: false).
- `--help, -h`: Show help for the `serve` command.

**Example:**

This will start the server on port 8080 and open it to the public.

```bash
jiotv_go serve --port 8080 --public
```

<div class="warning">
Use of the <code>--public</code> flag is not recommended. It exposes your server outside your local network. Use it only if it is necessary for you in some cases where you want to access JioTV Go server in your phone to TV or other devices.
</div>


#### 2. Update Command

The `update` command updates JioTV Go to the latest version.

```
jiotv_go update
```

#### 3. Help Command

The `help` command shows a list of commands or help for a specific command.

```
jiotv_go help [command]
```

**Example:**
```bash
jiotv_go help serve
```

## Support and Issues

For any issues or feature requests, please check the [GitHub repository](https://github.com/rabilrbl/jiotv_go) or create a new issue.

**Note:** Ensure that you have the necessary permissions and follow the terms of service when using JioTV Go.