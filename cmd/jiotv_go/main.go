package main

import (
	"net/http"
	"os"

	"github.com/rabilrbl/jiotv_go/v2/internal/handlers"
	"github.com/rabilrbl/jiotv_go/v2/internal/middleware"
	"github.com/rabilrbl/jiotv_go/v2/pkg/epg"
	"github.com/rabilrbl/jiotv_go/v2/pkg/utils"
	"github.com/rabilrbl/jiotv_go/v2/web"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/template/html/v2"
)

func main() {
	utils.Log = utils.GetLogger()

	// if os.Getenv("JIOTV_DEBUG") == "true" or file epg.xml.gz exists
	if os.Getenv("JIOTV_EPG") == "true" || utils.FileExists("epg.xml.gz") {
		go epg.Init()
	}

	engine := html.NewFileSystem(http.FS(web.GetViewFiles()), ".html")
	if os.Getenv("JIOTV_DEBUG") == "true" {
		engine.Reload(true)
	}

	app := fiber.New(fiber.Config{
		Views:             engine,
		Prefork:           os.Getenv("JIOTV_PREFORK") == "true",
		StreamRequestBody: true,
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

	app.Use(logger.New(logger.Config{
		TimeZone: "Asia/Kolkata",
		Format:   "[${time}] ${status} - ${latency} ${method} ${path} Params:[${queryParams}] ${error}\n",
		Output:   utils.Log.Writer(),
	}))

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(web.GetStaticFiles()),
		PathPrefix: "static",
		Browse:     false,
	}))

	app.Use(helmet.New())

	// Initialize the television object
	handlers.Init()

	app.Get("/", handlers.IndexHandler)
	app.Post("/login/sendOTP", handlers.LoginSendOTPHandler)
	app.Post("/login/verifyOTP", handlers.LoginVerifyOTPHandler)
	app.Post("/login", handlers.LoginPasswordHandler)
	app.Get("/logout", handlers.LogoutHandler)
	app.Get("/live/:id", handlers.LiveHandler)
	app.Get("/live/:quality/:id", handlers.LiveQualityHandler)
	app.Get("/render.m3u8", handlers.RenderHandler)
	app.Get("/render.ts", handlers.RenderTSHandler)
	app.Get("/render.key", handlers.RenderKeyHandler)
	app.Get("/channels", handlers.ChannelsHandler)
	app.Get("/playlist.m3u", handlers.PlaylistHandler)
	app.Get("/play/:id", handlers.PlayHandler)
	app.Get("/player/:id", handlers.PlayerHandler)
	app.Get("/clappr/:id", handlers.ClapprHandler)
	app.Get("/favicon.ico", handlers.FaviconHandler)
	app.Get("/jtvimage/:file", handlers.ImageHandler)
	app.Get("/epg.xml.gz", handlers.EPGHandler)

	addr := "localhost:5001"

	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	err := app.Listen(addr)
	if err != nil {
		utils.Log.Fatal(err)
	}
}
