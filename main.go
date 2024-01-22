package main

import (
	"log"
	"github.com/rabilrbl/jiotv_go/v2/cmd"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
        Name:  "JioTV Go",
        Usage: "Stream JioTV on any device",
		HelpName: "jiotv_go",
		Version: "v3.0.0",
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"run", "start"},
				Usage:   "Start JioTV Go server",
				Action: func(c *cli.Context) error {
					host := c.String("host")
					port := c.String("port")
					return cmd.JioTVServer(host, port)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "host",
						Aliases: []string{"H"},
						Value: "localhost",
						Usage: "Host to listen on",
					},
					&cli.StringFlag{
						Name:  "port",
						Aliases: []string{"p"},
						Value: "5001",
						Usage: "Port to listen on",
					},
				},
			},
		},
		
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}
