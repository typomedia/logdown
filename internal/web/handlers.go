package web

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"logdown/internal/dates"
	"logdown/internal/repo"

	"github.com/gofiber/fiber/v2"
)

// Handler wires HTTP routes to the repository and renderer.
type Handler struct {
	repo     *repo.Repo
	renderer *Renderer
	app      App
	year     int
}

// NewHandler constructs a Handler.
func NewHandler(r *repo.Repo, rd *Renderer, app App, year int) *Handler {
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

func (h *Handler) page(route string) Page {
	return Page{App: h.app, Route: route, Year: h.year}
}

func (h *Handler) render(c *fiber.Ctx, p Page) error {
	html, err := h.renderer.Render(p)
	if err != nil {
		return err
	}
	c.Type("html")
	return c.Send(html)
}

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
		median = append(median, roundStr(log["average"]))
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

// about shows system information. Ports AboutController::info.
func (h *Handler) about(c *fiber.Ctx) error {
	p := h.page("app_about_info")
	p.System = map[string]string{
		"os":    osRelease(),
		"cpu":   cpuModel(),
		"go":    strings.TrimPrefix(runtime.Version(), "go"),
		"fiber": fiber.Version,
	}
	return h.render(c, p)
}

// uploadForm renders the dropzone page. Ports UploadController::index.
func (h *Handler) uploadForm(c *fiber.Ctx) error {
	return h.render(c, h.page("app_upload_index"))
}

// upload accepts the uploaded logfiles and rebuilds the database from scratch.
// Ports UploadController::upload.
func (h *Handler) upload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "logdown-upload-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var paths []string
	for _, headers := range form.File {
		for _, fh := range headers {
			dst := filepath.Join(tmpDir, filepath.Base(fh.Filename))
			if err := c.SaveFile(fh, dst); err != nil {
				return err
			}
			paths = append(paths, dst)
		}
	}

	imported, err := h.repo.Rebuild(paths)
	if err != nil {
		return err
	}
	if imported == nil {
		imported = []string{}
	}
	return c.JSON(fiber.Map{"files": imported})
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

// osRelease returns a human-readable OS description, falling back to GOOS.
func osRelease() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		if v := field(string(data), "PRETTY_NAME="); v != "" {
			return v
		}
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}

// cpuModel returns the CPU model name, falling back to GOARCH.
func cpuModel() string {
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "model name") {
				if i := strings.Index(line, ":"); i >= 0 {
					return strings.TrimSpace(line[i+1:])
				}
			}
		}
	}
	return runtime.GOARCH
}

func field(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key) {
			return strings.Trim(strings.TrimPrefix(line, key), `"`)
		}
	}
	return ""
}
