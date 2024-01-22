package main

import (
	"fmt"
	"flag"
	"github.com/rabilrbl/jiotv_go/v2/cmd"
)

func main() {
	var config string
	var serve bool
	var host string
	var port string

	flag.StringVar(&config, "config", "", "Path to config file")

	flag.BoolVar(&serve, "serve", false, "Start JioTV Go server")

	flag.StringVar(&host, "host", "localhost", "Host to listen on")

	flag.StringVar(&port, "port", "5001", "Port to listen on")

	flag.Usage = func() {
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if serve {
		cmd.JioTVServer(host, port)
	}
	
	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

}
