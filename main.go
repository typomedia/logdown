package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"logdown/app/handler"
	"logdown/app/renderer"
	"logdown/app/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"

	flag "github.com/spf13/pflag"
)

const (
	appName    = "Logdown"
	appVersion = "2.0.0"
)

//go:embed app/views
var views embed.FS

//go:embed public
var public embed.FS

var engine *html.Engine

func main() {
	addr := flag.String("addr", envOr("LOGDOWN_ADDR", ":4000"), "listen address")
	dbPath := flag.String("db", envOr("LOGDOWN_DB", "sqlog.db"), "path to the SQLite database")
	publicDir := flag.String("public", envOr("LOGDOWN_PUBLIC", "public"), "directory holding the static assets")
	flag.Parse()

	r, err := repo.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer r.Close()

	rdr, err := renderer.NewRenderer()
	if err != nil {
		log.Fatalf("compile templates: %v", err)
	}

	engine = html.NewFileSystem(http.FS(views), ".html")

	app := fiber.New(fiber.Config{
		AppName:               appName,
		Views:                 engine,
		BodyLimit:             512 * 1024 * 1024, // matches the original 512M upload limit
		DisableStartupMessage: false,
		ReadTimeout:           10 * time.Minute,
		WriteTimeout:          10 * time.Minute,
	})
	app.Use(recover.New())
	app.Use(logger.New())

	// Serve the existing assets untouched (themes, fonts, libs, favicon).
	app.Static("/themes", filepath.Join(*publicDir, "themes"))

	h := handler.New(r, rdr, renderer.App{Name: appName, Version: appVersion}, time.Now().Year())
	h.Register(app)

	// publish static embedded certs like css, js, images
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(public),
		PathPrefix: "public",
	}))

	if err := app.Listen(*addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
