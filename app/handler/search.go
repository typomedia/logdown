package handler

import (
	"logdown/app/dates"

	"github.com/gofiber/fiber/v2"
)

// search lists log rows for a request/param/date. Ports LogController::search
// (Search.sql).
func (h *Handler) search(c *fiber.Ctx) error {
	params := map[string]string{
		"request": c.Query("request"),
		"param":   c.Query("param"),
		"date":    c.Query("date"),
	}
	logs, err := h.repo.Query("Search.sql", 0, params)
	if err != nil {
		return err
	}
	p := h.page("logs_search")
	p.Logs = logs
	p.View = dates.Analyze(c.Query("date"))
	return h.render(c, p)
}
