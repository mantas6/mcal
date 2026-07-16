// Command mcal is a GNU cal alternative for the Lithuanian calendar, rendering
// Monday-first month, multi-month and year views in English with Lithuanian
// public holidays marked.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mantas6/mcal/internal/calendar"
	"github.com/mantas6/mcal/internal/holidays"
)

// View selects what mcal renders.
type View int

const (
	// ViewMonth renders a single month.
	ViewMonth View = iota
	// ViewMonths renders Count consecutive months starting at Start (-3 / -n).
	ViewMonths
	// ViewYear renders a full year.
	ViewYear
	// ViewHolidays prints the dated holiday list for HolidayYear.
	ViewHolidays
	// ViewHelp prints usage text.
	ViewHelp
)

// Config is the fully resolved result of parsing command-line arguments.
type Config struct {
	View View

	// Year and Month are used by ViewMonth (both) and ViewYear (Year only).
	Year  int
	Month time.Month

	// Start and Count drive ViewMonths.
	Start time.Time
	Count int

	// HolidayYear is the year printed by ViewHolidays.
	HolidayYear int

	HolidayStyle calendar.HolidayStyle
	NoColor      bool
}

const usage = `mcal — a GNU cal alternative for the Lithuanian calendar (holidays in English)

Usage:
  mcal [options] [[month] year]

Arguments:
  month year          show the given month (1-12) of the given year
  year                show the full given year

Options:
  -3                  previous, current and next month
  -y, --year          show the whole year
  -n, --months N      show N months starting from the resolved date
  --holidays [year]   print the dated holiday list for the year (default: current)
  --holiday-style S   holiday marking: both|color|list|none (default: both)
  --no-color          disable ANSI color output
  -h, --help          show this help

With no arguments, the current month is shown.
View precedence when combined: -y > -n > -3 > single month.
`

// ParseArgs parses command-line arguments (excluding the program name) into a
// Config, resolving relative views against today. It never reads the
// environment or touches I/O, so it is fully deterministic and testable.
func ParseArgs(args []string, today time.Time) (Config, error) {
	var (
		has3, hasY, hasN    bool
		hasHolidays         bool
		noColor, help       bool
		nCount              int
		styleStr            string
		positionals         []int
		holidayYear         int
		holidayYearExplicit bool
	)

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			help = true
		case a == "-y" || a == "--year":
			hasY = true
		case a == "-3":
			has3 = true
		case a == "--no-color":
			noColor = true
		case a == "--holidays":
			hasHolidays = true
			// Optionally consume a following numeric year.
			if i+1 < len(args) {
				if y, err := strconv.Atoi(args[i+1]); err == nil {
					holidayYear = y
					holidayYearExplicit = true
					i++
				}
			}
		case a == "-n" || a == "--months":
			if i+1 >= len(args) {
				return Config{}, fmt.Errorf("option %s requires a count argument", a)
			}
			i++
			n, err := parseCount(args[i])
			if err != nil {
				return Config{}, err
			}
			nCount = n
			hasN = true
		case strings.HasPrefix(a, "--months="):
			n, err := parseCount(strings.TrimPrefix(a, "--months="))
			if err != nil {
				return Config{}, err
			}
			nCount = n
			hasN = true
		case strings.HasPrefix(a, "--holiday-style="):
			styleStr = strings.TrimPrefix(a, "--holiday-style=")
		case a == "--holiday-style":
			if i+1 >= len(args) {
				return Config{}, fmt.Errorf("option %s requires a value", a)
			}
			i++
			styleStr = args[i]
		case strings.HasPrefix(a, "-") && a != "-":
			return Config{}, fmt.Errorf("unknown option %q", a)
		default:
			n, err := strconv.Atoi(a)
			if err != nil {
				return Config{}, fmt.Errorf("invalid argument %q: expected a number", a)
			}
			positionals = append(positionals, n)
		}
	}

	style, err := calendar.ParseHolidayStyle(styleStr)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{HolidayStyle: style, NoColor: noColor}

	if help {
		cfg.View = ViewHelp
		return cfg, nil
	}

	// Resolve the reference date from positional arguments.
	var (
		resYear    = today.Year()
		resMonth   = today.Month()
		yearGiven  bool
		monthGiven bool
	)
	switch len(positionals) {
	case 0:
		// Use today.
	case 1:
		resYear = positionals[0]
		resMonth = time.January // -n start = January when only a year is given
		yearGiven = true
	case 2:
		m := positionals[0]
		if m < 1 || m > 12 {
			return Config{}, fmt.Errorf("invalid month %d: want 1-12", m)
		}
		resMonth = time.Month(m)
		resYear = positionals[1]
		monthGiven = true
		yearGiven = true
	default:
		return Config{}, fmt.Errorf("too many arguments: expected at most 'month year'")
	}

	if hasHolidays {
		cfg.View = ViewHolidays
		switch {
		case holidayYearExplicit:
			cfg.HolidayYear = holidayYear
		case yearGiven:
			cfg.HolidayYear = resYear
		default:
			cfg.HolidayYear = today.Year()
		}
		return cfg, nil
	}

	start := time.Date(resYear, resMonth, 1, 0, 0, 0, 0, time.UTC)

	switch {
	case hasY:
		cfg.View = ViewYear
		cfg.Year = resYear
	case hasN:
		cfg.View = ViewMonths
		cfg.Start = start
		cfg.Count = nCount
	case has3:
		cfg.View = ViewMonths
		cfg.Start = start.AddDate(0, -1, 0)
		cfg.Count = 3
	case monthGiven:
		cfg.View = ViewMonth
		cfg.Year = resYear
		cfg.Month = resMonth
	case yearGiven:
		cfg.View = ViewYear
		cfg.Year = resYear
	default:
		cfg.View = ViewMonth
		cfg.Year = resYear
		cfg.Month = resMonth
	}

	return cfg, nil
}

