package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rabilrbl/jiotv_go/internals/handlers"
	"github.com/rabilrbl/jiotv_go/internals/utils"
	"github.com/rabilrbl/jiotv_go/internals/middleware"
	"html/template"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()
	r.Use(middleware.CORS())
	
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
	r.GET("/render.m3u8", handlers.RenderHandler)
	r.GET("/render.key", handlers.RenderKeyHandler)
	r.GET("/channels", handlers.ChannelsHandler)
	r.GET("/playlist.m3u", handlers.PlaylistHandler)
	r.GET("/play/:id", handlers.PlayHandler)
	r.GET("/player/:id", handlers.PlayerHandler)
	r.GET("/clappr/:id", handlers.ClapprHandler)
	r.GET("/favicon.ico", handlers.FaviconHandler)
	
	if len(os.Args) > 1 {
		r.Run(os.Args[1])
	} else {
		r.Run("localhost:5001")
	}
}
