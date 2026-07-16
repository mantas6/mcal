package holidays

import (
	"strings"
	"testing"
	"time"
)

func mkDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func find(hs []Holiday, name string) (Holiday, bool) {
	for _, h := range hs {
		if h.Name == name {
			return h, true
		}
	}
	return Holiday{}, false
}

func TestEasterSunday(t *testing.T) {
	cases := []struct {
		year  int
		month time.Month
		day   int
	}{
		{2024, time.March, 31},
		{2025, time.April, 20},
		{2026, time.April, 5},
		{2027, time.March, 28},
		{2000, time.April, 23},
	}

	for _, c := range cases {
		got := easterSunday(c.year)
		want := mkDate(c.year, c.month, c.day)
		if !got.Equal(want) {
			t.Errorf("easterSunday(%d) = %s, want %s", c.year, got.Format("2006-01-02"), want.Format("2006-01-02"))
		}

		// Easter Monday must be Easter Sunday + 1 day.
		hs := ForYear(c.year)
		sun, ok := find(hs, "Easter Sunday")
		if !ok {
			t.Fatalf("ForYear(%d): Easter Sunday missing", c.year)
		}
		mon, ok := find(hs, "Easter Monday")
		if !ok {
			t.Fatalf("ForYear(%d): Easter Monday missing", c.year)
		}
		if !mon.Date.Equal(sun.Date.AddDate(0, 0, 1)) {
			t.Errorf("Easter Monday %s != Easter Sunday %s + 1 day",
				mon.Date.Format("2006-01-02"), sun.Date.Format("2006-01-02"))
		}
	}
}

func TestFixedHolidays(t *testing.T) {
	const year = 2025
	hs := ForYear(year)

	cases := []struct {
		name  string
		month time.Month
		day   int
	}{
		{"New Year's Day", time.January, 1},
		{"Day of Restoration of the State of Lithuania", time.February, 16},
		{"Day of Restoration of Independence", time.March, 11},
		{"International Labour Day", time.May, 1},
		{"Midsummer Day (Joninės)", time.June, 24},
		{"Statehood Day (King Mindaugas's Coronation Day)", time.July, 6},
		{"Assumption Day (Žolinė)", time.August, 15},
		{"All Saints' Day", time.November, 1},
		{"All Souls' Day (Vėlinės)", time.November, 2},
		{"Christmas Eve (Kūčios)", time.December, 24},
		{"Christmas Day", time.December, 25},
		{"Second Day of Christmas", time.December, 26},
	}

	for _, c := range cases {
		h, ok := find(hs, c.name)
		if !ok {
			t.Errorf("holiday %q not found", c.name)
			continue
		}
		want := mkDate(year, c.month, c.day)
		if !h.Date.Equal(want) {
			t.Errorf("holiday %q = %s, want %s", c.name,
				h.Date.Format("2006-01-02"), want.Format("2006-01-02"))
		}
	}
}

// firstSunday returns the first Sunday of the given month/year, computed
// independently of the package implementation.
func firstSunday(year int, month time.Month) time.Time {
	d := mkDate(year, month, 1)
	for d.Weekday() != time.Sunday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

func TestMothersAndFathersDay(t *testing.T) {
	for year := 2020; year <= 2030; year++ {
		wantMom := firstSunday(year, time.May)
		ok, name := IsHoliday(wantMom)
		if !ok || !strings.Contains(name, "Mother's Day") {
			t.Errorf("year %d: expected Mother's Day on %s, got ok=%v name=%q",
				year, wantMom.Format("2006-01-02"), ok, name)
		}
		if wantMom.Day() > 7 {
			t.Errorf("year %d: first Sunday of May computed wrong: %s", year, wantMom.Format("2006-01-02"))
		}

		wantDad := firstSunday(year, time.June)
		ok, name = IsHoliday(wantDad)
		if !ok || !strings.Contains(name, "Father's Day") {
			t.Errorf("year %d: expected Father's Day on %s, got ok=%v name=%q",
				year, wantDad.Format("2006-01-02"), ok, name)
		}
		if wantDad.Day() > 7 {
			t.Errorf("year %d: first Sunday of June computed wrong: %s", year, wantDad.Format("2006-01-02"))
		}
	}
}

func TestIsHoliday(t *testing.T) {
	positive := []struct {
		date time.Time
		name string
	}{
		{mkDate(2025, time.January, 1), "New Year's Day"},
		{mkDate(2025, time.December, 25), "Christmas Day"},
		{mkDate(2024, time.March, 31), "Easter Sunday"},
		{mkDate(2024, time.April, 1), "Easter Monday"},
	}
	for _, c := range positive {
		ok, name := IsHoliday(c.date)
		if !ok {
			t.Errorf("IsHoliday(%s) = false, want true", c.date.Format("2006-01-02"))
			continue
		}
		if name != c.name {
			t.Errorf("IsHoliday(%s) name = %q, want %q", c.date.Format("2006-01-02"), name, c.name)
		}
	}

	negative := []time.Time{
		mkDate(2025, time.January, 2),
		mkDate(2025, time.July, 7),
		mkDate(2025, time.April, 15),
	}
	for _, d := range negative {
		ok, name := IsHoliday(d)
		if ok {
			t.Errorf("IsHoliday(%s) = true (%q), want false", d.Format("2006-01-02"), name)
		}
	}

	// Time-of-day should not matter.
	withTime := time.Date(2025, time.December, 25, 13, 30, 0, 0, time.UTC)
	if ok, _ := IsHoliday(withTime); !ok {
		t.Errorf("IsHoliday with time-of-day should still detect Christmas Day")
	}
}

func TestForYearNoDuplicatesAndCount(t *testing.T) {
	for year := 2020; year <= 2030; year++ {
		hs := ForYear(year)

		// Sorted ascending.
		for i := 1; i < len(hs); i++ {
			if hs[i].Date.Before(hs[i-1].Date) {
				t.Errorf("ForYear(%d) not sorted at index %d", year, i)
			}
		}

		// No duplicate dates (same-date holidays are merged by ForYear).
		seen := make(map[string]bool)
		for _, h := range hs {
			key := h.Date.Format("2006-01-02")
			if seen[key] {
				t.Errorf("ForYear(%d) duplicate date %s (%q)", year, key, h.Name)
			}
			seen[key] = true
		}
	}

	// A typical year has 16 distinct non-work days.
	if got := len(ForYear(2025)); got != 16 {
		t.Errorf("ForYear(2025) count = %d, want 16", got)
	}

	// In 2022 the first Sunday of May is May 1, so Mother's Day merges with
	// International Labour Day, leaving 15 distinct dates.
	hs2022 := ForYear(2022)
	if got := len(hs2022); got != 15 {
		t.Errorf("ForYear(2022) count = %d, want 15", got)
	}
	if _, ok := find(hs2022, "International Labour Day / Mother's Day"); !ok {
		t.Errorf("ForYear(2022) missing merged Labour Day / Mother's Day entry")
	}
}
