package config

import (
	"fmt"
	"log/slog"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/phsym/console-slog"
)

var version = "(development)"

type ExecParams struct {
	Input   string    `arg:"--in,required"            help:"Fetcher module (dummy, url, bluesky, discord, feed, twitter, twlist, slack, youtube and stdin). Set as timeline.Source value"`
	Output  string    `arg:"--out"    default:"json"  help:"Flusher module (json, gmail)"`
	URL     *url.URL  `arg:"--url"                    help:"URL to fetch when --in url"`
	Filter  string    `arg:"--filter"                 help:"Set the API endpoint used to filter entries before output"`
	Hours   int       `arg:"--hours"  default:"1"     help:"Fetch entries from the previous N hours. Shortcut for --since and --until"`
	Since   time.Time `arg:"--since"  placeholder:"\"2026-07-25T12:00:00+09:00\""`
	Until   time.Time `arg:"--until"  placeholder:"\"2026-07-25T13:00:00+09:00\""`
	Subject string    `arg:"--subject"                help:"Set as timeline.Subject and Used with --out gmail. Defaults to --since in YYYY-MM-DD format"`
	RefID   string    `arg:"--refid"                  help:"Set as timeline.RefID and Used when --out gmail. Additional References keys besides --subject"`
}

type Config struct {
	ExecParams
	LogFormat             string         `arg:"--,env:LOG_FORMAT"              help:"Log format (text, json)"               default:"text"`
	LogLevel              slog.Level     `arg:"--,env:LOG_LEVEL"               help:"Log level (info, debug, warn, error)"  default:"info"`
	BlueskyHandle         string         `arg:"--,env:BLUESKY_HANDLE"          help:"Used when --in bluesky"`
	BlueskyAppPassword    string         `arg:"--,env:BLUESKY_APPPASS"         help:"Used when --in bluesky  See https://bsky.app/settings/app-passwords"`
	DiscordToken          string         `arg:"--,env:DISCORD_TOKEN"           help:"Used when --in discord"`
	DiscordChannelIDs     List           `arg:"--,env:DISCORD_CHANNELS"        help:"Used when --in discord  Newline/space/comma-separated list; # comments are ignored"`
	DiscordChannels       Sheet[Channel] `arg:"--,env:DISCORD_CHANNELS_SHEET"  help:"Used when --in discord  Spreadsheet URL containing the values. Used when DiscordChannelIDs is empty"`
	FeedReaderAPI         string         `arg:"--,env:FEEDREADER_API"          help:"Used when --in feed     Google Reader-compatible API"  default:"https://theoldreader.com/"`
	FeedReaderToken       string         `arg:"--,env:FEEDREADER_TOKEN"        help:"Used when --in feed     See https://github.com/theoldreader/api"`
	TwitterConsumerKey    string         `arg:"--,env:TWITTER_CONSUMER_KEY"    help:"Used when --in twitter/twlist  The X API requires a paid subscription"`
	TwitterConsumerSecret string         `arg:"--,env:TWITTER_CONSUMER_SECRET" help:"Used when --in twitter/twlist"`
	TwitterToken          string         `arg:"--,env:TWITTER_OAUTH1_TOKEN"    help:"Used when --in twitter/twlist"`
	TwitterTokenSecret    string         `arg:"--,env:TWITTER_OAUTH1_TOKEN_SECRET"  help:"Used when --in twitter/twlist"`
	TwitterUserID         string         `arg:"--,env:TWITTER_USERID"          help:"Used when --in twitter  Account ID, not @username"`
	TwitterListIDs        List           `arg:"--,env:TWITTER_LISTS"           help:"Used when --in twlist   Same format as DISCORD_CHANNELS"`
	SlackToken            string         `arg:"--,env:SLACK_TOKEN"             help:"Used when --in slack    Multiple workspaces are not supported"`
	SlackWorkspace        string         `arg:"--,env:SLACK_WORKSPACE"         help:"Used when --in slack    i.e., the xxx part of xxx.slack.com"`
	SlackChannelIDs       List           `arg:"--,env:SLACK_CHANNELS"          help:"Used when --in slack    Same format as DISCORD_CHANNELS"`
	SlackChannels         Sheet[Channel] `arg:"--,env:SLACK_CHANNELS_SHEET"    help:"Used when --in slack    Spreadsheet URL containing the values. Used when SlackChannelIDs is empty"`
	YoutubeAPIKey         string         `arg:"--,env:YOUTUBE_API_KEY"         help:"Used when --in youtube"`
	YoutubeChannelIDs     List           `arg:"--,env:YOUTUBE_CHANNELS"        help:"Used when --in youtube  Same format as DISCORD_CHANNELS"`
	YoutubeChannels       Sheet[Channel] `arg:"--,env:YOUTUBE_CHANNELS_SHEET"  help:"Used when --in youtube  Spreadsheet URL containing the values. Used when YoutubeChannelIDs is empty"`
	GmailAddress          mail.Address   `arg:"--,env:GMAIL_ADDRESS"           help:"Used when --out gmail   Gmail or Google Workspace email address"`
	GmailAppPassword      string         `arg:"--,env:GMAIL_APPPASS"           help:"Used when --out gmail   See https://myaccount.google.com/apppasswords"`
	GmailTemplateFile     string         `arg:"--,env:GMAIL_TEMPLATE"          help:"Used when --out gmail   Template file path to use instead of the default"`
	SheetCredentials      string         `arg:"--,env:SHEET_CREDENTIALS"       help:"Used with *_SHEET env   Sheets API Service Account credentials JSON"`
}

type Channel struct {
	ChannelID   string
	ChannelName string
}

func (c ExecParams) String() string {
	s := fmt.Sprintf("--in %s --out %s --since '%s' --until '%s'",
		c.Input, c.Output,
		c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339),
	)
	if c.Subject != "" {
		s += fmt.Sprintf(" --subject '%s'", c.Subject)
	}
	if c.RefID != "" {
		s += fmt.Sprintf(" --refid %s", c.RefID)
	}
	if c.Filter != "" {
		s += fmt.Sprintf(" --filter '%s'", c.Filter)
	}

	return s
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
