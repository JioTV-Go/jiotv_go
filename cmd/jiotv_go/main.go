package main

import (
	"embed"
	"net/http"
	"os"

	"github.com/rabilrbl/jiotv_go/internals/handlers"
	"github.com/rabilrbl/jiotv_go/internals/middleware"
	"github.com/rabilrbl/jiotv_go/internals/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

//go:embed views/*
var viewFiles embed.FS

//go:embed static/*
var staticFiles embed.FS

func main() {

	engine := html.NewFileSystem(http.FS(viewFiles), ".html")
	if os.Getenv("DEBUG") == "true" {
		engine.Reload(true)
	}

	app := fiber.New(fiber.Config{
		Views:             engine,
		CaseSensitive:     false,
		StrictRouting:     false,
		EnablePrintRoutes: false,
		ServerHeader:      "JioTV Go",
		AppName:           "JioTV Go",
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(middleware.CORS())

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFiles),
		PathPrefix: "static",
		Browse:     false,
	}))

	utils.Log = utils.GetLogger()

	// Initialize the television object
	handlers.Init()

	app.Get("/", handlers.IndexHandler)
	app.Post("/login/sendOTP", handlers.LoginSendOTPHandler)
	app.Post("/login/verifyOTP", handlers.LoginVerifyOTPHandler)
	app.Get("/live/:id", handlers.LiveHandler)
	app.Get("/render.m3u8", handlers.RenderHandler)
	app.Get("/render.key", handlers.RenderKeyHandler)
	app.Get("/channels", handlers.ChannelsHandler)
	app.Get("/playlist.m3u", handlers.PlaylistHandler)
	app.Get("/play/:id", handlers.PlayHandler)
	app.Get("/player/:id", handlers.PlayerHandler)
	app.Get("/clappr/:id", handlers.ClapprHandler)
	app.Get("/favicon.ico", handlers.FaviconHandler)

	addr := "localhost:5001"

	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	app.Listen(addr)
}
