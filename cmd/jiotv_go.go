package cmd

import (
	"fmt"
	"log" // Added import for *log.Logger type
	"net/http"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants"
	"github.com/jiotv-go/jiotv_go/v3/internal/handlers"
	"github.com/jiotv-go/jiotv_go/v3/internal/middleware"
	"github.com/jiotv-go/jiotv_go/v3/pkg/epg"
	"github.com/jiotv-go/jiotv_go/v3/pkg/scheduler"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/jiotv-go/jiotv_go/v3/web"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

// LoadConfig loads the application configuration from the given path.
func LoadConfig(configPath string) error {
	return config.Cfg.Load(configPath)
}

// InitializeLogger initializes the global logger.
// This should be called after LoadConfig.
func InitializeLogger() {
	utils.Log = utils.GetLogger()
}

// Logger returns the initialized global logger.
// Ensure InitializeLogger has been called before using this.
func Logger() *log.Logger { // Corrected to *log.Logger
	return utils.Log
}

type JioTVServerConfig struct {
	Host        string
	Port        string
	TLS         bool
	TLSCertPath string
	TLSKeyPath  string
}

// JioTVServer starts the JioTV server.
// Assumes config and logger are already initialized.
// It initializes secure URLs, EPG, store, and handlers.
// It then configures the Fiber app with middleware and routes.
// It starts listening on the provided host and port.
// Returns an error if listening fails.
func JioTVServer(jiotvServerConfig JioTVServerConfig) error {
	// Config, Logger and Store are assumed to be initialized in main.go

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
		Network:           fiber.NetworkTCP,
		StreamRequestBody: true,
		CaseSensitive:     false,
		StrictRouting:     false,
		EnablePrintRoutes: false,
		ServerHeader:      "JioTV Go",
		AppName:           fmt.Sprintf("JioTV Go %s", constants.Version),
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
	app.Get("/favicon.ico", handlers.FaviconHandler)
	app.Get("/jtvimage/:file", handlers.ImageHandler)
	app.Get("/epg.xml.gz", handlers.EPGHandler)
	app.Get("/epg/:channelID/:offset", handlers.WebEPGHandler)
	app.Get("/jtvposter/:date/:file", handlers.PosterHandler)
	app.Get("/mpd/:channelID", handlers.LiveMpdHandler)
	app.Post("/drm", handlers.DRMKeyHandler)
	app.Get("/dashtime", handlers.DASHTimeHandler)

	app.Get("/render.mpd", handlers.MpdHandler)
	app.Use("/render.dash", handlers.DashHandler)

	if jiotvServerConfig.TLS {
		if jiotvServerConfig.TLSCertPath == "" || jiotvServerConfig.TLSKeyPath == "" {
			return fmt.Errorf("TLS cert and key paths are required for HTTPS. Please provide them using --tls-cert and --tls-key flags")
		}
		return app.ListenTLS(fmt.Sprintf("%s:%s", jiotvServerConfig.Host, jiotvServerConfig.Port), jiotvServerConfig.TLSCertPath, jiotvServerConfig.TLSKeyPath)
	} else {
		return app.Listen(fmt.Sprintf("%s:%s", jiotvServerConfig.Host, jiotvServerConfig.Port))
	}
}
