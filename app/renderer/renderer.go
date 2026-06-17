package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"

	"logdown/app/dates"
	"logdown/app/repo"
	"logdown/app/views"
)

// App holds the application metadata shown in the layout (name/version).
type App struct {
	Name    string
	Version string
}

// Page is the data passed to every template. Page-specific fields are left at
// their zero value when unused.
type Page struct {
	App    App
	Route  string // route name, used for the body class and nav highlighting
	Year   int
	Search string

	Logs []repo.Row
	View string
	Info repo.Row

	Labels template.JS
	Number template.JS
	Median template.JS

	System map[string]string
}

// Renderer compiles the layout together with each page template and renders
// them to a buffer.
type Renderer struct {
	pages map[string]*template.Template
}

// pages maps a route name to its template file. Each is parsed together with
// base.html so the page's {{define}} blocks override the layout's defaults.
var pageFiles = map[string]string{
	"logs_index":       "index.html",
	"logs_search":      "search.html",
	"app_chart_chart":  "chart.html",
	"app_about_info":   "about.html",
	"app_upload_index": "upload.html",
}

// Funcs is the template function map used by every page template. It is
// exported so callers that compile templates outside this package (e.g. the
// Fiber view engine in main.go) can register the same helpers.
var Funcs = template.FuncMap{
	"inc":       func(i int) int { return i + 1 },
	"round":     Round,
	"hasPrefix": strings.HasPrefix,
	"fmtDate":   dates.Reformat,
	"chartURL":  chartURL,
	"searchURL": searchURL,
}

// NewRenderer compiles all page templates.
func NewRenderer() (*Renderer, error) {
	r := &Renderer{pages: map[string]*template.Template{}}
	for route, file := range pageFiles {
		t, err := template.New("base.html").Funcs(Funcs).ParseFS(views.FS,
			"base.html", file)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}
		r.pages[route] = t
	}
	return r, nil
}

// Render executes the template for the page's route and returns the HTML.
func (r *Renderer) Render(p Page) ([]byte, error) {
	t, ok := r.pages[p.Route]
	if !ok {
		return nil, fmt.Errorf("no template for route %q", p.Route)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base.html", p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Round rounds a numeric string to the nearest integer, mirroring Twig's
// |round filter as used for durations. It is the template "round" function and
// is also reused by the chart handler for the median series.
func Round(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return strconv.FormatFloat(math.Round(f), 'f', 0, 64)
}

// chartURL builds the /chart link for a row, formatting the date with the given
// Go layout. Ports the path('app_chart_chart', {...}) calls in the templates.
func chartURL(row repo.Row, dateLayout string) template.URL {
	q := url.Values{}
	q.Set("request", row["request"])
	q.Set("param", row["param"])
	q.Set("status", row["status"])
	q.Set("date", dates.Reformat(row["datetime"], dateLayout))
	return template.URL("/chart?" + q.Encode())
}

// searchURL builds the /search link for a row. Ports path('logs_search', {...}).
func searchURL(row repo.Row, dateLayout string) template.URL {
	q := url.Values{}
	q.Set("request", row["request"])
	q.Set("param", row["param"])
	q.Set("date", dates.Reformat(row["datetime"], dateLayout))
	return template.URL("/search?" + q.Encode())
}
