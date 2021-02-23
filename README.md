**This repo has now been archived as Runalyze now has [native support for syncing with Oura](https://blog.runalyze.com/features/sync-with-oura/).**

# Oura to Runalyze

![Tests Status Badge](https://github.com/lildude/oura-to-runalyze/workflows/Tests/badge.svg)

Sync health data from Oura to Runalyze.

## Installation

Download a binary from the [releases](https://github.com/lildude/oura-to-runalyze/releases) page, or run `go get github.com/lildude/oura-to-runalyze`.

## Usage

Obtain an access token from both Oura and Runalyze and set `OURA_ACCESS_TOKEN` and `RUNALYZE_ACCESS_TOKEN` environment variables respectively.

```console
$ export OURA_ACCESS_TOKEN="secret1" 
$ export RUNALYZE_ACCESS_TOKEN="secret2"
```

You can also set these via a `.env` file in the current directory.

You're now ready to start synchronising.

```
Usage:
  -start string
        Start date in the format: YYYY-MM-DD. If not provided, defaults to Oura's default of one week ago.
  -end string
        End date in the form: YYYY-MM-DD. If not provided, defaults to Oura's default of today.
  -yesterday
        Use yesterday's date as the start date.
  -version
        Print the version and exit.
```

Sync the last week using Oura's defaults:

```console
$ oura-to-runalyze
Successfully sync'd to Runalyze
$
```

Sync just last night's sleep:

```console
$ oura-to-runalyze -yesterday
Successfully sync'd to Runalyze
$
```

Sync a range of dates:

```console
$ oura-to-runalyze -start 2020-01-01 -end 2020-03-31
Successfully sync'd to Runalyze
$
```
