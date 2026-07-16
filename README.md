# mcal

A GNU `cal` alternative for the Lithuanian calendar. It renders Monday-first
month, multi-month and year views in English and marks Lithuanian public
holidays (non-work days). Written in Go using only the standard library.

## Installation

### go install (recommended)

Install directly without cloning the repository:

```sh
go install github.com/mantas6/mcal@latest
```

The `mcal` binary is placed in `$GOBIN`, or `$HOME/go/bin` if `GOBIN` is not
set. Make sure that directory is on your `PATH`.

### Build from source

```sh
git clone https://github.com/mantas6/mcal.git
cd mcal
go build -o mcal .
```

## Usage

```
mcal [options] [[month] year]
```

### Arguments

| Argument     | Description                              |
| ------------ | ---------------------------------------- |
| `month year` | show the given month (1-12) of the year  |
| `year`       | show the full given year                 |

With no arguments, the current month is shown.

### Options

| Option                | Description                                                   |
| --------------------- | ------------------------------------------------------------- |
| `-3`                  | previous, current and next month                              |
| `-y`, `--year`        | show the whole year                                           |
| `-n`, `--months N`    | show N months starting from the resolved date                 |
| `--holidays [year]`   | print the dated holiday list for the year (default: current)  |
| `--holiday-style S`   | holiday marking: `both`\|`color`\|`list`\|`none` (default: `both`) |
| `--no-color`          | disable ANSI color output                                     |
| `-h`, `--help`        | show help                                                     |

View precedence when combined: `-y` > `-n` > `-3` > single month.

### Holiday styles

The `--holiday-style` option controls how holidays are marked:

- `both` (default) — color the holiday dates in the grid and print a dated list.
- `color` — only color the holiday dates in the grid.
- `list` — only print the dated list below the grid.
- `none` — no holiday marking.

### Color detection

ANSI color is emitted only when writing to a terminal. It is automatically
disabled when stdout is not a TTY (e.g. piped or redirected), when the
`NO_COLOR` environment variable is set, or when `--no-color` is passed.

### Examples

Show a single month:

```
$ mcal 7 2026 --no-color
      July 2026     
Mo Tu We Th Fr Sa Su
       1  2  3  4  5
 6  7  8  9 10 11 12
13 14 15 16 17 18 19
20 21 22 23 24 25 26
27 28 29 30 31      
                    

Holidays:
2026-07-06  Statehood Day (King Mindaugas's Coronation Day)
```

Show previous, current and next month:

```
$ mcal -3 7 2026 --no-color
      June 2026             July 2026            August 2026    
Mo Tu We Th Fr Sa Su  Mo Tu We Th Fr Sa Su  Mo Tu We Th Fr Sa Su
 1  2  3  4  5  6  7         1  2  3  4  5                  1  2
 8  9 10 11 12 13 14   6  7  8  9 10 11 12   3  4  5  6  7  8  9
15 16 17 18 19 20 21  13 14 15 16 17 18 19  10 11 12 13 14 15 16
22 23 24 25 26 27 28  20 21 22 23 24 25 26  17 18 19 20 21 22 23
29 30                 27 28 29 30 31        24 25 26 27 28 29 30
                                            31                  

Holidays:
2026-06-07  Father's Day
2026-06-24  Midsummer Day (Joninės)
2026-07-06  Statehood Day (King Mindaugas's Coronation Day)
2026-08-15  Assumption Day (Žolinė)
```

List the holidays for a year:

```
$ mcal --holidays 2026
Lithuanian holidays 2026:
2026-01-01  New Year's Day
2026-02-16  Day of Restoration of the State of Lithuania
2026-03-11  Day of Restoration of Independence
2026-04-05  Easter Sunday
2026-04-06  Easter Monday
2026-05-01  International Labour Day
2026-05-03  Mother's Day
2026-06-07  Father's Day
2026-06-24  Midsummer Day (Joninės)
2026-07-06  Statehood Day (King Mindaugas's Coronation Day)
2026-08-15  Assumption Day (Žolinė)
2026-11-01  All Saints' Day
2026-11-02  All Souls' Day (Vėlinės)
2026-12-24  Christmas Eve (Kūčios)
2026-12-25  Christmas Day
2026-12-26  Second Day of Christmas
```

## Supported holidays

### Fixed dates

| Date   | Holiday                                             |
| ------ | --------------------------------------------------- |
| Jan 1  | New Year's Day                                      |
| Feb 16 | Day of Restoration of the State of Lithuania        |
| Mar 11 | Day of Restoration of Independence                  |
| May 1  | International Labour Day                             |
| Jun 24 | Midsummer Day (Joninės)                             |
| Jul 6  | Statehood Day (King Mindaugas's Coronation Day)     |
| Aug 15 | Assumption Day (Žolinė)                             |
| Nov 1  | All Saints' Day                                     |
| Nov 2  | All Souls' Day (Vėlinės)                            |
| Dec 24 | Christmas Eve (Kūčios)                              |
| Dec 25 | Christmas Day                                        |
| Dec 26 | Second Day of Christmas                             |

### Movable dates

| Holiday       | Rule                                                        |
| ------------- | ----------------------------------------------------------- |
| Easter Sunday | Meeus/Jones/Butcher (anonymous Gregorian) algorithm         |
| Easter Monday | Easter Sunday + 1 day                                       |
| Mother's Day  | first Sunday of May                                         |
| Father's Day  | first Sunday of June                                        |
