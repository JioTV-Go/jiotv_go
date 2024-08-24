package cmd

import (
	"net/http"

	"github.com/rabilrbl/jiotv_go/v3/internal/config"
	"github.com/rabilrbl/jiotv_go/v3/internal/handlers"
	"github.com/rabilrbl/jiotv_go/v3/internal/middleware"
	"github.com/rabilrbl/jiotv_go/v3/pkg/epg"
	"github.com/rabilrbl/jiotv_go/v3/pkg/scheduler"
	"github.com/rabilrbl/jiotv_go/v3/pkg/secureurl"
	"github.com/rabilrbl/jiotv_go/v3/pkg/store"
	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
	"github.com/rabilrbl/jiotv_go/v3/web"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

// JioTVServer starts the JioTV server.
// It loads the config, initializes logging, secure URLs, and EPG.
// It then configures the Fiber app with middleware and routes.
// It starts listening on the provided host and port.
// Returns an error if listening fails.
func JioTVServer(host, port, configPath string, prefork bool) error {
	// Load the config file
	if err := config.Cfg.Load(configPath); err != nil {
		return err
	}

	// Initialize the logger object
	utils.Log = utils.GetLogger()

	// Initialize the store object
	if err := store.Init(); err != nil {
		return err
	}

	// Initialize the secureurl object
	secureurl.Init()

	// if config EPG is true or file epg.xml.gz exists
	if config.Cfg.EPG || utils.FileExists("epg.xml.gz") {
		go epg.Init()
	}

	// Start Scheduler
	scheduler.Init()
	defer scheduler.Stop()

	engine := html.NewFileSystem(http.FS(web.GetViewFiles()), ".html")
	if config.Cfg.Debug {
		engine.Reload(true)
	}

	app := fiber.New(fiber.Config{
		Views:             engine,
		Prefork:           prefork,
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

	// Handle all /bpk-tv/* routes
	app.Use("/bpk-tv/", handlers.BpkProxyHandler)

	// Handle all /out/* routes
	app.Use("/out/", handlers.SLHandler)

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
	app.Get("/epg/:channelID/:offset", handlers.WebEPGHandler)
	app.Get("/jtvposter/:date/:file", handlers.PosterHandler)
	app.Get("/mpd/:channelID", handlers.LiveMpdHandler)
	app.Post("/drm", handlers.DRMKeyHandler)
	app.Get("/dashtime", handlers.DASHTimeHandler)

	return app.Listen(host + ":" + port)
}
