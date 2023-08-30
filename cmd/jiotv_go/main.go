package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rabilrbl/jiotv_go/internals/handlers"
	"github.com/rabilrbl/jiotv_go/internals/utils"
	"html/template"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()
	
	utils.Log = utils.GetLogger()
	
	r.StaticFS("/static", http.FS(staticEmbed))
	tmpl := template.Must(template.ParseFS(tmplEmbed, "templates/*"))
	r.SetHTMLTemplate(tmpl)
	
	// Initialize the television client
	handlers.Init()
	
	r.GET("/", handlers.IndexHandler)
	r.GET("/login", handlers.LoginHandler)
	r.POST("/login", handlers.LoginHandler)
	r.GET("/live/:id", handlers.LiveHandler)
	r.GET("/render", handlers.RenderHandler)
	r.GET("/renderKey", handlers.RenderKeyHandler)
	r.GET("/channels", handlers.ChannelsHandler)
	r.GET("/play/:id", handlers.PlayHandler)
	r.GET("/player/:id", handlers.PlayerHandler)
	r.GET("/clapper/:id", handlers.ClapprHandler)
	
	if len(os.Args) > 1 {
		r.Run(os.Args[1])
	} else {
		r.Run("localhost:5001")
	}
}
