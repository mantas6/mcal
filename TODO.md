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
- [x] Jan 1 — New Year's Day
- [x] Feb 16 — Day of Restoration of the State of Lithuania
- [x] Mar 11 — Day of Restoration of Independence
- [x] May 1 — International Labour Day
- [x] Jun 24 — Midsummer Day (Joninės)
- [x] Jul 6 — Statehood Day (King Mindaugas's Coronation Day)
- [x] Aug 15 — Assumption Day (Žolinė)
- [x] Nov 1 — All Saints' Day
- [x] Nov 2 — All Souls' Day (Vėlinės)
- [x] Dec 24 — Christmas Eve (Kūčios)
- [x] Dec 25 — Christmas Day
- [x] Dec 26 — Second Day of Christmas

### Movable
- [x] Easter Sunday — Meeus/Jones/Butcher (anonymous Gregorian) algorithm
- [x] Easter Monday — Easter Sunday + 1
- [x] Mother's Day — 1st Sunday of May
- [x] Father's Day — 1st Sunday of June

## Rendering (`internal/calendar`)

- [x] Monday-first weeks, English headers (`Mo Tu We Th Fr Sa Su`).
- [x] Single month grid matching GNU `cal` spacing/centering.
- [x] Shared `renderMonths(start, count, cols=3)` powering `-y` (12), `-n` (N), `-3` (start-1, 3).
- [x] Today highlight (reverse video).
- [x] Holiday marking via `--holiday-style=both|color|list|none` (default `both`).
- [x] Color auto-off on non-TTY / `NO_COLOR` / `--no-color`.
- [x] Render fns take injected reference date + color bool (deterministic).

## CLI (`main.go`, parsing in testable `ParseArgs`)

- [x] default → current month
- [x] `<month> <year>` / `<year>` → specific month / full year
- [x] `-3`, `-y`, `-n <count>` / `--months <count>`
- [x] `--holidays [year]` → dated holiday list only
- [x] `--holiday-style`, `--no-color`, `-h/--help`
- [x] view precedence: `-y` > `-n` > `-3` > single month
- [x] `-n` start = resolved date (Jan if only year given), replicating GNU `cal`
- [x] validate: bad month (13), bad `-n` (0/neg/non-numeric), non-numeric args → error

## Tests

### holidays_test.go
- [x] Easter table: 2024-03-31, 2025-04-20, 2026-04-05, 2027-03-28, 2000-04-23; Monday = +1
- [x] Fixed holidays present with correct names
- [x] Mother's/Father's Day land on 1st Sunday of May/June
- [x] `IsHoliday` positive + negative cases
- [x] No duplicate dates / expected count

### calendar_test.go
- [x] Month grid exact-string layout (color off)
- [x] Monday alignment: 1st on a Sunday → last column
- [x] Holiday styles: `list` legend text, `color` escape wrapping, `none` = plain
- [x] `-n` spans year boundary (Nov +4 → Nov, Dec, Jan, Feb)
- [x] `-3` / `-y` blocks and headers present
- [x] Color on/off wrapper behavior

### ParseArgs
- [x] `-n 6`, `-n 6 2026`, `-3`, `-y`, `--holidays`, `--holiday-style`, `--no-color`
- [x] invalid: `13 2026`, `abc`, `-n 0`, `-n -1`, `-n foo`

## Verification
- [x] `go test ./... -v`
- [x] `go vet ./...`
- [x] `gofmt -l .` clean
