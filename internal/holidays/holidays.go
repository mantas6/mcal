// Package holidays provides Lithuanian public holidays (non-work days).
package holidays

import (
	"sort"
	"time"
)

// Holiday represents a single non-work day with its English name
// (Lithuanian term in parentheses where applicable).
type Holiday struct {
	Date time.Time
	Name string
}

// date builds a UTC midnight time.Time for the given year/month/day.
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// easterSunday computes the date of Easter Sunday for the given year using the
// Meeus/Jones/Butcher (anonymous Gregorian) algorithm.
func easterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return date(year, time.Month(month), day)
}

// firstSundayOfMonth returns the first Sunday of the given month/year.
func firstSundayOfMonth(year int, month time.Month) time.Time {
	d := date(year, month, 1)
	// Weekday(): Sunday == 0. Days to add to reach the first Sunday.
	offset := (int(time.Sunday) - int(d.Weekday()) + 7) % 7
	return d.AddDate(0, 0, offset)
}

// ForYear returns all Lithuanian public holidays (non-work days) for the given
// year, sorted by date.
func ForYear(year int) []Holiday {
	easter := easterSunday(year)

	list := []Holiday{
		{date(year, time.January, 1), "New Year's Day"},
		{date(year, time.February, 16), "Day of Restoration of the State of Lithuania"},
		{date(year, time.March, 11), "Day of Restoration of Independence"},
		{easter, "Easter Sunday"},
		{easter.AddDate(0, 0, 1), "Easter Monday"},
		{date(year, time.May, 1), "International Labour Day"},
		{firstSundayOfMonth(year, time.May), "Mother's Day"},
		{firstSundayOfMonth(year, time.June), "Father's Day"},
		{date(year, time.June, 24), "Midsummer Day (Joninės)"},
		{date(year, time.July, 6), "Statehood Day (King Mindaugas's Coronation Day)"},
		{date(year, time.August, 15), "Assumption Day (Žolinė)"},
		{date(year, time.November, 1), "All Saints' Day"},
		{date(year, time.November, 2), "All Souls' Day (Vėlinės)"},
		{date(year, time.December, 24), "Christmas Eve (Kūčios)"},
		{date(year, time.December, 25), "Christmas Day"},
		{date(year, time.December, 26), "Second Day of Christmas"},
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Date.Before(list[j].Date)
	})

	// Merge holidays that fall on the same date (e.g. in years where the first
	// Sunday of May coincides with International Labour Day) so the result has
	// no duplicate dates.
	merged := list[:0]
	for _, h := range list {
		if n := len(merged); n > 0 && merged[n-1].Date.Equal(h.Date) {
			merged[n-1].Name += " / " + h.Name
			continue
		}
		merged = append(merged, h)
	}

	return merged
}

// IsHoliday reports whether the given date is a Lithuanian public holiday and,
// if so, returns its name. Only the year/month/day components are considered.
func IsHoliday(d time.Time) (bool, string) {
	y, m, day := d.Date()
	for _, h := range ForYear(y) {
		hy, hm, hd := h.Date.Date()
		if hy == y && hm == m && hd == day {
			return true, h.Name
		}
	}
	return false, ""
}
