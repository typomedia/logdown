package parser

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

const iisFields = "date time s-ip cs-method cs-uri-stem cs-uri-query s-port " +
	"cs-username c-ip cs(User-Agent) cs(Referer) sc-status sc-substatus " +
	"sc-win32-status time-taken"

func tmpLog(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logparser_*.log")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func collect(t *testing.T, p *Parser, path string) []map[string]string {
	t.Helper()
	var rows []map[string]string
	if err := p.ParseFile(path, func(r map[string]string) error {
		rows = append(rows, r)
		return nil
	}); err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	return rows
}

func TestMapsEveryKnownField(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: " + iisFields)
	row := p.ParseLine("2024-02-27 04:00:31 10.0.0.1 GET /a/b key=1 443 alice 10.0.0.2 Mozilla/5.0 http://ref 200 0 0 2171")

	want := map[string]string{
		"date": "2024-02-27 04:00:31", "server": "10.0.0.1", "method": "GET",
		"request": "/a/b", "param": "key=1", "port": "443", "user": "alice",
		"client": "10.0.0.2", "agent": "Mozilla/5.0", "referer": "http://ref",
		"status": "200", "substatus": "0", "win32": "0", "duration": "2171",
	}
	for k, v := range want {
		if row[k] != v {
			t.Errorf("%s: expected %q, got %q", k, v, row[k])
		}
	}
	if len(row) != 14 {
		t.Errorf("expected 14 mapped columns, got %d", len(row))
	}
}

func TestMergesDateAndTime(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: date time sc-status")
	row := p.ParseLine("2024-01-02 03:04:05 200")
	if row["date"] != "2024-01-02 03:04:05" {
		t.Errorf("got %q", row["date"])
	}
	if _, ok := row["time"]; ok {
		t.Error("no separate time column expected")
	}
}

func TestDynamicFieldOrder(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: time-taken cs-method date time sc-status cs-uri-stem")
	row := p.ParseLine("123 GET 2024-01-02 03:04:05 404 /foo")
	for k, v := range map[string]string{
		"duration": "123", "method": "GET", "date": "2024-01-02 03:04:05",
		"status": "404", "request": "/foo",
	} {
		if row[k] != v {
			t.Errorf("%s: expected %q, got %q", k, v, row[k])
		}
	}
}

func TestDropsUnmappedFields(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: date time cs-method sc-bytes cs-bytes time-taken")
	row := p.ParseLine("2026-05-16 00:00:01 POST 6671 6021 489")
	if row["duration"] != "489" {
		t.Errorf("duration should be time-taken, got %q", row["duration"])
	}
	if _, ok := row["sc-bytes"]; ok {
		t.Error("sc-bytes should be dropped")
	}
	if len(row) != 3 {
		t.Errorf("only date, method, duration expected; got %d", len(row))
	}
}

func TestParsesSubsetOfFields(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: date time c-ip sc-status")
	row := p.ParseLine("2024-01-01 00:00:00 1.2.3.4 500")
	if row["date"] != "2024-01-01 00:00:00" || row["client"] != "1.2.3.4" || row["status"] != "500" {
		t.Errorf("unexpected row %v", row)
	}
	if len(row) != 3 {
		t.Errorf("expected 3 columns, got %d", len(row))
	}
}

func TestMissingTrailingValuesDefaultToDash(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: date time cs-method sc-status time-taken")
	row := p.ParseLine("2024-01-01 00:00:00 GET")
	if row["method"] != "GET" || row["status"] != "-" || row["duration"] != "-" {
		t.Errorf("unexpected row %v", row)
	}
}

func TestCommentLinesReturnNil(t *testing.T) {
	p := New()
	for _, l := range []string{
		"#Software: Microsoft Internet Information Services 10.0",
		"#Version: 1.0",
		"#Date: 2024-02-27 04:00:31",
	} {
		if p.ParseLine(l) != nil {
			t.Errorf("expected nil for %q", l)
		}
	}
}

