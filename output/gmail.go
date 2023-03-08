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
  margin: 1rem 0; 
}

p {
  margin: 0;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  margin: 0 0 0.4em;
}

li img {
  height: 80px;
  max-width: 200px;
  margin: 5px 10px 0 0;
}

li blockquote {
  color: gray;
  border-left: 2px solid silver;
  margin: 3px 0 0 0;
  padding: 1px 10px;
}
</style>
{{- $started := false }}
{{- $section := "" }}
{{- range .Messages }}

{{- if ne $section .Section }}{{ $section = .Section }}
{{ if $started }}</ul>{{ end }}

<h2>{{ $section }}</h2>
<ul>{{ $started = true }}

{{- else if not $started }}
<ul>{{ $started = true }}
{{- end }}

  <li><div>{{ if .Permalink }}<a href="{{ .Permalink }}">{{ .Lead }}</a>&nbsp;{{ end }}
    {{ .Text | compact | chomp | nl2br }}
  {{- with .Attachments }}<br>
    {{ range . }}
      {{- if eq .Type "image" }}<img src="{{ .Permalink | safe }}">
      {{- else }}
    <blockquote>{{ .Text | compact | max 800 | chomp | nl2br }}</blockquote>{{ end }}
    {{- end }}
  {{- end }}</div></li>

{{- end }}
</ul>`

	tmpl, err := template.New("body").Funcs(template.FuncMap{
		"chomp": func(text string) string {
			return strings.TrimRightFunc(text, func(c rune) bool {
				return unicode.IsSpace(c) || c == '\r' || c == '\n'
			})
		},
		"compact": func(text string) string {
			return regexp.MustCompile(`\s*\n\s*\n`).ReplaceAllString(text, "\n")
		},
		"nl2br": func(text string) template.HTML {
			text = html.UnescapeString(text)
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(text), "\n", "<br>\n"))
		},
		"max": func(max int, text string) string {
			if s := []rune(text); len(s) > max {
				return string(s[:max]) + "..."
			} else {
				return text
			}
		},
		"quoted": func(mark string, text string) string {
			if text != "" {
				text = mark + text
			}

			return strings.ReplaceAll(text, "\n", "  \n"+mark)
		},
		"safe": func(s string) template.URL {
			return template.URL(s)
		},
	}).Parse(body)
	if err != nil {
		return "", err
	}

	var buff strings.Builder
	if err := tmpl.Execute(&buff, timeline); err != nil {
		return "", err
	}

	return inliner.Inline(buff.String())
}
