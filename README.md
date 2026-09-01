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