func TestBlankLinesReturnNil(t *testing.T) {
	p := New()
	for _, l := range []string{"", "   ", "\n"} {
		if p.ParseLine(l) != nil {
			t.Errorf("expected nil for %q", l)
		}
	}
}

func TestDataBeforeFieldsReturnsNil(t *testing.T) {
	p := New()
	if p.ParseLine("2024-01-01 00:00:00 1.2.3.4 GET") != nil {
		t.Error("expected nil before #Fields")
	}
	if len(p.Fields()) != 0 {
		t.Error("expected no fields")
	}
}

func TestLaterFieldsRedefineColumns(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: date time cs-method")
	first := p.ParseLine("2024-01-01 00:00:00 GET")
	if first["method"] != "GET" {
		t.Errorf("got %q", first["method"])
	}

	p.ParseLine("#Fields: date time sc-status")
	second := p.ParseLine("2024-01-01 00:00:01 404")
	if second["status"] != "404" {
		t.Errorf("got %q", second["status"])
	}
	if _, ok := second["method"]; ok {
		t.Error("old field should no longer be mapped")
	}
}

func TestGetFieldsReflectsDirective(t *testing.T) {
	p := New()
	p.ParseLine("#Fields: date time sc-status")
	got := p.Fields()
	want := []string{"date", "time", "sc-status"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v", got)
		}
	}
}

func TestFieldDirectiveToleratesWhitespace(t *testing.T) {
	p := New()
	p.ParseLine("#Fields:   date    time   sc-status  ")
	row := p.ParseLine("2024-01-01 00:00:00 200")
	if row["status"] != "200" {
		t.Errorf("got %q", row["status"])
	}
}

func TestParseFileStreamsOnlyDataRows(t *testing.T) {
	path := tmpLog(t, "#Software: Microsoft Internet Information Services 10.0\n"+
		"#Version: 1.0\n#Date: 2024-02-27 04:00:31\n#Fields: "+iisFields+"\n"+
		"2024-02-27 04:00:31 10.0.0.1 GET /a - 443 - 10.0.0.2 Agent - 200 0 0 10\n"+
		"2024-02-27 04:00:32 10.0.0.1 GET /b - 443 - 10.0.0.2 Agent - 200 0 0 20\n")
	rows := collect(t, New(), path)
	if len(rows) != 2 {
		t.Fatalf("expected 2 data rows, got %d", len(rows))
	}
	if rows[0]["request"] != "/a" || rows[1]["duration"] != "20" {
		t.Errorf("unexpected rows %v", rows)
	}
}

func TestParseFileMultipleSegments(t *testing.T) {
	seg := func(stem string) string {
		return "#Software: Microsoft Internet Information Services 10.0\n" +
			"#Fields: date time cs-method cs-uri-stem time-taken\n" +
			"2024-02-27 04:00:31 GET " + stem + " 99"
	}
	path := tmpLog(t, seg("/one")+"\n"+seg("/two")+"\n")
	rows := collect(t, New(), path)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["request"] != "/one" || rows[1]["request"] != "/two" || rows[1]["duration"] != "99" {
		t.Errorf("unexpected rows %v", rows)
	}
}

