package output

import (
	"fmt"
	"html"
	"html/template"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strings"
	"unicode"

	"github.com/aymerick/douceur/inliner"
	"github.com/jordan-wright/email"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

var GmailDomain = "gmail.com"

type Gmail struct {
	account     string
	appPassword string
}

func GmailFlusher(c config.Config) Flusher {
	return &Gmail{
		account:     c.GmailAccount,
		appPassword: c.GmailAppPassword,
	}
}

func (p Gmail) Flush(timeline message.Timeline) error {
	body, err := p.html(timeline)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s+%s@%s", p.account, timeline.Source, GmailDomain)
	msg := &email.Email{
		To:   []string{addr},
		From: fmt.Sprintf("%s <%s>", timeline.Source, addr),
		Headers: textproto.MIMEHeader{
			"References": []string{fmt.Sprintf("<%s+%s-%s-centre@%s>", p.account, timeline.Subject, timeline.RefID, GmailDomain)},
		},
		Subject: timeline.Subject,
		HTML:    []byte(body),
	}

	err = msg.Send(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", p.account+"@"+GmailDomain, p.appPassword, "smtp.gmail.com"),
	)
	if err != nil {
		return fmt.Errorf("mail send error error: %w", err)
	}

	return nil
}

func (p Gmail) html(timeline message.Timeline) (string, error) {
	body := `
<style>
h2 {
  font-size: 1rem;
  color: gray;
}

div {
  margin: 0 0 0.4em;
  color: #222;
}

div img {
  height: 80px;
  max-width: 200px;
  margin: 5px 10px 0 0;
  border-radius: 3px;
}

div blockquote {
  color: gray;
  border-left: 2px solid silver;
  margin: 3px 0 0 0;
  padding: 1px .5rem;
}
</style>

{{- $section := "" }}
{{- $channel := "" }}
{{- range .Messages }}

{{- if ne $section .Section }}
{{ if .Section }}<h2>{{ .Section }}</h2>{{ end }}
{{ end }}
{{- $section = .Section }}

{{- if ne $channel .Channel }}
{{ if .Channel }}<h3>{{ .Channel }}</h3>{{ end }}
{{ end }}
{{- $channel = .Channel }}
{{- $lead := .UserName }}{{ if not .UserName }}{{ $lead = .Timestamp.Format "15:04" }}{{ end }}

<div>{{ if .Permalink }}<a href="{{ .Permalink }}" title="{{ .Timestamp.Format "2006-01-02 15:04:05" }}">{{ $lead }}</a> &nbsp;{{ else }}{{ $lead }} &nbsp;{{ end }}
{{- if .Reply }}» {{ end }}{{ .Text | compact | nl2br }}
{{- with .Attachments }}<br>
{{ range . }}
	{{- if eq .Type "image" }}<img src="{{ .Permalink | safe }}">
	{{- else }}
  <blockquote>{{ .Text | compact | max 800 | nl2br }}</blockquote>{{ end }}
{{- end }}
{{- end }}</div>
{{- end }}
`

	var buff strings.Builder
	err := template.Must(template.New("body").
		Funcs(template.FuncMap{
			"compact": func(text string) string {
				text = regexp.MustCompile(`\s*\n\s*\n`).ReplaceAllString(text, "\n")
				text = strings.TrimRightFunc(text, func(c rune) bool {
					return unicode.IsSpace(c) || c == '\r' || c == '\n'
				})
				return text
			},
			"max": func(max int, text string) string {
				if s := []rune(text); len(s) > max {
					return string(s[:max]) + "..."
				} else {
					return text
				}
			},
			"nl2br": func(text string) template.HTML {
				text = html.UnescapeString(text)
				text = template.HTMLEscapeString(text)
				text = regexp.MustCompile(`(?m)^(\s+)`).ReplaceAllStringFunc(text, func(s string) string {
					return strings.Repeat("&nbsp;", len(s))
				})
				text = strings.ReplaceAll(text, "\n", "<br>\n")
				return template.HTML(text)
			},
			"safe": func(s string) template.URL {
				return template.URL(s)
			},
		}).
		Parse(body)).
		Execute(&buff, timeline)
	if err != nil {
		return "", err
	}

	return inliner.Inline(buff.String())
}
