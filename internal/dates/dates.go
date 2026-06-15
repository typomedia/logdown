// Package dates ports the date helpers from the original Symfony app:
// granularity detection (DateHelper::analyzeDate) and the SQLite strftime
// formats used to group rows by view (AppBundle::VIEWFORMAT).
package dates

import "time"

// View granularities, matching DateHelper::analyzeDate return values.
const (
	Month = "month"
	Day   = "day"
	Hour  = "hour"
	None  = "" // input did not match any recognised granularity
)

// dateFormats maps a granularity to the Go layout its input must match exactly,
// mirroring AppBundle::DATEFORMAT (Y-m, Y-m-d, Y-m-d H).
var dateFormats = []struct {
	view, layout string
}{
	{Month, "2006-01"},
	{Day, "2006-01-02"},
	{Hour, "2006-01-02 15"},
}

// viewFormat ports AppBundle::VIEWFORMAT: the strftime() format bound to :view
// for each (mode, granularity). Mode 0 is the detail/list query, mode 1 the
// single-row summary query. None maps to "" so an unrecognised date yields an
// empty strftime format (matching the original's null binding).
var viewFormat = map[int]map[string]string{
	0: {
		Month: "%Y-%m-%d",
		Day:   "%Y-%m-%d %H:00",
		Hour:  "%Y-%m-%d %H:%M",
		None:  "",
	},
	1: {
		Month: "%Y-%m",
		Day:   "%Y-%m-%d",
		Hour:  "%Y-%m-%d %H",
		None:  "",
	},
}

// Analyze reports the finest granularity that the given date string matches
// exactly, or None if it matches none. Ports DateHelper::analyzeDate.
func Analyze(date string) string {
	for _, f := range dateFormats {
		if t, err := time.Parse(f.layout, date); err == nil && t.Format(f.layout) == date {
			return f.view
		}
	}
	return None
}

// ViewFormat returns the strftime format to bind to :view for the given query
// mode and date string.
func ViewFormat(mode int, date string) string {
	m, ok := viewFormat[mode]
	if !ok {
		m = viewFormat[0]
	}
	return m[Analyze(date)]
}

// flexibleLayouts are tried in order when reformatting a datetime string whose
// exact shape depends on the view that produced it.
var flexibleLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02 15",
	"2006-01-02",
	"2006-01",
}

// Parse interprets a datetime string emitted by strftime, trying the known
// layouts from most to least specific.
func Parse(s string) (time.Time, bool) {
	for _, layout := range flexibleLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Reformat parses s with Parse and reformats it using the given Go layout. If
// the string cannot be parsed it is returned unchanged.
func Reformat(s, layout string) string {
	if t, ok := Parse(s); ok {
		return t.Format(layout)
	}
	return s
}

// ChartLabel formats a datetime string for a chart axis label according to the
// view, mirroring ChartController: day -> hour, hour -> hour:minute,
// month -> day-of-month. With no view the raw string is returned.
func ChartLabel(view, datetime string) string {
	switch view {
	case Day:
		return Reformat(datetime, "15")
	case Hour:
		return Reformat(datetime, "15:04")
	case Month:
		return Reformat(datetime, "02")
	default:
		return datetime
	}
}
