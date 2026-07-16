package calendar

import (
	"strings"
	"testing"
	"time"
)

// refDay is an arbitrary reference "today" far from the dates under test so it
// never triggers the highlight unless a test intends it to.
var refDay = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func plainOpts() Options {
	return Options{Today: refDay, Color: false, HolidayStyle: StyleNone}
}

func TestRenderMonthExactLayout(t *testing.T) {
	want := strings.Join([]string{
		"      July 2025     ",
		"Mo Tu We Th Fr Sa Su",
		"    1  2  3  4  5  6",
		" 7  8  9 10 11 12 13",
		"14 15 16 17 18 19 20",
		"21 22 23 24 25 26 27",
		"28 29 30 31         ",
		"                    ",
	}, "\n")

	got := RenderMonth(2025, time.July, plainOpts())
	if got != want {
		t.Errorf("month grid mismatch\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestMondayAlignmentSundayFirst(t *testing.T) {
	// June 2025 starts on a Sunday, so the 1 must land in the last column.
	got := RenderMonth(2025, time.June, plainOpts())
	lines := strings.Split(got, "\n")

	if lines[1] != dayHeader {
		t.Fatalf("unexpected header %q", lines[1])
	}
	// First week row: six blank cells then " 1" in the Sunday (last) column.
	wantRow := "                   1"
	if lines[2] != wantRow {
		t.Errorf("first week row = %q, want %q", lines[2], wantRow)
	}
}

func TestHolidayStyleList(t *testing.T) {
	opt := Options{Today: refDay, Color: false, HolidayStyle: StyleList}
	got := RenderMonth(2025, time.January, opt)

	if !strings.Contains(got, "Holidays:") {
		t.Errorf("list style missing legend header:\n%s", got)
	}
	if !strings.Contains(got, "2025-01-01  New Year's Day") {
		t.Errorf("list style missing dated holiday entry:\n%s", got)
	}
	// No color escapes when color is off.
	if strings.Contains(got, "\x1b[") {
		t.Errorf("list style with color off must not emit escapes:\n%q", got)
	}
}

func TestHolidayStyleColor(t *testing.T) {
	opt := Options{Today: refDay, Color: true, HolidayStyle: StyleColor}
	got := RenderMonth(2025, time.January, opt)

	// New Year's Day (Jan 1) must be wrapped in the red escape.
	if !strings.Contains(got, redOn+" 1"+redOff) {
		t.Errorf("color style did not wrap holiday day number:\n%q", got)
	}
	// color style does not append a legend.
	if strings.Contains(got, "Holidays:") {
		t.Errorf("color style must not append legend:\n%s", got)
	}
}

func TestHolidayStyleNone(t *testing.T) {
	opt := Options{Today: refDay, Color: true, HolidayStyle: StyleNone}
	got := RenderMonth(2025, time.January, opt)

	if strings.Contains(got, redOn) {
		t.Errorf("none style must not color holidays:\n%q", got)
	}
	if strings.Contains(got, "Holidays:") {
		t.Errorf("none style must not append legend:\n%s", got)
	}
}

func TestHolidayStyleBothDegradesWithoutColor(t *testing.T) {
	opt := Options{Today: refDay, Color: false, HolidayStyle: StyleBoth}
	got := RenderMonth(2025, time.January, opt)

	if !strings.Contains(got, "2025-01-01  New Year's Day") {
		t.Errorf("both style must still list holidays without color:\n%s", got)
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("both style with color off must not emit escapes:\n%q", got)
	}
}

func TestRenderMonthsSpansYearBoundary(t *testing.T) {
	// November 2025 + 4 months -> Nov, Dec (2025), Jan, Feb (2026).
	start := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	got := RenderMonths(start, 4, 3, plainOpts())

	for _, title := range []string{
		"November 2025", "December 2025", "January 2026", "February 2026",
	} {
		if !strings.Contains(got, title) {
			t.Errorf("missing month title %q in:\n%s", title, got)
		}
	}
}

func TestRenderThreeMonthsHeaders(t *testing.T) {
	start := time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC)
	got := RenderMonths(start, 3, 3, plainOpts())

	for _, title := range []string{"June 2025", "July 2025", "August 2025"} {
		if !strings.Contains(got, title) {
			t.Errorf("missing month title %q in:\n%s", title, got)
		}
	}
	// Three months in three columns -> the header row lists three weekday rows.
	firstLine := strings.SplitN(got, "\n", 2)[0]
	if strings.Count(firstLine, "June 2025") != 1 {
		t.Errorf("expected titles on a single row, got:\n%s", firstLine)
	}
}

func TestRenderYearHeadersAndTitle(t *testing.T) {
	got := RenderYear(2025, plainOpts())
	lines := strings.Split(got, "\n")

	// Centered year header on the first line.
	if strings.TrimSpace(lines[0]) != "2025" {
		t.Errorf("first line should be centered year, got %q", lines[0])
	}
	if lines[1] != "" {
		t.Errorf("second line should be blank, got %q", lines[1])
	}

	// All twelve month names must appear.
	for m := time.January; m <= time.December; m++ {
		if !strings.Contains(got, m.String()) {
			t.Errorf("year view missing month %q", m.String())
		}
	}
	// Year-view month titles are month-only (no year suffix on the header row).
	if strings.Contains(got, "January 2025") {
		t.Errorf("year view titles must not include the year:\n%s", got)
	}
}

func TestTodayHighlightColorOnOff(t *testing.T) {
	today := time.Date(2025, time.July, 15, 0, 0, 0, 0, time.UTC)

	on := RenderMonth(2025, time.July, Options{Today: today, Color: true, HolidayStyle: StyleNone})
	if !strings.Contains(on, reverseOn+"15"+reverseOff) {
		t.Errorf("today not reverse-highlighted with color on:\n%q", on)
	}

	off := RenderMonth(2025, time.July, Options{Today: today, Color: false, HolidayStyle: StyleNone})
	if strings.Contains(off, reverseOn) {
		t.Errorf("today must not be highlighted with color off:\n%q", off)
	}
	// With color off the grid is plain text.
	if strings.Contains(off, "\x1b[") {
		t.Errorf("no escapes expected with color off:\n%q", off)
	}
}

func TestPastDaysGray(t *testing.T) {
	// July 2025: 14th is a past weekday, 15th is today, 16th is future.
	today := time.Date(2025, time.July, 15, 0, 0, 0, 0, time.UTC)
	got := RenderMonth(2025, time.July, Options{Today: today, Color: true, HolidayStyle: StyleNone})

	// Past weekday (Mon Jul 14) must be gray.
	if !strings.Contains(got, grayOn+"14"+grayOff) {
		t.Errorf("past day not grayed:\n%q", got)
	}
	// Today (Jul 15) must not be gray; it keeps the reverse highlight.
	if strings.Contains(got, grayOn+"15") {
		t.Errorf("today must not be grayed:\n%q", got)
	}
	if !strings.Contains(got, reverseOn+"15"+reverseOff) {
		t.Errorf("today must keep reverse highlight:\n%q", got)
	}
	// Future weekday (Wed Jul 16) must not be gray.
	if strings.Contains(got, grayOn+"16") {
		t.Errorf("future day must not be grayed:\n%q", got)
	}
}

func TestWeekendBold(t *testing.T) {
	// Reference today far in the past so no gray styling interferes.
	got := RenderMonth(2025, time.July, Options{Today: refDay, Color: true, HolidayStyle: StyleNone})

	// Saturday Jul 5 must be bold.
	if !strings.Contains(got, boldOn+" 5"+boldOff) {
		t.Errorf("Saturday not bold:\n%q", got)
	}
	// Sunday Jul 6 must be bold.
	if !strings.Contains(got, boldOn+" 6"+boldOff) {
		t.Errorf("Sunday not bold:\n%q", got)
	}
	// Weekday Tue Jul 1 must not be bold.
	if strings.Contains(got, boldOn+" 1") {
		t.Errorf("weekday must not be bold:\n%q", got)
	}
}

func TestPastWeekendGrayAndBold(t *testing.T) {
	// today Jul 15 2025 -> Sat Jul 5 is a past weekend: gray + bold.
	today := time.Date(2025, time.July, 15, 0, 0, 0, 0, time.UTC)
	got := RenderMonth(2025, time.July, Options{Today: today, Color: true, HolidayStyle: StyleNone})

	if !strings.Contains(got, boldOn+grayOn+" 5"+grayOff+boldOff) {
		t.Errorf("past weekend not gray+bold:\n%q", got)
	}
}

func TestTodayOnWeekendReverseAndBold(t *testing.T) {
	// today Sat Jul 19 2025 -> reverse video + bold, and not gray.
	today := time.Date(2025, time.July, 19, 0, 0, 0, 0, time.UTC)
	got := RenderMonth(2025, time.July, Options{Today: today, Color: true, HolidayStyle: StyleNone})

	if !strings.Contains(got, reverseOn+boldOn+"19"+boldOff+reverseOff) {
		t.Errorf("today-on-weekend not reverse+bold:\n%q", got)
	}
	if strings.Contains(got, grayOn+"19") {
		t.Errorf("today must not be grayed:\n%q", got)
	}
}

func TestPastHolidayStaysRedNotGray(t *testing.T) {
	// today Jan 15 2025 -> New Year's Day (Wed Jan 1) is a past holiday.
	// It must stay red, not gray (red precedence). It is a weekday, so no bold.
	today := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	got := RenderMonth(2025, time.January, Options{Today: today, Color: true, HolidayStyle: StyleColor})

	if !strings.Contains(got, redOn+" 1"+redOff) {
		t.Errorf("past holiday must stay red:\n%q", got)
	}
	if strings.Contains(got, grayOn+" 1") {
		t.Errorf("past holiday must not be grayed:\n%q", got)
	}
}

func TestDateStylingColorOff(t *testing.T) {
	// With color off, no gray/bold escapes even for past weekend days.
	today := time.Date(2025, time.July, 15, 0, 0, 0, 0, time.UTC)
	got := RenderMonth(2025, time.July, Options{Today: today, Color: false, HolidayStyle: StyleNone})

	if strings.Contains(got, "\x1b[") {
		t.Errorf("color off must not emit escapes:\n%q", got)
	}
}

func TestParseHolidayStyle(t *testing.T) {
	cases := map[string]HolidayStyle{
		"both":  StyleBoth,
		"":      StyleBoth,
		"color": StyleColor,
		"LIST":  StyleList,
		"none":  StyleNone,
	}
	for in, want := range cases {
		got, err := ParseHolidayStyle(in)
		if err != nil {
			t.Errorf("ParseHolidayStyle(%q) unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("ParseHolidayStyle(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := ParseHolidayStyle("bogus"); err == nil {
		t.Errorf("ParseHolidayStyle(\"bogus\") expected error")
	}
}
