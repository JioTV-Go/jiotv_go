package main

import (
	"github.com/rabilrbl/jiotv_go/v2/cmd"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "JioTV Go",
		Usage:    "Stream JioTV on any device",
		HelpName: "jiotv_go",
		Version:  "v3.0.0",
		Copyright: "© JioTV Go by Mohammed Rabil (https://github.com/rabilrbl/jiotv_go)",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "serve",
				Aliases:     []string{"run", "start"},
				Usage:       "Start JioTV Go server",
				Description: "The serve command starts JioTV Go server, and listens on the host and port. The default host is localhost and port is 5001.",
				Action: func(c *cli.Context) error {
					host := c.String("host")
					port := c.String("port")
					return cmd.JioTVServer(host, port)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Aliases: []string{"H"},
						Value:   "localhost",
						Usage:   "Host to listen on",
					},
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   "5001",
						Usage:   "Port to listen on",
					},
				},
			},
			{
				Name:        "update",
				Aliases:     []string{"upgrade", "u"},
				Usage:       "Update JioTV Go to latest version",
				Description: "The update command updates JioTV Go by identifying the operating system and architecture, downloading the latest release from GitHub, and replacing the current binary with the latest one.",
				Action: func(c *cli.Context) error {
					return cmd.Update()
				},
			},
		},
		CommandNotFound: func(c *cli.Context, command string) {
			log.Printf("Command '%s' not found.\n", command)
			// Print help for invalid commands
			cli.ShowAppHelpAndExit(c, 3)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
