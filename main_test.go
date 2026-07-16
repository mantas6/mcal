package main

import (
	"testing"
	"time"

	"github.com/mantas6/mcal/internal/calendar"
)

// today is a fixed reference date used across the parsing tests.
var today = time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC)

func TestParseArgsValid(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Config
	}{
		{
			name: "default no args",
			args: nil,
			want: Config{View: ViewMonth, Year: 2026, Month: time.July, HolidayStyle: calendar.StyleBoth},
		},
		{
			name: "single month and year",
			args: []string{"2", "2026"},
			want: Config{View: ViewMonth, Year: 2026, Month: time.February, HolidayStyle: calendar.StyleBoth},
		},
		{
			name: "year alone is full year",
			args: []string{"2026"},
			want: Config{View: ViewYear, Year: 2026, HolidayStyle: calendar.StyleBoth},
		},
		{
			name: "-y full year uses today's year",
			args: []string{"-y"},
			want: Config{View: ViewYear, Year: 2026, HolidayStyle: calendar.StyleBoth},
		},
		{
			name: "-3 previous current next",
			args: []string{"-3"},
			want: Config{
				View:         ViewMonths,
				Start:        time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
				Count:        3,
				HolidayStyle: calendar.StyleBoth,
			},
		},
		{
			name: "-n count from today",
			args: []string{"-n", "6"},
			want: Config{
				View:         ViewMonths,
				Start:        time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				Count:        6,
				HolidayStyle: calendar.StyleBoth,
			},
		},
		{
			name: "-n count with year starts in January",
			args: []string{"-n", "6", "2026"},
			want: Config{
				View:         ViewMonths,
				Start:        time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Count:        6,
				HolidayStyle: calendar.StyleBoth,
			},
		},
		{
			name: "--months long form",
			args: []string{"--months", "4"},
			want: Config{
				View:         ViewMonths,
				Start:        time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				Count:        4,
				HolidayStyle: calendar.StyleBoth,
			},
		},
		{
			name: "--months= form",
			args: []string{"--months=4"},
			want: Config{
				View:         ViewMonths,
				Start:        time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				Count:        4,
				HolidayStyle: calendar.StyleBoth,
			},
		},
		{
			name: "--holidays default current year",
			args: []string{"--holidays"},
			want: Config{View: ViewHolidays, HolidayYear: 2026, HolidayStyle: calendar.StyleBoth},
		},
		{
			name: "--holidays explicit year",
			args: []string{"--holidays", "2030"},
			want: Config{View: ViewHolidays, HolidayYear: 2030, HolidayStyle: calendar.StyleBoth},
		},
		{
			name: "--holiday-style equals form",
			args: []string{"--holiday-style=list"},
			want: Config{View: ViewMonth, Year: 2026, Month: time.July, HolidayStyle: calendar.StyleList},
		},
		{
			name: "--holiday-style space form",
			args: []string{"--holiday-style", "none"},
			want: Config{View: ViewMonth, Year: 2026, Month: time.July, HolidayStyle: calendar.StyleNone},
		},
		{
			name: "--no-color",
			args: []string{"--no-color"},
			want: Config{View: ViewMonth, Year: 2026, Month: time.July, HolidayStyle: calendar.StyleBoth, NoColor: true},
		},
		{
			name: "help",
			args: []string{"--help"},
			want: Config{View: ViewHelp, HolidayStyle: calendar.StyleBoth},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args, today)
			if err != nil {
				t.Fatalf("ParseArgs(%v) unexpected error: %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("ParseArgs(%v)\n got  = %+v\n want = %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseArgsInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"bad month", []string{"13", "2026"}},
		{"month zero", []string{"0", "2026"}},
		{"non-numeric positional", []string{"abc"}},
		{"-n zero", []string{"-n", "0"}},
		{"-n negative", []string{"-n", "-1"}},
		{"-n non-numeric", []string{"-n", "foo"}},
		{"-n missing value", []string{"-n"}},
		{"too many positionals", []string{"1", "2", "2026"}},
		{"bad holiday style", []string{"--holiday-style", "rainbow"}},
		{"unknown flag", []string{"--nope"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseArgs(tt.args, today); err == nil {
				t.Errorf("ParseArgs(%v) = nil error, want error", tt.args)
			}
		})
	}
}

func TestParseArgsPrecedence(t *testing.T) {
	// -y beats everything else.
	got, err := ParseArgs([]string{"-y", "-n", "6", "-3"}, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.View != ViewYear {
		t.Errorf("-y precedence: got view %v, want ViewYear", got.View)
	}

	// -n beats -3.
	got, err = ParseArgs([]string{"-n", "6", "-3"}, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.View != ViewMonths || got.Count != 6 {
		t.Errorf("-n over -3: got view %v count %d, want ViewMonths count 6", got.View, got.Count)
	}

	// -3 beats a single month.
	got, err = ParseArgs([]string{"-3", "2", "2026"}, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantStart := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	if got.View != ViewMonths || got.Count != 3 || !got.Start.Equal(wantStart) {
		t.Errorf("-3 over single month: got %+v, want ViewMonths start %v count 3", got, wantStart)
	}
}

func TestRenderHolidays(t *testing.T) {
	out := Render(Config{View: ViewHolidays, HolidayYear: 2026}, today, false)
	if !contains(out, "Lithuanian holidays 2026:") {
		t.Errorf("holiday output missing header:\n%s", out)
	}
	if !contains(out, "2026-01-01  New Year's Day") {
		t.Errorf("holiday output missing New Year's Day:\n%s", out)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
