// Package repo ports the original LogRepository: it runs the named SQL queries
// against the SQLite database, binding request parameters exactly as the
// Symfony app did, and rebuilds the database from uploaded IIS logfiles.
package repo

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"logdown/internal/dates"
	"logdown/internal/parser"

	_ "modernc.org/sqlite"
)

//go:embed queries/*.sql
var queryFS embed.FS

// Row is a single result row keyed by column name. As with PDO's associative
// fetch against SQLite, every value is represented as a string.
type Row map[string]string

var varPattern = regexp.MustCompile(`:(\w+)`)

// Repo holds the connection to the on-disk SQLite database and a small result
// cache. The database is replaced wholesale on upload, so reads and rebuilds
// are guarded by a RWMutex.
type Repo struct {
	path  string
	cache bool

	mu   sync.RWMutex
	db   *sql.DB
	memo map[string]cacheEntry
}

type cacheEntry struct {
	rows []Row
}

// Open opens (or lazily creates) the SQLite database at path. cache enables the
// in-memory result cache, mirroring the original "cache" parameter.
func Open(path string, cache bool) (*Repo, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := openDB(path)
	if err != nil {
		return nil, err
	}
	return &Repo{path: path, cache: cache, db: db, memo: map[string]cacheEntry{}}, nil
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// Close releases the database connection.
func (r *Repo) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.db.Close()
}

// Query runs the named query (e.g. "Request.sql") in the given mode, binding
// values drawn from params. Mode 0 selects the detail strftime formats, mode 1
// the summary formats. Ports LogRepository::get with FETCH_ALL semantics.
func (r *Repo) Query(name string, mode int, params map[string]string) ([]Row, error) {
	text, err := queryFS.ReadFile("queries/" + name)
	if err != nil {
		return nil, err
	}
	query := string(text)
	args := bindArgs(query, mode, params)

	key := cacheKey(name, mode, args)
	if r.cache {
		r.mu.RLock()
		entry, ok := r.memo[key]
		r.mu.RUnlock()
		if ok {
			return entry.rows, nil
		}
	}

	r.mu.RLock()
	rows, err := runQuery(r.db, query, args)
	r.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	if r.cache {
		r.mu.Lock()
		r.memo[key] = cacheEntry{rows: rows}
		r.mu.Unlock()
	}
	return rows, nil
}

// QueryOne runs the named query and returns the first row, or nil if there are
// none. Ports LogRepository::get with FETCH_ONE semantics.
func (r *Repo) QueryOne(name string, mode int, params map[string]string) (Row, error) {
	rows, err := r.Query(name, mode, params)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// bindArgs builds the named query arguments from the query text and request
// parameters, reproducing LogRepository::get's per-key binding rules:
//   - only variables that appear in the SQL are considered;
//   - "date"   binds value+"%" and derives :view from the date's granularity;
//   - "search" binds "%"+value+"%";
//   - everything else binds the trimmed value verbatim.
func bindArgs(query string, mode int, params map[string]string) []any {
	var args []any
	seen := map[string]bool{}

	for _, m := range varPattern.FindAllStringSubmatch(query, -1) {
		name := m[1]
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true

		value := strings.TrimSpace(params[name])

		switch key {
		case "view":
			// Bound as part of the "date" case below.
			continue
		case "date":
			args = append(args, sql.Named("date", value+"%"))
			args = append(args, sql.Named("view", dates.ViewFormat(mode, value)))
		case "search":
			args = append(args, sql.Named("search", "%"+value+"%"))
		default:
			args = append(args, sql.Named(key, value))
		}
	}
	return args
}

func runQuery(db *sql.DB, query string, args []any) ([]Row, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var out []Row
	for rows.Next() {
		cells := make([]any, len(cols))
		for i := range cells {
			cells[i] = new(any)
		}
		if err := rows.Scan(cells...); err != nil {
			return nil, err
		}
		row := make(Row, len(cols))
		for i, c := range cols {
			row[c] = toString(*(cells[i].(*any)))
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func cacheKey(name string, mode int, args []any) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s|%d", name, mode)
	for _, a := range args {
		if n, ok := a.(sql.NamedArg); ok {
			fmt.Fprintf(&b, "|%s=%v", n.Name, n.Value)
		}
	}
	return b.String()
}

// Rebuild creates a fresh SQLite database from the given logfile paths and
// atomically swaps it in for the live database, mirroring UploadController:
// the database is built off to the side, then moved into place in one pass.
// It returns the names of the files that were imported.
func (r *Repo) Rebuild(paths []string) ([]string, error) {
	schema, err := queryFS.ReadFile("queries/Create.sql")
	if err != nil {
		return nil, err
	}

	tmp := r.path + ".build"
	_ = os.Remove(tmp)

	imported, err := buildDatabase(tmp, string(schema), paths)
	if err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.db.Close(); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, r.path); err != nil {
		// Best effort to restore a usable connection.
		r.db, _ = openDB(r.path)
		return nil, err
	}
	db, err := openDB(r.path)
	if err != nil {
		return nil, err
	}
	r.db = db
	r.memo = map[string]cacheEntry{}
	return imported, nil
}

func buildDatabase(path, schema string, paths []string) ([]string, error) {
	db, err := openDB(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	for _, pragma := range []string{schema, "PRAGMA foreign_keys = OFF", "PRAGMA journal_mode = OFF", "PRAGMA synchronous = OFF"} {
		if _, err := db.Exec(pragma); err != nil {
			return nil, err
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	stmts := map[string]*sql.Stmt{}
	defer func() {
		for _, s := range stmts {
			s.Close()
		}
	}()

	p := parser.New()
	var imported []string
	for _, path := range paths {
		insertErr := p.ParseFile(path, func(row map[string]string) error {
			return insertRow(tx, stmts, row)
		})
		if insertErr != nil {
			_ = tx.Rollback()
			return nil, insertErr
		}
		imported = append(imported, filepath.Base(path))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return imported, nil
}

// insertRow inserts a parsed row, building the statement from the columns the
// parser actually produced so that varying field sets are supported. Prepared
// statements are cached per distinct column set. Ports UploadController::insertRow.
func insertRow(tx *sql.Tx, stmts map[string]*sql.Stmt, row map[string]string) error {
	cols := make([]string, 0, len(row))
	for c := range row {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	key := strings.Join(cols, ",")

	stmt, ok := stmts[key]
	if !ok {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",")
		query := fmt.Sprintf("INSERT INTO Log (%s) VALUES (%s)", strings.Join(cols, ", "), placeholders)
		var err error
		stmt, err = tx.Prepare(query)
		if err != nil {
			return err
		}
		stmts[key] = stmt
	}

	values := make([]any, len(cols))
	for i, c := range cols {
		values[i] = row[c]
	}
	_, err := stmt.Exec(values...)
	return err
}
