package main

import (
	"os"
	"log"
	"github.com/gin-gonic/gin"
)

var Log *log.Logger

func main() {
	Log = log.New(os.Stdout, "", log.LstdFlags)
	r := gin.Default()
	r.GET("/", indexHandler)
	r.GET("/login", loginHandler)
	r.GET("/live/:id", getLive)
	r.Run(":8080")
}
