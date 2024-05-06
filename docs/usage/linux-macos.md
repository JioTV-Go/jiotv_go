# Running JioTV Go on Linux and macOS

Once you have downloaded the [latest release of JioTV Go](../get_started.md#pre-built-binaries), follow these steps to run it seamlessly on Linux and macOS:

## Automatic Install

1. **Open Terminal:**
   - Launch the terminal on your Linux or macOS system.

   > Use [automatic install script](../get_started.md#automatic-installation-recommended) to install JioTV Go on Linux and macOS.

## Manual Install

> If you want to install manually, follow the steps below:

1. **Navigate to Downloaded Directory:**
   - Move to the directory where you have downloaded JioTV Go.

2. **Make the File Executable:**
   - Execute the following command to make the file executable:
   
     ```sh
     chmod +x jiotv_go-linux-{arch}
     ```

     Replace `{arch}` with your architecture. For example, if your architecture is `amd64`, use the command:

     ```sh
     chmod +x jiotv_go-linux-amd64
     ```

     If you are unsure about your architecture, check the [Identify your architecture](../get_started.md#identifying-your-os-and-architecture) section in the [Get Started](../get_started.md) page.

3. **Run JioTV Go:**
   - Start JioTV Go by running the following command:

     ```sh
     ./jiotv_go-linux-{arch} serve
     ```

4. **Access the Server:**
   - Open your web browser and go to [http://localhost:5001/](http://localhost:5001/) to access JioTV Go.
  
## JioTV Go as a systemd service (Autostart)

This guide walks you through setting up JioTV Go systemd services on Linux.

Ensure `jiotv_go` is installed. If necessary modify the service file `ExecStart` lines to point to alternative paths.

### 1. Create a service specific user

For security it is best to run a service with a specific user and group.
You can create one using the following command:

```console
sudo adduser --system  --gecos "JioTV Go Service" --disabled-password --group --home /var/lib/jiotv_go jiotv_go
```

This creates a new system user and group named `jiotv_go` with no login access and home directory `/var/lib/jiotv_go` which will be the default location for the config files.

In addition you can add to the `jiotv_go` group any users you wish to be able to easily manage or access files, for example:

```console
sudo adduser <username> jiotv_go
```

### 2. Daemon (jiotv_go) service

Create the file `/etc/systemd/system/jiotv_go.service` containing the following:

```console
[Unit]
Description=JioTV Go Client Daemon
After=network-online.target

[Service]
Type=simple
UMask=007

ExecStart=/usr/bin/jiotv_go serve -P

Restart=on-failure

# Time to wait before forcefully stopped.
TimeoutStopSec=300

[Install]
WantedBy=multi-user.target
```

The `ExecStart` line needs to point to the exact installation path of the `jiotv_go` program.

Note: Please make appropriate changes to the `ExecStart` line, the current one serves to the public.

### 3. User configuration

To run the service using the previously created user e.g. `jiotv_go`, first create the service configuration directory:

```console
sudo mkdir /etc/systemd/system/jiotv_go.service.d/
```

Then create a user file `/etc/systemd/system/jiotv_go.service.d/user.conf` with the following contents:

```console
# Override service user
[Service]
User=jiotv_go
Group=jiotv_go
```

### 4. Start jiotv_go service

Now enable it to start up on boot, start the service and verify it is running:
```console
sudo systemctl enable /etc/systemd/system/jiotv_go.service
sudo systemctl start jiotv_go
sudo systemctl status jiotv_go
```

Enjoy your JioTV Go streaming experience on Linux and macOS! If you encounter any issues or have questions, refer to the [Support and Issues](#support-and-issues) section in the user guide or visit the [GitHub repository](https://github.com/rabilrbl/jiotv_go).
