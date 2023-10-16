# Usage Guide

This section provides a detailed guide on how to use JioTV Go for various platforms and how to customize your experience. Whether you prefer to use JioTV Go natively, on Android, or with Docker, we've got you covered. Additionally, you'll find information on optional customizations and advanced features.

## Using JioTV Go Natively

1. **Download the Latest Binary:**
   - Start your JioTV Go journey by downloading the latest binary for your operating system from the [releases page](https://github.com/rabilrbl/jiotv_go/releases/latest).

2. **Linux/Unix Users - Grant Executable Permissions:**
   - For Linux/Unix users, grant executable permissions to the downloaded binary. Use the command `chmod +x jiotv_go-...`, replacing `jiotv_go-...` with the actual binary name you've downloaded.

3. **Run JioTV Go:**
   - Execute the binary with `./jiotv_go-...`.

4. **Access JioTV Go in Your Browser:**
   - Open your web browser and visit `http://localhost:5001`.

5. **Log In to Access Content:**
   - To access JioTV content, click on the Login button and enter your credentials.

6. **Explore Live TV Channels:**
   - Choose from a variety of channels and embark on your live TV adventure.

7. **IPTV Enthusiasts - Access M3U Playlist:**
   - If you're an IPTV enthusiast, you can access the M3U playlist by visiting `http://localhost:5001/playlist.m3u`.

## Android Users, We've Got You Covered!

1. **Install Termux:**
   - Dive into the Android world by first downloading [Termux](https://github.com/termux/termux-app/releases/latest).

2. **Execute Android Script:**
   - Open Termux and execute the command `DEBIAN_FRONTEND=noninteractive pkg update -y && pkg upgrade -y && pkg install curl openssl -y`.

3. **Download Android Script:**
   - Download the Android script by running `curl -Lo jiotv_go.sh https://raw.githubusercontent.com/rabilrbl/jiotv_go/main/android.sh`.

4. **Grant Executable Permissions:**
   - Grant executable permissions to the script with `chmod +x jiotv_go.sh`.

5. **Install JioTV Go:**
   - Execute the install script with `./jiotv_go.sh install`. The script will automatically download the latest binary for your device and install it.

6. **Start the Server:**
   - Start the server with `./jiotv_go.sh run`.

7. **Access JioTV Go in Your Browser:**
   - Open your web browser and visit `http://localhost:5001`.

8. **Update to the Latest Version:**
   - To update to the latest version, run `./jiotv_go.sh update`.

## Docker Enthusiasts, Here's Your Shortcut!

1. **Install Docker:**
   - If you're a Docker enthusiast, begin by installing [Docker](https://docs.docker.com/get-docker/).

2. **Run JioTV Go with Docker:**
   - Run the command:
     ```sh
     docker run -p 5001:5001 -v ./.jiotv_go/secrets:/app/secrets ghcr.io/rabilrbl/jiotv_go
     ```

3. **Access JioTV Go in Your Browser:**
   - Open your web browser and visit `http://localhost:5001`.

4. **Update to the Latest Version:**
   - To update to the latest version, run:
     ```sh
     docker pull ghcr.io/rabilrbl/jiotv_go:latest
     ```

## Optional Customizations

- Want to specify a custom port or host? No problem! Simply pass `host:port` or `:port` as an argument to the binary like this: `./jiotv_go "host:port"`. If you are using the Android script, you can pass the port as an argument to the script like this: `./jiotv_go.sh run "host:port".

- Disable TS Handler by setting the environment variable `JIOTV_DISABLE_TS_HANDLER=true` before running the binary. This can help reduce the load on the server and decrease latency, particularly on low-end devices.

- Customize the path or folder for your `credentials.json` by setting the environment variable `JIOTV_CREDENTIALS_PATH=/path` before running the binary.

## EPG (Electronic Program Guide)

- To enable EPG, set the environment variable `JIOTV_EPG=true` before running the binary. This will generate an EPG file at `/epg.xml.gz` for use in IPTV players. The M3U playlist will contain a link to this EPG file automatically.

- If enabled, the EPG file is scheduled to be updated at a random time during midnight.

## Proxy 🌐

If you want to use a proxy, set the environment variable `JIOTV_PROXY` before running the binary. Examples include Socks5 Proxy (`socks5://user:pass@host:port`) or All Other Proxy (`user:pass@host:port`).

## Remote Deployment Made Easy

In cases where remote server permissions prevent the creation of the `jiotv_credentials_v2.json` file, follow these steps:

1. On your local machine, log in to JioTV Go to generate your `jiotv_credentials_v2.json`.

2. Configure the following environment variables on your remote server:
   - `JIOTV_SSO_TOKEN` - The `ssoToken` from the credentials file.
   - `JIOTV_CRM` - The `crm` from the credentials file.
   - `JIOTV_UNIQUE_ID` - The `uniqueId` from the credentials file.

With these environment variables set, the `credentials.json` will be bypassed, and your JioTV Go deployment will proceed smoothly. Enjoy the journey!
