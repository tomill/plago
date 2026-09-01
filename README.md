# ♺ Plago


## Install

```bash
go install github.com/tomill/plago@latest
```

## Usage

```bash
# Fetch from hard-coded example entries and output as JSON for testing.
$ plago --in dummy --out json | jq .

# --in stdin accepts input piped from external programs.
$ echo '{"entries":[{"text":"foobar"}]}' | go run main.go --in stdin --out json | jq .
```

Plago uses environment variables for API access configuration. I recommend using direnv or doppler.

```bash
# Fetch posts from bluesky for the last 1 hour and output to Gmail with doppler.
$ doppler run -- plago --in bluesky --out gmail --hours 1
2026-07-25 17:19:34 INF plago --in bluesky --out gmail --since '2026-07-25T16:00:00+09:00' --until '2026-07-25T17:00:00+09:00'
2026-07-25 17:19:35 INF plago fetched 10 entries from bluesky. flushing them to gmail ...
```

The [lambda/](lambda/) directory contains an example of running Plago on AWS Lambda that I use. This handler is designed for Lambda Function URL and accepts JSON POST from external programs.

## Options
```
Options:
  --in IN                Fetcher module (dummy, url, bluesky, discord, feed, twitter, twlist, slack, youtube and stdin). Set as timeline.Source value
  --out OUT              Flusher module (json, gmail) [default: json]
  --url URL              URL to fetch when --in url
  --filter FILTER        Set the API endpoint used to filter entries before output
  --hours HOURS          Fetch entries from the previous N hours. Shortcut for --since and --until [default: 1]
  --since "2026-07-25T12:00:00+09:00"
  --until "2026-07-25T13:00:00+09:00"
  --subject SUBJECT      Set as timeline.Subject and Used with --out gmail. Defaults to --since in YYYY-MM-DD format
  --refid REFID          Set as timeline.RefID and Used when --out gmail. Additional References keys besides --subject
  --help, -h             display this help and exit
  --version              display version and exit

Environment variables:
  LOG_FORMAT             Optional. Log format (text, json) [default: text]
  LOG_LEVEL              Optional. Log level (info, debug, warn, error) [default: info]
  BLUESKY_HANDLE         Optional. Used when --in bluesky
  BLUESKY_APPPASS        Optional. Used when --in bluesky  See https://bsky.app/settings/app-passwords
  DISCORD_TOKEN          Optional. Used when --in discord
  DISCORD_CHANNELS       Optional. Used when --in discord  Newline/space/comma-separated list; # comments are ignored
  DISCORD_CHANNELS_SHEET
                         Optional. Used when --in discord  Spreadsheet URL containing the values. Used when DiscordChannelIDs is empty
  FEEDREADER_API         Optional. Used when --in feed     Google Reader-compatible API [default: https://theoldreader.com/]
  FEEDREADER_TOKEN       Optional. Used when --in feed     See https://github.com/theoldreader/api
  TWITTER_CONSUMER_KEY   Optional. Used when --in twitter/twlist  The X API requires a paid subscription
  TWITTER_CONSUMER_SECRET
                         Optional. Used when --in twitter/twlist
  TWITTER_OAUTH1_TOKEN   Optional. Used when --in twitter/twlist
  TWITTER_OAUTH1_TOKEN_SECRET
                         Optional. Used when --in twitter/twlist
  TWITTER_USERID         Optional. Used when --in twitter  Account ID, not @username
  TWITTER_LISTS          Optional. Used when --in twlist   Same format as DISCORD_CHANNELS
  SLACK_TOKEN            Optional. Used when --in slack    Multiple workspaces are not supported
  SLACK_WORKSPACE        Optional. Used when --in slack    i.e., the xxx part of xxx.slack.com
  SLACK_CHANNELS         Optional. Used when --in slack    Same format as DISCORD_CHANNELS
  SLACK_CHANNELS_SHEET   Optional. Used when --in slack    Spreadsheet URL containing the values. Used when SlackChannelIDs is empty
  YOUTUBE_API_KEY        Optional. Used when --in youtube
  YOUTUBE_CHANNELS       Optional. Used when --in youtube  Same format as DISCORD_CHANNELS
  YOUTUBE_CHANNELS_SHEET
                         Optional. Used when --in youtube  Spreadsheet URL containing the values. Used when YoutubeChannelIDs is empty
  GMAIL_ADDRESS          Optional. Used when --out gmail   Gmail or Google Workspace email address
  GMAIL_APPPASS          Optional. Used when --out gmail   See https://myaccount.google.com/apppasswords
  GMAIL_TEMPLATE         Optional. Used when --out gmail   Template file path to use instead of the default
  SHEET_CREDENTIALS      Optional. Used with *_SHEET env   Sheets API Service Account credentials JSON
