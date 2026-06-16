// Package parser reads the W3C Extended Log File Format used by IIS.
package parser

import (
	"bufio"
	"os"
	"strings"
)

// FieldMap maps W3C extended log field identifiers to Log table columns.
// Fields not listed here (e.g. sc-bytes, cs-bytes) are ignored. The "date"
// and "time" fields are handled separately and merged into the "date" column.
var FieldMap = map[string]string{
	"s-ip":            "server",
	"cs-method":       "method",
	"cs-uri-stem":     "request",
	"cs-uri-query":    "param",
	"s-port":          "port",
	"cs-username":     "user",
	"c-ip":            "client",
	"cs(User-Agent)":  "agent",
	"cs(Referer)":     "referer",
	"sc-status":       "status",
	"sc-substatus":    "substatus",
	"sc-win32-status": "win32",
	"time-taken":      "duration",
}

// Columns returns every Log column the parser may emit (mapped columns + date).
func Columns() []string {
	cols := make([]string, 0, len(FieldMap)+1)
	for _, c := range FieldMap {
		cols = append(cols, c)
	}
	return append(cols, "date")
}

// format is the access log dialect a Parser has locked onto. A parser starts
// in formatUnknown and is pinned the moment it sees its first IIS directive or
// a line that matches the CLF/Combined pattern, after which subsequent lines
// are dispatched accordingly.
type format int

const (
	formatUnknown format = iota
	formatIIS
	formatAccess
)

// Parser holds the field order currently in effect, taken from the most recent
// "#Fields:" directive. A single Parser may be reused across many lines and
// files; each "#Fields:" directive redefines the columns.
type Parser struct {
	fields []string
	format format
}

// New returns a Parser with no fields defined yet.
func New() *Parser {
	return &Parser{}
}

// Fields returns the field identifiers currently in effect, in column order.
func (p *Parser) Fields() []string {
	return p.fields
}

// SetFields defines the column order from the content of a "#Fields:" directive.
func (p *Parser) SetFields(directive string) {
	p.fields = strings.Fields(strings.TrimSpace(directive))
}

// ParseLine parses a single line.
//
// Directive lines (starting with "#") pin the parser to the IIS dialect and
// return nil; "#Fields:" defines the column order used for following IIS data
// lines. Non-directive lines are dispatched to the IIS or CLF/Combined parser
// depending on which dialect the parser has locked onto; a parser that has
// not yet seen a directive auto-detects CLF/Combined on the first match.
// Returns nil when the line cannot be interpreted.
func (p *Parser) ParseLine(line string) map[string]string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if line[0] == '#' {
		p.format = formatIIS
		if strings.HasPrefix(strings.ToLower(line), "#fields:") {
			p.SetFields(line[len("#Fields:"):])
		}
		return nil
	}

	switch p.format {
	case formatIIS:
		if len(p.fields) == 0 {
			return nil
		}
		return p.mapValues(strings.Fields(line))
	case formatAccess:
		return parseAccess(line)
	}

	if row := parseAccess(line); row != nil {
		p.format = formatAccess
		return row
	}
	return nil
}

// ParseFile parses a whole logfile, invoking fn for every data row. Iteration
// stops and the error is returned if fn returns an error or the file cannot be
// read.
func (p *Parser) ParseFile(path string, fn func(map[string]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// IIS request/referer/agent fields can make for very long lines.
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		if row := p.ParseLine(sc.Text()); row != nil {
			if err := fn(row); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

// mapValues maps a row of raw values to a row keyed by Log column name. Missing
// trailing values default to "-"; the "date" and "time" fields are merged into
// a single "date" column.
func (p *Parser) mapValues(values []string) map[string]string {
	row := make(map[string]string)
	var date, dtime string
	var haveDate, haveTime bool

	for i, field := range p.fields {
		value := "-"
		if i < len(values) {
			value = values[i]
		}

		switch field {
		case "date":
			date, haveDate = value, true
			continue
		case "time":
			dtime, haveTime = value, true
			continue
		}

		if col, ok := FieldMap[field]; ok {
			row[col] = value
		}
	}

	if haveDate || haveTime {
		row["date"] = strings.TrimSpace(date + " " + dtime)
	}

	return row
}
