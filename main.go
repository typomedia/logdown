package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
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
	Name    = "Logdown"
	Version = "2.0.1"
)

//go:embed app/views
var views embed.FS

//go:embed public
var public embed.FS

var engine *html.Engine

func main() {
	port := flag.IntP("port", "p", 4000, "Port to listen on")
	database := flag.StringP("database", "d", "logdown.db", "SQLite database file")
	flag.Parse()

	r, err := repo.Open(*database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer r.Close()

	rdr, err := renderer.NewRenderer()
	if err != nil {
		log.Fatalf("compile templates: %v", err)
	}

	engine = html.NewFileSystem(http.FS(views), ".html")
	// The renderer package owns the canonical template func map (round,
	// inc, fmtDate, chartURL, searchURL, hasPrefix). Register it here so
	// Fiber's view engine can parse the same templates without warnings.
	engine.AddFuncMap(renderer.Funcs)

	app := fiber.New(fiber.Config{
		AppName:               Name,
		Views:                 engine,
		BodyLimit:             512 * 1024 * 1024, // matches the original 512M upload limit
		DisableStartupMessage: false,
		ReadTimeout:           10 * time.Minute,
		WriteTimeout:          10 * time.Minute,
	})
	app.Use(recover.New())
	app.Use(logger.New())

	h := handler.New(r, rdr, renderer.App{Name: Name, Version: Version}, time.Now().Year())
	h.Register(app)

	// publish static embedded certs like css, js, images
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(public),
		PathPrefix: "public",
	}))

	log.Fatal(app.Listen(fmt.Sprintf(":%d", *port)))
}
