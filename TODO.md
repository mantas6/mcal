# mcal — implementation plan

A GNU `cal` alternative in Go for the **Lithuanian calendar**, output in **English**,
showing Lithuanian public holidays / non-work days.

Module: `github.com/mantas6/mcal` — stdlib only.

## Layout

```
main.go                        CLI dispatch + TTY/clock detection
internal/holidays/holidays.go  + holidays_test.go
internal/calendar/calendar.go  + calendar_test.go
README.md                      usage + holiday list
```

## Holidays (`internal/holidays`)

- `Holiday{ Date time.Time; Name string }`, `ForYear(year int) []Holiday`, `IsHoliday(date) (bool, string)`.
- English names with the Lithuanian term in parentheses.

### Fixed non-work days
- [ ] Jan 1 — New Year's Day
- [ ] Feb 16 — Day of Restoration of the State of Lithuania
- [ ] Mar 11 — Day of Restoration of Independence
- [ ] May 1 — International Labour Day
- [ ] Jun 24 — Midsummer Day (Joninės)
- [ ] Jul 6 — Statehood Day
- [ ] Aug 15 — Assumption Day (Žolinė)
- [ ] Nov 1 — All Saints' Day
- [ ] Nov 2 — All Souls' Day (Vėlinės)
- [ ] Dec 24 — Christmas Eve (Kūčios)
- [ ] Dec 25 — Christmas Day
- [ ] Dec 26 — Second Day of Christmas

### Movable
- [ ] Easter Sunday — Meeus/Jones/Butcher (anonymous Gregorian) algorithm
- [ ] Easter Monday — Easter Sunday + 1
- [ ] Mother's Day — 1st Sunday of May
- [ ] Father's Day — 1st Sunday of June

## Rendering (`internal/calendar`)

- [ ] Monday-first weeks, English headers (`Mo Tu We Th Fr Sa Su`).
- [ ] Single month grid matching GNU `cal` spacing/centering.
- [ ] Shared `renderMonths(start, count, cols=3)` powering `-y` (12), `-n` (N), `-3` (start-1, 3).
- [ ] Today highlight (reverse video).
- [ ] Holiday marking via `--holiday-style=both|color|list|none` (default `both`).
- [ ] Color auto-off on non-TTY / `NO_COLOR` / `--no-color`.
- [ ] Render fns take injected reference date + color bool (deterministic).

## CLI (`main.go`, parsing in testable `ParseArgs`)

- [ ] default → current month
- [ ] `<month> <year>` / `<year>` → specific month / full year
- [ ] `-3`, `-y`, `-n <count>` / `--months <count>`
- [ ] `--holidays [year]` → dated holiday list only
- [ ] `--holiday-style`, `--no-color`, `-h/--help`
- [ ] view precedence: `-y` > `-n` > `-3` > single month
- [ ] `-n` start = resolved date (Jan if only year given), replicating GNU `cal`
- [ ] validate: bad month (13), bad `-n` (0/neg/non-numeric), non-numeric args → error

## Tests

### holidays_test.go
- [ ] Easter table: 2024-03-31, 2025-04-20, 2026-04-05, 2027-03-28, 2000-04-23; Monday = +1
- [ ] Fixed holidays present with correct names
- [ ] Mother's/Father's Day land on 1st Sunday of May/June
- [ ] `IsHoliday` positive + negative cases
- [ ] No duplicate dates / expected count

### calendar_test.go
- [ ] Month grid exact-string layout (color off)
- [ ] Monday alignment: 1st on a Sunday → last column
- [ ] Holiday styles: `list` legend text, `color` escape wrapping, `none` = plain
- [ ] `-n` spans year boundary (Nov +4 → Nov, Dec, Jan, Feb)
- [ ] `-3` / `-y` blocks and headers present
- [ ] Color on/off wrapper behavior

### ParseArgs
- [ ] `-n 6`, `-n 6 2026`, `-3`, `-y`, `--holidays`, `--holiday-style`, `--no-color`
- [ ] invalid: `13 2026`, `abc`, `-n 0`, `-n -1`, `-n foo`

## Verification
- [ ] `go test ./... -v`
- [ ] `go vet ./...`
- [ ] `gofmt -l .` clean
