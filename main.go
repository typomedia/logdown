package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"logdown/app/handler"
	"logdown/app/renderer"
	"logdown/app/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	flag "github.com/spf13/pflag"
)

const (
	appName    = "Logdown"
	appVersion = "2.0.0"
)

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

	app := fiber.New(fiber.Config{
		AppName:               appName,
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
