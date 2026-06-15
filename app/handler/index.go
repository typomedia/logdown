package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// index lists aggregated requests, optionally filtered by ?search. Ports
// LogController::index (Request.sql).
func (h *Handler) index(c *fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("search"))
	logs, err := h.repo.Query("Request.sql", 0, map[string]string{"search": search})
	if err != nil {
		return err
	}
	p := h.page("logs_index")
	p.Search = search
	p.Logs = logs
	return h.render(c, p)
}
