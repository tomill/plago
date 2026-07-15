package output

import (
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/smtp"
	"net/textproto"
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
	address  string
	password string
}

func GmailFlusher(c config.Config) (Flusher, error) {
	return &Gmail{
		address:  c.GmailAddress.Address,
		password: c.GmailAppPassword,
	}, nil
}

func (p Gmail) Flush(timeline entry.Timeline) error {
	if len(timeline.Entries) == 0 {
		return nil
	}

	body, err := p.body(timeline)
	if err != nil {
		return err
	}

	username, domain, _ := strings.Cut(p.address, "@")
	to := fmt.Sprintf("%s+%s@%s", username, timeline.Source, domain)
	msg := &email.Email{
		To:   []string{to},
		From: fmt.Sprintf("%s <%s>", timeline.Source, to),
		Headers: textproto.MIMEHeader{
			"References": []string{
				fmt.Sprintf("<%s+plago-%s%s@%s>", username, timeline.Subject, timeline.RefID, domain),
			},
		},
		Subject: timeline.Subject,
		HTML:    []byte(body),
	}

	slog.Debug("", "body", body)

	return msg.Send(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", p.address, p.password, "smtp.gmail.com"),
	)
}

var templateBody = `
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
		{{- $lead := .User }}{{ if not .User }}{{ $lead = .Timestamp.Format "15:04" }}{{ end }}
		{{ if .URL }}<a href="{{ .URL }}">{{ $lead | max 18 }}</a> &nbsp;
		{{- else }}{{ $lead }} &nbsp;
		{{- end }}

		{{- if .Reply }}» {{ end }}{{ .Text | compact | nl2br }}

		{{- with .Images }}<br>
		{{ range . }}<img src="{{ . | safe }}">{{ end }}
		{{- end }}

		{{- with .Attachments }}<br>
		{{ range . }}
			<blockquote>{{ .Text | compact | max 500 | nl2br }}
				{{- with .Images }}<br>
				{{ range . }}<img src="{{ . | safe }}">{{ end }}
				{{ end }}
			</blockquote>
		{{- end }}
		{{- end }}
	</div>

{{- end }}
`

var (
	reEmptyLines = regexp.MustCompile(`[\p{Z}\s]*\n[\p{Z}\s]*\n`)
)

func (p Gmail) body(timeline entry.Timeline) (string, error) {
	tmpl := template.New("body").Funcs(template.FuncMap{
		"max": func(max int, text string) string {
			return runewidth.Truncate(text, max, "…")
		},
		"compact": func(text string) string {
			text = reEmptyLines.ReplaceAllString(text, "\n")
			text = strings.TrimRightFunc(text, unicode.IsSpace)
			return text
		},
		"nl2br": func(text string) template.HTML {
			text = html.UnescapeString(text)
			text = template.HTMLEscapeString(text)
			text = strings.ReplaceAll(text, "\n", "<br>\n")
			text = strings.ReplaceAll(text, "\t", "&nbsp;&nbsp;")
			return template.HTML(text)
		},
		"safe": func(s string) template.URL {
			return template.URL(s)
		},
	})

	var buff strings.Builder
	err := template.Must(tmpl.Parse(templateBody)).Execute(&buff, timeline)
	if err != nil {
		return "", err
	}

	return inliner.Inline(buff.String())
}
