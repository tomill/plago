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
	Input   string    `arg:"--in,required"          help:"Fetcher module (dummy, bluesky, discord, feed, twitter, twlist, slack and stdin)"`
	Output  string    `arg:"--out"   default:"json" help:"Flusher module (json, gmail)"`
	Hours   int       `arg:"--hours" default:"1"    placeholder:"1"   help:"Fetch entries from the previous N hours. Shortcut for --since and --until"`
	Since   time.Time `arg:"--since"                placeholder:"\"2026-07-25T12:00:00+09:00\""`
	Until   time.Time `arg:"--until"                placeholder:"\"2026-07-25T13:00:00+09:00\""`
	Subject string    `arg:"--subject"              help:"Used when --out gmail  Default is --since in YYYY-MM-DD format"`
	RefID   string    `arg:"--refid"                help:"Used when --out gmail  Additional References keys besides --subject"`
}

func (c ExecParams) String() string {
	return fmt.Sprintf("--in %s --out %s --since '%s' --until '%s'%s",
		c.Input, c.Output,
		c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339),
		lo.Ternary(c.RefID != "", " --refid "+c.RefID, ""),
	)
}

type List []string

func (l *List) UnmarshalText(b []byte) error {
	for v := range strings.SplitSeq(string(b), "\n") {
		v, _, _ = strings.Cut(v, "#")
		*l = append(*l, strings.FieldsFunc(v, func(r rune) bool {
			return unicode.IsSpace(r) || r == ','
		})...)
	}

	return nil
}

type Config struct {
	ExecParams
	LogFormat             string       `arg:"--,env:LOG_FORMAT"       help:"Log format (text, json)"              default:"text"`
	LogLevel              slog.Level   `arg:"--,env:LOG_LEVEL"        help:"Log level (info, debug, warn, error)" default:"info"`
	BlueskyHandle         string       `arg:"--,env:BLUESKY_HANDLE"   help:"Used when --in bluesky"`
	BlueskyAppPassword    string       `arg:"--,env:BLUESKY_APPPASS"  help:"Used when --in bluesky  See https://bsky.app/settings/app-passwords"`
	DiscordToken          string       `arg:"--,env:DISCORD_TOKEN"    help:"Used when --in discord"`
	DiscordChannels       List         `arg:"--,env:DISCORD_CHANNELS" help:"Used when --in discord  A newline/space/comma separated list (after # in line is ignored)"`
	FeedReaderAPI         string       `arg:"--,env:FEEDREADER_API"   help:"Used when --in feed     Google Reader compatible API" default:"https://theoldreader.com/"`
	FeedReaderToken       string       `arg:"--,env:FEEDREADER_TOKEN" help:"Used when --in feed     See https://github.com/theoldreader/api"`
	TwitterConsumerKey    string       `arg:"--,env:TWITTER_CONSUMER_KEY"        help:"Used when --in twitter/twlist"`
	TwitterConsumerSecret string       `arg:"--,env:TWITTER_CONSUMER_SECRET"     help:"Used when --in twitter/twlist"`
	TwitterToken          string       `arg:"--,env:TWITTER_OAUTH1_TOKEN"        help:"Used when --in twitter/twlist"`
	TwitterTokenSecret    string       `arg:"--,env:TWITTER_OAUTH1_TOKEN_SECRET" help:"Used when --in twitter/twlist"`
	TwitterUserID         string       `arg:"--,env:TWITTER_USERID"   help:"Used when --in twitter  Account ID not @username"`
	TwitterListIDs        List         `arg:"--,env:TWITTER_LISTS"    help:"Used when --in twlist   Same format as DISCORD_CHANNELS"`
	SlackToken            string       `arg:"--,env:SLACK_TOKEN"      help:"Used when --in slack    Multiple workspaces are not supported"`
	SlackWorkspace        string       `arg:"--,env:SLACK_WORKSPACE"  help:"Used when --in slack"`
	SlackChannelIDs       List         `arg:"--,env:SLACK_CHANNELS"   help:"Used when --in slack    Same format as DISCORD_CHANNELS"`
	GmailAddress          mail.Address `arg:"--,env:GMAIL_ADDRESS"    help:"Used when --out gmail   Gmail or Google Workspace email address"`
	GmailAppPassword      string       `arg:"--,env:GMAIL_APPPASS"    help:"Used when --out gmail   That https://myaccount.google.com/apppasswords"`
	GmailTemplateFile     string       `arg:"--,env:GMAIL_TEMPLATE"   help:"Used when --out gmail   Template file path to use instead of the default"`
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
	if strings.ToLower(c.LogFormat) == "json" {
		logger = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: c.LogLevel})
	} else {
		logger = console.NewHandler(os.Stderr, &console.HandlerOptions{Level: c.LogLevel})
	}
	slog.SetDefault(slog.New(logger))

	return c
}

func (c Config) Version() string {
	return "Plago build:" + version
}
