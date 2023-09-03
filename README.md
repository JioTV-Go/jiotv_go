# JioTV Go 🌟

Introducing JioTV Go: Your gateway to streaming Live TV channels on the web and IPTV! 📺🌐 Say goodbye to the JioTV App, as this project lets you enjoy JioTV's offerings hassle-free. Currently in the development stage, it's designed purely for educational exploration.

Ready to dive in? Download the latest binary for your operating system from [here](https://github.com/rabilrbl/jiotv_go/releases/latest), and explore the [API endpoints](#api-endpoints) to start your JioTV Go adventure!  🚀📡🔥

## Table of Contents

- [JioTV Go 🌟](#jiotv-go-)
  - [Table of Contents](#table-of-contents)
  - [Usage Guide 🚀](#usage-guide-)
    - [**Using JioTV Go Natively**](#using-jiotv-go-natively)
    - [**Android Users, We've Got You Covered!**](#android-users-weve-got-you-covered)
    - [**Docker Enthusiasts, Here's Your Shortcut!**](#docker-enthusiasts-heres-your-shortcut)
    - [**Optional Customizations**](#optional-customizations)
    - [**Remote Deployment Made Easy**](#remote-deployment-made-easy)
  - [Explore JioTV Go's Paths and Endpoints 🌐](#explore-jiotv-gos-paths-and-endpoints-)
    - [**Web Paths**](#web-paths)
    - [**API Endpoints**](#api-endpoints)
  - [Sneak Peek with Screenshots 📷](#sneak-peek-with-screenshots-)
    - [**Home Page**](#home-page)
    - [**Player Page**](#player-page)
    - [**Login Dialog**](#login-dialog)
    - [**Playing Live TV on VLC Media Player**](#playing-live-tv-on-vlc-media-player)
  - [Ready to Contribute? Let's Dive into Development! 🚀](#ready-to-contribute-lets-dive-into-development-)
  - [Customize the Look with TailwindCSS 🎨](#customize-the-look-with-tailwindcss-)
  - [Building JioTV Go from Source 🛠️](#building-jiotv-go-from-source-️)
  - [Let's Make JioTV Go Better Together! 🤝](#lets-make-jiotv-go-better-together-)
    - [**Report Bugs**](#report-bugs)
    - [**Ready to Contribute? Join the Journey! 🚀**](#ready-to-contribute-join-the-journey-)
  - [**License: Attribution-NonCommercial 4.0 International (CC BY-NC 4.0)**](#license-attribution-noncommercial-40-international-cc-by-nc-40)

## Usage Guide 🚀

### **Using JioTV Go Natively**

1. Start your journey by downloading the latest binary for your operating system from the [releases page](https://github.com/rabilrbl/jiotv_go/releases/latest).

2. For Linux/Unix users, grant executable permissions to the downloaded binary. Use the command `chmod +x jiotv_go-...`, replacing `jiotv_go-...` with the actual binary name you've downloaded.

3. Run the binary with `./jiotv_go-...`.

4. Fire up your favorite web browser and visit `http://localhost:5001`.

5. To access JioTV content, click on the Login button and enter your credentials.

6. Choose from a variety of channels and embark on your live TV adventure!

7. If you're an IPTV enthusiast, you can access the M3U playlist by visiting `http://localhost:5001/playlist.m3u`.

### **Android Users, We've Got You Covered!**

1. Dive into the Android world by first downloading [Termux](https://github.com/termux/termux-app/releases/latest).

2. Open Termux and execute the command `pkg install wget openssl -y`.

3. Copy the download link of the `jiotv_go-linux-(arch)` binary from the [releases page](https://github.com/rabilrbl/jiotv_go/releases/latest). Typically, choose the arm64 version for your aarch.

4. Download the file using `wget <link>`, where `<link>` is the one you copied in the previous step.

5. Now, follow the instructions for using [JioTV Go natively (mentioned above)](#using-jiotv-go-natively) starting from step 2.

### **Docker Enthusiasts, Here's Your Shortcut!**

1. If you're a Docker enthusiast, begin by installing [Docker](https://docs.docker.com/get-docker/).

2. Run the command `docker run -p 5001:5001 -v ./.jiotv_go/secrets:/app/secrets ghcr.io/rabilrbl/jiotv_go`.

3. Open your web browser and visit `http://localhost:5001`.

### **Optional Customizations**

- Want to specify a custom port or host? No problem! Simply pass `host:port` or `:port` as an argument to the binary like this: `./jiotv_go "host:port"`.

- If you prefer a specific path or folder for your `credentials.json`, set the environment variable `JIOTV_CREDENTIALS_PATH=/path` before running the binary.

- Need to use a custom DNS? Set the environment variable `JIOTV_DNS=1.1.1.1` before launching the binary.

### **Remote Deployment Made Easy**

In cases where remote server permissions prevent the creation of the `credentials.json` file, follow these steps:

1. On your local machine, log in to JioTV to generate your `credentials.json`.

2. Configure the following environment variables on your remote server:
   - `JIOTV_SSO_TOKEN` - The `ssoToken` from your `credentials.json`.
   - `JIOTV_CRM` - The `crm` from your `credentials.json`.
   - `JIOTV_UNIQUE_ID` - The `uniqueId` from your `credentials.json`.

With these environment variables set, the `credentials.json` will be bypassed, and your JioTV Go deployment will proceed smoothly. Enjoy the journey!

## Explore JioTV Go's Paths and Endpoints 🌐

### **Web Paths**

- **`/`**: The gateway to the Home Page, where your JioTV Go adventure begins.

- **`/play/:channel_id`**: Dive into the world of specific channels with ID `channel_id`.

- **`/playlist.m3u`**: Instantly obtain an M3U playlist for IPTV. (Redirects to `/channels?type=m3u` for your convenience.)

- **`/player/:channel_id`**: Immerse yourself with the default player (Flowplayer) for Channel `channel_id`.

- **`/clappr/:channel_id`**: Experience the magic of the Clappr player for Channel `channel_id`.

### **API Endpoints**

- **`/login?username=<username>&password=<password>`**: Forgot your password? No worries, visit the [JioTV Password Recovery](https://www.jio.com/selfcare/signup/forgot-password) page to reset it. You can also perform a POST request to this endpoint with data in the body.

- **`/channels`**: Discover the complete list of available channels.

- **`/channels?type=m3u`**: Effortlessly acquire an M3U playlist for IPTV.

- **`/live/:channel_id`**: Tune in to live TV with the specified channel ID.

## Sneak Peek with Screenshots 📷

### **Home Page**

![Home Page](./assets/home.webp)

### **Player Page**

![Player Page](./assets/player.webp)

### **Login Dialog**

![Login Page](./assets/login.webp)

### **Playing Live TV on VLC Media Player**

![Playing Live TV on VLC Media Player](./assets/image.webp)

## Ready to Contribute? Let's Dive into Development! 🚀

JioTV Go is powered by [Golang](https://golang.org/), so let's get started! Here's how to set up and run the server:

1. First, make sure you have Golang installed on your system.

2. Enable debug mode with the command:
   ```sh
   export GO_ENV=development
   ```

3. Install the necessary dependencies using:
    ```sh
    go mod download
    ```

4. Finally, fire up the server with:
   ```sh
   go run ./cmd/jiotv_go
    ```

That's it! You're now all set to explore and contribute to JioTV Go. Happy coding! 🖥️👩‍💻👨‍💻

## Customize the Look with TailwindCSS 🎨

At JioTV Go, we use the versatile [TailwindCSS](https://tailwindcss.com/) to style our project. If you're eager to make some style enhancements, here's how you can do it:

1. Ensure you have NodeJS installed on your system.

2. Open a new terminal window and navigate to the project's root directory.

3. Install the necessary dependencies by running:
    ```sh
      npm install
    ```

4. To keep TailwindCSS up to date with your changes in real-time, run the following command:
    ```sh
    npm run watch
    ```

5. Once you're satisfied with your style modifications, it's time to build the minified CSS file:
    ```sh
    npm run build
    ```

## Building JioTV Go from Source 🛠️

Building JioTV Go from source is a breeze! Follow these simple steps:

1. Open your terminal and navigate to the project's root directory.

2. Run the following command to build JioTV Go:
    ```sh
    go build./cmd/jiotv_g
    ```

## Let's Make JioTV Go Better Together! 🤝

### **Report Bugs**

Found a pesky bug? No worries! Please help us improve JioTV Go by creating an issue [here](https://github.com/rabilrbl/jiotv_go/issues/new). Be sure to include detailed steps to reproduce the bug, describe the expected behavior, and, if possible, attach screenshots. Your feedback is invaluable!

### **Ready to Contribute? Join the Journey! 🚀**

We wholeheartedly welcome your contributions. If you have ideas, fixes, or enhancements in mind, don't hesitate to create a pull request with your changes. For significant alterations, start by creating an issue to discuss your plans with us. Together, we can make JioTV Go even more incredible.

## **License: Attribution-NonCommercial 4.0 International (CC BY-NC 4.0)**

**Embrace the Spirit of Free Software!** JioTV Go is open-source and free to use. We're committed to keeping it accessible to everyone. If you come across any unauthorized attempts to sell this project, please report them to [me](mailto:rabil@rbls.eu.org) so we can take swift action. Your support is essential in safeguarding our project's values. 🙌📜💼
