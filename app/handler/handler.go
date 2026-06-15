// Package handler implements the HTTP route handlers, one file per route.
// It owns the request/response wiring and delegates rendering to app/renderer.
package handler

import (
	"logdown/app/renderer"
	"logdown/app/repo"

	"github.com/gofiber/fiber/v2"
)

// Handler wires HTTP routes to the repository and renderer.
type Handler struct {
	repo     *repo.Repo
	renderer *renderer.Renderer
	app      renderer.App
	year     int
}

// New constructs a Handler.
func New(r *repo.Repo, rd *renderer.Renderer, app renderer.App, year int) *Handler {
	return &Handler{repo: r, renderer: rd, app: app, year: year}
}

// Register mounts every route on the Fiber app, mirroring the original
// controller routes.
func (h *Handler) Register(app *fiber.App) {
	app.Get("/", h.index)
	app.Get("/search", h.search)
	app.Get("/chart", h.chart)
	app.Get("/about", h.about)
	app.Get("/about/", h.about)
	app.Get("/upload", h.uploadForm)
	app.Get("/upload/", h.uploadForm)
	app.Post("/upload/upload", h.upload)
}

func (h *Handler) page(route string) renderer.Page {
	return renderer.Page{App: h.app, Route: route, Year: h.year}
}

func (h *Handler) render(c *fiber.Ctx, p renderer.Page) error {
	html, err := h.renderer.Render(p)
	if err != nil {
		return err
	}
	c.Type("html")
	return c.Send(html)
}
