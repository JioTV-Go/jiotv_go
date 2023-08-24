# JioTV Go

JioTV Go is a web application that allows you to watch Live TV channels. This project helps you to watch JioTV without the JioTV App. The project is currently in development and is made for educational purposes only.

Download latest binary for your operating system from [here](https://github.com/rabilrbl/jiotv_go/releases/latest) and refer following api endpoints to use it.

## API Endpoints

| Endpoint | Description |
| --- | --- |
| `/` | Home Page |
| `/login?username=<username>&password=<password>` | Login to JioTV (Mandatory). If you forgot your password, you can use the [JioTV Password Recovery](https://www.jio.com/selfcare/signup/forgot-password) page to reset your password. |
| `/channels` | List of all channels |
| `/channels?type=m3u` | List of all channels in m3u format for IPTV and VLC Media Player |
| `/live/:channel_id` | Watch Live TV |

## Installation

JioTV Go requires [Golang](https://golang.org/) to run.

Install the dependencies and start the server.

```sh
go mod download
go run .
```

## License

GNU GENERAL PUBLIC LICENSE v3.0

**Free Software, Hell Yeah!**. The project is completely open-source and free to use. Any attempt to sell this project will be considered as a violation of the license and will be taken down immediately. If you notice any such activity, please report it to [me](mailto:rabil@rbls.eu.org).
