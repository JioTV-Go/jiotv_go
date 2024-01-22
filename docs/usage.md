# Usage

JioTV Go is a command-line application. It can be used to start the server, update JioTV Go and control certain aspects of the server.

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
- `--host value, -H value`: Host to listen on (default: "localhost").
- `--port value, -p value`: Port to listen on (default: "5001").
- `--public, -P`: Open the server to the public. This will expose your server outside your local network. Equivalent to passing `--host 0.0.0.0` (default: false).
- `--prefork`: Enable prefork. This will enable preforking the server to multiple processes. This is useful for production deployment (default: false).
- `--help, -h`: Show help for the `serve` command.

**Example:**
```bash
jiotv_go serve --host 127.0.0.1 --port 8080 --public
```

#### 2. Update Command

The `update` command updates JioTV Go to the latest version.

```
jiotv_go update [command options] [arguments...]
```

**Options:**

- `--help, -h`: Show help for the `update` command.

**Example:**
```bash
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