// parseCount parses and validates an -n / --months count (must be >= 1).
func parseCount(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid count %q: expected a number", s)
	}
	if n < 1 {
		return 0, fmt.Errorf("invalid count %d: must be positive", n)
	}
	return n, nil
}

// Render produces the output for cfg, injecting today and color into the
// calendar options. It returns the rendered string without a trailing newline.
func Render(cfg Config, today time.Time, color bool) string {
	opt := calendar.Options{
		Today:        today,
		Color:        color,
		HolidayStyle: cfg.HolidayStyle,
	}

	switch cfg.View {
	case ViewHelp:
		return strings.TrimRight(usage, "\n")
	case ViewHolidays:
		return renderHolidays(cfg.HolidayYear)
	case ViewYear:
		return calendar.RenderYear(cfg.Year, opt)
	case ViewMonths:
		return calendar.RenderMonths(cfg.Start, cfg.Count, 3, opt)
	default:
		return calendar.RenderMonth(cfg.Year, cfg.Month, opt)
	}
}

// renderHolidays formats the dated holiday list for the given year.
func renderHolidays(year int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Lithuanian holidays %d:\n", year)
	for _, h := range holidays.ForYear(year) {
		fmt.Fprintf(&b, "%s  %s\n", h.Date.Format("2006-01-02"), h.Name)
	}
	return strings.TrimRight(b.String(), "\n")
}

// wantColor decides whether ANSI color should be emitted, honoring the
// --no-color flag, the NO_COLOR environment variable and whether stdout is a
// character device (TTY).
func wantColor(noColor bool) bool {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func main() {
	cfg, err := ParseArgs(os.Args[1:], time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcal:", err)
		fmt.Fprintln(os.Stderr, "run 'mcal --help' for usage")
		os.Exit(2)
	}

	if cfg.View == ViewHelp {
		fmt.Println(Render(cfg, time.Now(), false))
		return
	}

	color := wantColor(cfg.NoColor)
	fmt.Println(Render(cfg, time.Now(), color))
}
