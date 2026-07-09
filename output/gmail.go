package output

import (
	"fmt"
	"html"
	"html/template"
	"net/smtp"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/aymerick/douceur/inliner"
	"github.com/jordan-wright/email"
	"github.com/mattn/go-runewidth"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/entry"
)

type Gmail struct {
	email       string
	appPassword string
}

func GmailFlusher(c config.Config) (Flusher, error) {
	return &Gmail{
		email:       c.GmailAddress.Address,
		appPassword: c.GmailAppPassword,
	}, nil
}

func (p Gmail) Flush(timeline entry.Timeline) error {
	body, err := p.html(timeline)
	if err != nil {
		return err
	}

	atIndex := strings.LastIndexByte(p.email, '@')
	username := p.email[:atIndex]
	domain := p.email[atIndex+1:]

	addr := fmt.Sprintf("%s+%s@%s", username, timeline.Source, domain)
	msg := &email.Email{
		To:   []string{addr},
		From: fmt.Sprintf("%s <%s>", timeline.Source, addr),
		Headers: textproto.MIMEHeader{
			"References": []string{fmt.Sprintf("<%s+plago-%s%s@%s>", username, timeline.Subject, timeline.RefID, domain)},
		},
		Subject: timeline.Subject,
		HTML:    []byte(body),
	}

	if os.Getenv("DEBUG") != "" {
		fmt.Printf("--\n%#v\n", msg)
		fmt.Println(body)
		return nil
	}

	err = msg.Send(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", username+"@"+domain, p.appPassword, "smtp.gmail.com"),
	)
	if err != nil {
		return fmt.Errorf("mail send error error: %w", err)
	}

	return nil
}

func (p Gmail) html(timeline entry.Timeline) (string, error) {
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
{{- range .Entries }}

	{{- if ne $section .Section }}
	{{ if .Section }}<h2>{{ .Section }}</h2>{{ end }}
	{{ end }}
	{{- $section = .Section }}

	{{- if ne $channel .Channel }}
	{{ if .Channel }}<h3>{{ .Channel }}</h3>{{ end }}
	{{ end }}
	{{- $channel = .Channel }}

	<div>
		{{- $lead := .UserName }}{{ if not .UserName }}{{ $lead = .Timestamp.Format "15:04" }}{{ end }}
		{{ if .URL }}<a href="{{ .URL }}">{{ $lead | max 18 }}</a> &nbsp;
		{{- else }}{{ $lead }} &nbsp;
		{{- end }}

		{{- if .Reply }}» {{ end }}{{ .Text | compact | nl2br }}

		{{- with .Images }}<br>
		{{ range . }}<img src="{{ . | safe }}">{{ end }}
		{{- end }}

		{{- with .Attachments }}<br>
		{{ range . }}
			<blockquote>{{ .Text | compact | max 400 | nl2br }}
				{{- with .Images }}<br>
				{{ range . }}<img src="{{ . | safe }}">{{ end }}
				{{ end }}
			</blockquote>
		{{- end }}
		{{- end }}
	</div>

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
				return runewidth.Truncate(text, max, "…")
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
