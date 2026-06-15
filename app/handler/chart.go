package handler

import (
	"encoding/json"
	"html/template"
	"strings"

	"logdown/app/dates"
	"logdown/app/renderer"
	"logdown/app/repo"

	"github.com/gofiber/fiber/v2"
)

// chart renders the request timeline. Ports ChartController::chart (Chart.sql).
func (h *Handler) chart(c *fiber.Ctx) error {
	params := map[string]string{
		"request": c.Query("request"),
		"param":   c.Query("param"),
		"status":  c.Query("status"),
		"date":    c.Query("date"),
	}
	view := dates.Analyze(c.Query("date"))

	logs, err := h.repo.Query("Chart.sql", 0, params)
	if err != nil {
		return err
	}

	labels := make([]string, 0, len(logs))
	number := make([]string, 0, len(logs))
	median := make([]string, 0, len(logs))
	for _, log := range logs {
		labels = append(labels, dates.ChartLabel(view, log["datetime"]))
		number = append(number, log["number"])
		median = append(median, renderer.Round(log["average"]))
	}

	info, err := h.repo.QueryOne("Chart.sql", 1, params)
	if err != nil {
		return err
	}
	// The template reads info fields unconditionally; when no row matches the
	// requested params, seed empty defaults so missing map keys don't surface
	// as invalid template values (the original Twig tolerated a missing row).
	if info == nil {
		info = repo.Row{}
	}
	for _, k := range []string{"datetime", "method", "request", "param", "port", "status", "number", "average"} {
		if _, ok := info[k]; !ok {
			info[k] = ""
		}
	}

	p := h.page("app_chart_chart")
	p.Info = info
	p.View = view
	p.Labels = jsonStrings(labels)
	p.Number = jsonNumbers(number)
	p.Median = jsonNumbers(median)
	return h.render(c, p)
}

// jsonStrings marshals labels as a JSON array of strings for the chart.
func jsonStrings(values []string) template.JS {
	b, err := json.Marshal(values)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}

// jsonNumbers builds a JSON array from numeric strings, leaving them unquoted
// so Chartist receives real numbers.
func jsonNumbers(values []string) template.JS {
	if len(values) == 0 {
		return template.JS("[]")
	}
	return template.JS("[" + strings.Join(values, ",") + "]")
}
