// Package calendar renders Monday-first month, multi-month and year grids in
// the style of GNU cal, with Lithuanian holiday marking and a today highlight.
//
// All rendering functions take an explicit reference "today" date and a color
// flag through Options, so output is fully deterministic and testable.
package calendar

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mantas6/mcal/internal/holidays"
)

// Layout constants matching GNU cal conventions.
const (
	cellWidth = 2  // width of a single day cell ("%2d")
	gridWidth = 20 // 7 cells + 6 separating spaces
	weekRows  = 6  // GNU cal always emits 6 week rows per month

	multiGap = 2 // column gap for -3 / -n multi-month views
	yearGap  = 3 // column gap for the -y year view
	yearCols = 3 // columns in the year view
)

// ANSI escape sequences.
const (
	reverseOn  = "\x1b[7m"
	reverseOff = "\x1b[27m"
	redOn      = "\x1b[31m"
	redOff     = "\x1b[39m"
	// grayOn/grayOff dim passed (past) day numbers. Bright-black (90) is used
	// rather than faint (2) for wider terminal support and because faint shares
	// its reset (22m) with bold, which would clash when both apply.
	grayOn  = "\x1b[90m"
	grayOff = "\x1b[39m"
	// boldOn/boldOff embolden weekend day numbers.
	boldOn  = "\x1b[1m"
	boldOff = "\x1b[22m"
)

// dayHeader is the Monday-first English weekday header (exactly gridWidth wide).
const dayHeader = "Mo Tu We Th Fr Sa Su"

// HolidayStyle controls how holidays are marked in rendered output.
type HolidayStyle int

const (
	// StyleBoth colors holiday day numbers and appends a dated legend (default).
	StyleBoth HolidayStyle = iota
	// StyleColor colors holiday day numbers only.
	StyleColor
	// StyleList appends a dated legend only.
	StyleList
	// StyleNone renders holidays plainly.
	StyleNone
)

// ParseHolidayStyle converts a CLI string ("both", "color", "list", "none")
// into a HolidayStyle.
func ParseHolidayStyle(s string) (HolidayStyle, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "both", "":
		return StyleBoth, nil
	case "color":
		return StyleColor, nil
	case "list":
		return StyleList, nil
	case "none":
		return StyleNone, nil
	default:
		return StyleBoth, fmt.Errorf("invalid holiday style %q (want both|color|list|none)", s)
	}
}

// wantsColor reports whether holiday day numbers should be colored.
func (s HolidayStyle) wantsColor() bool { return s == StyleBoth || s == StyleColor }

// wantsList reports whether a holiday legend should be appended.
func (s HolidayStyle) wantsList() bool { return s == StyleBoth || s == StyleList }

// Options configures rendering. Today is the reference date used for the
// highlight; Color enables ANSI escapes; HolidayStyle selects holiday marking.
type Options struct {
	Today        time.Time
	Color        bool
	HolidayStyle HolidayStyle
}

// RenderMonth renders a single month grid (title "Month Year") plus, depending
// on the holiday style, a legend of holidays in that month.
func RenderMonth(year int, month time.Month, opt Options) string {
	start := monthStart(year, month)
	grid := renderMonths(start, 1, 1, multiGap, true, opt)
	return withLegend(grid, start, 1, opt)
}

// RenderMonths renders count consecutive months starting at the month of start,
// arranged in the given number of columns (GNU cal -3 / -n style, titles show
// "Month Year"). A holiday legend for the whole range may be appended.
func RenderMonths(start time.Time, count, cols int, opt Options) string {
	start = monthStart(start.Year(), start.Month())
	grid := renderMonths(start, count, cols, multiGap, true, opt)
	return withLegend(grid, start, count, opt)
}

// RenderYear renders a full year (12 months, 3 columns) with a centered year
// header in the style of GNU cal -y, optionally followed by a holiday legend.
func RenderYear(year int, opt Options) string {
	start := monthStart(year, time.January)
	grid := renderMonths(start, 12, yearCols, yearGap, false, opt)

	total := yearCols*gridWidth + (yearCols-1)*yearGap
	header := center(fmt.Sprintf("%d", year), total)

	out := header + "\n\n" + grid
	return withLegend(out, start, 12, opt)
}

