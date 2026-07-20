package config

import (
	"fmt"
	"log/slog"
	"net/mail"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/alexflint/go-arg"
	"github.com/phsym/console-slog"
	"github.com/samber/lo"
)

var (
	version = "(development)"
)

type ExecParams struct {
	Input     string     `arg:"--in,required"`
	Output    string     `arg:"--out" default:"json"`
	Hours     int        `arg:"--hours" default:"1"`
	Since     time.Time  `arg:"--since"`
	Until     time.Time  `arg:"--until"`
	Subject   string     `arg:"--subject"`
	RefID     string     `arg:"--refid"`
	LogFormat string     `arg:"env:LOG_FORMAT" default:"text"`
	LogLevel  slog.Level `arg:"env:LOG_LEVEL" default:"info"`
}

func (c ExecParams) String() string {
	return fmt.Sprintf("--in %s --out %s --since '%s' --until '%s' %s",
		c.Input, c.Output,
		c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339),
		lo.Ternary(c.RefID != "", "--refid "+c.RefID, ""),
	)
}

type List []string

type Config struct {
	ExecParams
	GmailAddress          mail.Address `arg:"env:GMAIL_ADDRESS"`
	GmailAppPassword      string       `arg:"env:GMAIL_APPPASS"`
	GmailTemplateFile     string       `arg:"env:GMAIL_TEMPLATE"`
	SheetID               string       `arg:"env:SHEET_ID"`
	SheetCredentials      string       `arg:"env:SHEET_CREDENTIALS"`
	SlackToken            string       `arg:"env:SLACK_TOKEN"`
	SlackWorkspace        string       `arg:"env:SLACK_WORKSPACE"`
	SlackChannels         List         `arg:"env:SLACK_CHANNELS"`
	BlueskyAppKey         string       `arg:"env:BLUESKY_APPKEY"`
	BlueskyHandle         string       `arg:"env:BLUESKY_HANDLE"`
	DiscordToken          string       `arg:"env:DISCORD_TOKEN"`
	DiscordChannels       List         `arg:"env:DISCORD_CHANNELS"`
	TwitterConsumerKey    string       `arg:"env:TWITTER_CONSUMER_KEY"`
	TwitterConsumerSecret string       `arg:"env:TWITTER_CONSUMER_SECRET"`
	TwitterToken          string       `arg:"env:TWITTER_OAUTH1_TOKEN"`
	TwitterTokenSecret    string       `arg:"env:TWITTER_OAUTH1_TOKEN_SECRET"`
	TwitterUserID         string       `arg:"env:TWITTER_USERID"`
	TwitterLists          List         `arg:"env:TWITTER_LISTS"`
	FeedReaderAPI         string       `arg:"env:FEEDREADER_API" default:"https://theoldreader.com/"`
	FeedReaderToken       string       `arg:"env:FEEDREADER_TOKEN"`
}

func GetOptions() Config {
	var c Config
	arg.MustParse(&c)

	if c.Until.IsZero() {
		c.Until = time.Now().Truncate(time.Hour)
	}
	if c.Since.IsZero() {
		c.Since = c.Until.Add(-time.Duration(c.Hours) * time.Hour).Truncate(time.Hour)
	}

	if c.Subject == "" {
		c.Subject = c.Since.Format(time.DateOnly)
	}

	var logger slog.Handler
	if c.LogFormat == "json" {
		logger = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: c.LogLevel})
	} else {
		logger = console.NewHandler(os.Stderr, &console.HandlerOptions{Level: c.LogLevel})
	}
	slog.SetDefault(slog.New(logger))

	return c
}

func (c Config) Version() string {
	return "plago " + version
}

func (l *List) UnmarshalText(b []byte) error {
	for v := range strings.SplitSeq(string(b), "\n") {
		v, _, _ = strings.Cut(v, "#")
		*l = append(*l, strings.FieldsFunc(v, func(r rune) bool {
			return unicode.IsSpace(r) || r == ','
		})...)
	}

	return nil
}