func TestParseFileMissingFile(t *testing.T) {
	err := New().ParseFile("/no/such/file.log", func(map[string]string) error { return nil })
	if err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestParsesCombinedLogFormat(t *testing.T) {
	p := New()
	row := p.ParseLine(`127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif?x=1 HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`)
	want := map[string]string{
		"client": "127.0.0.1", "user": "frank",
		"date":   "2000-10-10 13:55:36",
		"method": "GET", "request": "/apache_pb.gif", "param": "x=1",
		"status":  "200",
		"referer": "http://www.example.com/start.html",
		"agent":   "Mozilla/4.08 [en] (Win98; I ;Nav)",
	}
	for k, v := range want {
		if row[k] != v {
			t.Errorf("%s: expected %q, got %q", k, v, row[k])
		}
	}
}

func TestParsesCommonLogFormat(t *testing.T) {
	p := New()
	row := p.ParseLine(`10.0.0.1 - - [10/Oct/2000:13:55:36 +0000] "POST /api/v1/users HTTP/1.1" 201 512`)
	if row["client"] != "10.0.0.1" || row["method"] != "POST" ||
		row["request"] != "/api/v1/users" || row["param"] != "-" ||
		row["status"] != "201" || row["date"] != "2000-10-10 13:55:36" {
		t.Errorf("unexpected row %v", row)
	}
	if _, ok := row["referer"]; ok {
		t.Error("plain CLF should not produce a referer column")
	}
	if _, ok := row["agent"]; ok {
		t.Error("plain CLF should not produce an agent column")
	}
}

func TestParseFileAutoDetectsAccessLog(t *testing.T) {
	path := tmpLog(t, `198.51.100.21 - alice [13/Jun/2026:00:01:34 +0000] "GET /api/v1/orders?status=open HTTP/1.1" 200 1742 "-" "curl/8.4.0"
198.51.100.21 - - [13/Jun/2026:00:02:48 +0000] "GET /missing.html HTTP/1.1" 404 21 "-" "Mozilla/5.0"
`)
	rows := collect(t, New(), path)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["request"] != "/api/v1/orders" || rows[0]["param"] != "status=open" {
		t.Errorf("row 0 unexpected %v", rows[0])
	}
	if rows[1]["status"] != "404" || rows[1]["agent"] != "Mozilla/5.0" {
		t.Errorf("row 1 unexpected %v", rows[1])
	}
}

func TestParsesCombinedExtendedWithPortAndDuration(t *testing.T) {
	p := New()
	row := p.ParseLine(`198.51.100.21 - alice [13/Jun/2026:00:01:34 +0000] "GET /api/v1/orders?status=open HTTP/1.1" 200 4912 "-" "curl/8.4.0" 443 274`)
	if row["port"] != "443" || row["duration"] != "274" {
		t.Errorf("expected port=443 duration=274, got %v", row)
	}
}

func TestAccessLogLockoutFromIISMixed(t *testing.T) {
	p := New()
	// Once an IIS directive is seen the access-log path is no longer tried,
	// so a CLF-looking line afterwards is treated as malformed IIS data and
	// returns nil instead of getting misclassified.
	p.ParseLine("#Fields: date time cs-method")
	if row := p.ParseLine(`127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET / HTTP/1.0" 200 1`); row != nil {
		if _, ok := row["client"]; ok {
			t.Errorf("CLF line should not be parsed once the parser is pinned to IIS, got %v", row)
		}
	}
}

func TestAllRepoLogfilesParse(t *testing.T) {
	root := filepath.Join("..", "..")
	var logs []string
	for _, pat := range []string{"u_ex*.log", "public/u_ex*.log"} {
		m, _ := filepath.Glob(filepath.Join(root, pat))
		logs = append(logs, m...)
	}
	if len(logs) == 0 {
		t.Skip("no repo logfiles found to integration-test")
	}

	valid := make(map[string]bool)
	for _, c := range Columns() {
		valid[c] = true
	}

	for _, path := range logs {
		name := filepath.Base(path)
		var count int
		err := New().ParseFile(path, func(row map[string]string) error {
			count++
			cols := make([]string, 0, len(row))
			for c := range row {
				cols = append(cols, c)
			}
			sort.Strings(cols)
			for _, c := range cols {
				if !valid[c] {
					t.Errorf("%s: unknown column %q", name, c)
				}
			}
			if row["date"] == "" {
				t.Errorf("%s: missing date", name)
			}
			return nil
		})
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if count == 0 {
			t.Errorf("%s: parsed no rows", name)
		}
	}
}