// renderMonths lays out count months starting at start into rows of `cols`
// columns separated by `gap` spaces. When showYear is true titles read
// "Month Year", otherwise just "Month".
func renderMonths(start time.Time, count, cols, gap int, showYear bool, opt Options) string {
	if cols < 1 {
		cols = 1
	}

	blocks := make([][]string, count)
	for i := 0; i < count; i++ {
		m := start.AddDate(0, i, 0)
		blocks[i] = monthBlock(m.Year(), m.Month(), showYear, opt)
	}

	var b strings.Builder
	sep := strings.Repeat(" ", gap)
	for row := 0; row < count; row += cols {
		end := row + cols
		if end > count {
			end = count
		}
		group := blocks[row:end]
		lines := len(group[0])
		for ln := 0; ln < lines; ln++ {
			parts := make([]string, len(group))
			for c := range group {
				parts[c] = group[c][ln]
			}
			b.WriteString(strings.Join(parts, sep))
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// monthBlock renders the lines of a single month: title, weekday header and
// weekRows week lines. Every line is exactly gridWidth columns of visible text.
func monthBlock(year int, month time.Month, showYear bool, opt Options) []string {
	title := month.String()
	if showYear {
		title = fmt.Sprintf("%s %d", month.String(), year)
	}

	lines := make([]string, 0, weekRows+2)
	lines = append(lines, center(title, gridWidth))
	lines = append(lines, dayHeader)

	first := monthStart(year, month)
	// Monday-first column index of the 1st (Monday=0 ... Sunday=6).
	lead := (int(first.Weekday()) - int(time.Monday) + 7) % 7
	daysIn := daysInMonth(year, month)

	day := 1
	for r := 0; r < weekRows; r++ {
		cells := make([]string, 7)
		for c := 0; c < 7; c++ {
			if (r == 0 && c < lead) || day > daysIn {
				cells[c] = strings.Repeat(" ", cellWidth)
				continue
			}
			cells[c] = renderCell(day, monthStart(year, month).AddDate(0, 0, day-1), opt)
			day++
		}
		lines = append(lines, strings.Join(cells, " "))
	}
	return lines
}

// renderCell renders a single day cell, applying date-based styling (gray for
// passed days, bold for weekends), holiday coloring and the today highlight
// where enabled.
//
// Styling rules (only when opt.Color is true; plain output is byte-identical
// to the color-off path):
//   - Foreground color layer: a holiday (when the style marks it) is red and
//     takes precedence over gray, so past holidays stay red rather than dim.
//     Otherwise a strictly-past day (compared by calendar date, not time of
//     day) is gray. Today is never gray.
//   - Bold weekends: Saturday/Sunday are bold regardless of holiday marking or
//     gray state, so a past weekday is gray, a past weekend is gray + bold, and
//     a weekend holiday is red + bold.
//   - The today reverse-video highlight wraps everything, so today on a weekend
//     is reverse video + bold.
func renderCell(day int, d time.Time, opt Options) string {
	cell := fmt.Sprintf("%*d", cellWidth, day)
	if !opt.Color {
		return cell
	}

	isHoliday, _ := holidays.IsHoliday(d)
	isToday := sameDay(d, opt.Today)

	// Foreground color: holiday red beats gray; otherwise dim passed days.
	switch {
	case opt.HolidayStyle.wantsColor() && isHoliday:
		cell = redOn + cell + redOff
	case beforeDay(d, opt.Today):
		cell = grayOn + cell + grayOff
	}

	// Weekend emphasis, independent of holiday marking and gray state.
	if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
		cell = boldOn + cell + boldOff
	}

	// Today highlight wraps any inner styling.
	if isToday {
		cell = reverseOn + cell + reverseOff
	}
	return cell
}

// withLegend appends a dated holiday legend for the given month range when the
// holiday style requests a list. The legend is emitted even when color is off.
func withLegend(grid string, start time.Time, count int, opt Options) string {
	if !opt.HolidayStyle.wantsList() {
		return grid
	}

	list := holidaysInRange(start, count)
	if len(list) == 0 {
		return grid
	}

	var b strings.Builder
	b.WriteString(grid)
	b.WriteString("\n\nHolidays:\n")
	for _, h := range list {
		fmt.Fprintf(&b, "%s  %s\n", h.Date.Format("2006-01-02"), h.Name)
	}
	return strings.TrimRight(b.String(), "\n")
}

// holidaysInRange returns holidays falling within the count months starting at
// start, sorted by date.
func holidaysInRange(start time.Time, count int) []holidays.Holiday {
	first := monthStart(start.Year(), start.Month())
	lastMonth := first.AddDate(0, count-1, 0)
	last := monthStart(lastMonth.Year(), lastMonth.Month()).
		AddDate(0, 1, -1) // last day of the final month

	var out []holidays.Holiday
	for y := first.Year(); y <= last.Year(); y++ {
		for _, h := range holidays.ForYear(y) {
			if !h.Date.Before(first) && !h.Date.After(last) {
				out = append(out, h)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out
}

// center pads s to width, centered, with the left bias GNU cal uses.
func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	left := (width - len(s) + 1) / 2
	right := width - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// monthStart returns midnight UTC on the first day of the given month.
func monthStart(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

// daysInMonth returns the number of days in the given month.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// sameDay reports whether a and b fall on the same calendar day.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// beforeDay reports whether a falls strictly before b by calendar date,
// ignoring the time of day.
func beforeDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	if ay != by {
		return ay < by
	}
	if am != bm {
		return am < bm
	}
	return ad < bd
}
