# Logdown

Logdown is a log analyzer for IIS logs, rewritten in Go with
[Fiber](https://gofiber.io/). It parses W3C Extended Log Format files, stores
them in SQLite and serves the request/search/chart views. The front-end assets
(Bootstrap, jQuery, Chartist, Dropzone, fonts) are kept exactly as they were.

## Requirements

* Go 1.21+ (built and tested with Go 1.25)

The SQLite driver is the pure-Go [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite),
so no CGO or system libraries are needed.

## Run

    make run        # go run .
    make build      # builds ./logdown
    make test       # runs the unit tests

Then open <http://localhost:4000>.

## Configuration

All options have sensible defaults and can be set via flag or environment variable:

| Flag      | Env             | Default             | Description                       |
|-----------|-----------------|---------------------|-----------------------------------|
| `-addr`   | `LOGDOWN_ADDR`  | `:4000`             | Listen address                    |
| `-db`     | `LOGDOWN_DB`    | `sqlog.db`          | SQLite database path              |
| `-public` | `LOGDOWN_PUBLIC`| `public`            | Static asset directory            |

The upload body limit is 512 MB and read/write timeouts are 10 minutes, matching
the limits the original PHP deployment used.

The queries run uncached: the monthly bucket uses `substr(date,1,7)` instead of
`strftime`, and expression indexes (`ix_month`, `ix_filter`, created after each
upload) let SQLite serve the request/search/chart views directly, so no result
cache is needed.

## Usage

* **Requests** (`/`) — aggregated requests per month; filter with the search box.
* **Search** (`/search`) — drill into a single request/param over a date range.
* **Chart** (`/chart`) — request timeline rendered with Chartist.
* **Upload** (`/upload/`) — drag & drop IIS logfiles; the database is rebuilt
  from scratch on every upload (up to 40 parallel uploads).

## Layout

    main.go                 entry point, Fiber wiring, static assets
    app/parser              W3C Extended Log Format parser (+ tests)
    app/dates               granularity detection and strftime view formats
    app/repo                SQL queries, named-parameter binding, DB rebuild
    app/handler             HTTP route handlers (one file per route)
    app/renderer            html/template rendering and view helpers
    app/views               embedded HTML templates
    public/                 front-end assets (unchanged)
