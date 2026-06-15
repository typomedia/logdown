// Command logdown is a Go/Fiber rewrite of the Logdown IIS log analyzer.
package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"logdown/internal/repo"
	"logdown/internal/web"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	flag "github.com/spf13/pflag"
)

const (
	appName    = "Logdown"
	appVersion = "1.1.0"
)

func main() {
	addr := flag.String("addr", envOr("LOGDOWN_ADDR", ":4000"), "listen address")
	dbPath := flag.String("db", envOr("LOGDOWN_DB", "var/data/sqlog.db"), "path to the SQLite database")
	webDir := flag.String("web", envOr("LOGDOWN_WEB", "web"), "directory holding the static assets")
	cache := flag.Bool("cache", envOr("LOGDOWN_CACHE", "true") == "true", "enable the query result cache")
	flag.Parse()

	r, err := repo.Open(*dbPath, *cache)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer r.Close()

	renderer, err := web.NewRenderer()
	if err != nil {
		log.Fatalf("compile templates: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               appName,
		BodyLimit:             512 * 1024 * 1024, // matches the original 512M upload limit
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Minute,
		WriteTimeout:          10 * time.Minute,
	})
	app.Use(recover.New())
	app.Use(logger.New())

	// Serve the existing assets untouched (themes, fonts, libs, favicon).
	app.Static("/themes", filepath.Join(*webDir, "themes"))
	app.Static("/favicon.ico", filepath.Join(*webDir, "favicon.ico"))

	handler := web.NewHandler(r, renderer, web.App{Name: appName, Version: appVersion}, time.Now().Year())
	handler.Register(app)

	log.Printf("%s listening on %s (db=%s, web=%s)", appName, *addr, *dbPath, *webDir)
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
