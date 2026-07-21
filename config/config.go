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
	LogFormat string     `arg:"env:LOG_FORMAT" default:"text" placeholder:"text|json"`
	LogLevel  slog.Level `arg:"env:LOG_LEVEL"  default:"info" placeholder:"info|debug|warn|error"`
	Input     string     `arg:"--in,required"           placeholder:"fetcher" help:"(required)"`
	Output    string     `arg:"--out"   default:"json"  placeholder:"flusher" help:""`
	Hours     int        `arg:"--hours" default:"1"     help:"Fetch entries from the previous N hours. Use --since/until to specify a time range"`
	Since     time.Time  `arg:"--since"                 placeholder:""  help:"ISO8601 format"`
	Until     time.Time  `arg:"--until"                 placeholder:""  help:"ISO8601 format"`
	Subject   string     `arg:"--subject"               placeholder:""  help:"for --out gmail: Default: \"since\" in YYYY-MM-DD format"`
	RefID     string     `arg:"--refid"                 placeholder:""  help:"for --out gmail: Additional keys to add to References besides Subject"`
}

func (c ExecParams) String() string {
	return fmt.Sprintf("--in %s --out %s --since '%s' --until '%s'%s",
		c.Input, c.Output,
		c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339),
		lo.Ternary(c.RefID != "", " --refid "+c.RefID, ""),
	)
}

type List []string

type Config struct {
	ExecParams
	GmailAddress          mail.Address `placeholder:"" arg:"env:GMAIL_ADDRESS"    help:"for --out gmail"`
	GmailAppPassword      string       `placeholder:"" arg:"env:GMAIL_APPPASS"    help:"for --out gmail"`
	GmailTemplateFile     string       `placeholder:"" arg:"env:GMAIL_TEMPLATE"   help:"for --out gmail (optional)"`
	BlueskyAppKey         string       `placeholder:"" arg:"env:BLUESKY_APPKEY"   help:"for --out bluesky"`
	BlueskyHandle         string       `placeholder:"" arg:"env:BLUESKY_HANDLE"   help:"for --out bluesky"`
	DiscordToken          string       `placeholder:"" arg:"env:DISCORD_TOKEN"    help:"for --out discord"`
	DiscordChannels       List         `placeholder:"" arg:"env:DISCORD_CHANNELS" help:"for --out discord: A newline/space/comma separated list. After # in line is ignored"`
	FeedReaderAPI         string       `placeholder:"" arg:"env:FEEDREADER_API"   help:"for --out feedreader: GoogleReader compatible API" default:"https://theoldreader.com/"`
	FeedReaderToken       string       `placeholder:"" arg:"env:FEEDREADER_TOKEN" help:"for --out feedreader: See https://github.com/theoldreader/api"`
	TwitterConsumerKey    string       `placeholder:"" arg:"env:TWITTER_CONSUMER_KEY"        help:"for --out twitter"`
	TwitterConsumerSecret string       `placeholder:"" arg:"env:TWITTER_CONSUMER_SECRET"     help:"for --out twitter"`
	TwitterToken          string       `placeholder:"" arg:"env:TWITTER_OAUTH1_TOKEN"        help:"for --out twitter"`
	TwitterTokenSecret    string       `placeholder:"" arg:"env:TWITTER_OAUTH1_TOKEN_SECRET" help:"for --out twitter"`
	TwitterUserID         string       `placeholder:"" arg:"env:TWITTER_USERID"   help:"for --out twitter: Account ID not the @username"`
	TwitterListIDs        List         `placeholder:"" arg:"env:TWITTER_LISTS"    help:"for --out twilist (Same format as DISCORD_CHANNELS)"`
	SlackToken            string       `placeholder:"" arg:"env:SLACK_TOKEN"      help:"for --out slack"`
	SlackWorkspace        string       `placeholder:"" arg:"env:SLACK_WORKSPACE"  help:"for --out slack"`
	SlackChannelIDs       List         `placeholder:"" arg:"env:SLACK_CHANNELS"   help:"for --out slack (Same format as DISCORD_CHANNELS)"`
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
