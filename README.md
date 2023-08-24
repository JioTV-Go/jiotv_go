# JioTV Go

JioTV Go is a web application that allows you to watch Live TV channels. This project helps you to watch JioTV without the JioTV App. The project is currently in development and is made for educational purposes only.

Download the latest binary for your operating system from [here](https://github.com/rabilrbl/jiotv_go/releases/latest) and refer to [API endpoints](#api-endpoints) to use it.

## Steps to use JioTV Go on VLC Media Player or any IPTV Player

1. Download the latest binary for your operating system from the releases page.
2. Give executable permission to the binary.
3. Run the binary.
4. Open `http://localhost:5001` in your browser. Expect a success message.
5. Login to JioTV by opening `http://localhost:5001/login?username=<username>&password=<password>` in your browser. Expect a JSON response with some credentials.
6. Open `http://localhost:5001/channels?type=m3u` on your browser and download the m3u file.
7. Open the m3u file in VLC Media Player or any IPTV Player.
8. Enjoy Live TV.

## Screenshots

### Playing Live TV on VLC Media Player

![Alt text](assets/image.png)

## API Endpoints

| Endpoint | Description |
| --- | --- |
| `/` | Home Page |
| `/login?username=<username>&password=<password>` | Login to JioTV (Mandatory). If you forgot your password, you can use the [JioTV Password Recovery](https://www.jio.com/selfcare/signup/forgot-password) page to reset your password. |
| `/channels` | List of all channels |
| `/channels?type=m3u` | List of all channels in m3u format for IPTV and VLC Media Player |
| `/live/:channel_id` | Watch Live TV |

## Usage from Source

JioTV Go requires [Golang](https://golang.org/) to run.

Install the dependencies and start the server.

```sh
go mod download
go run .
```

## License

Attribution-NonCommercial 4.0 International (CC BY-NC 4.0)

**Free Software, Hell Yeah!**. The project is open-source and free to use. Any attempt to sell this project will be considered a violation of the license and will be taken down immediately. If you notice any such activity, please report it to [me](mailto:rabil@rbls.eu.org).
