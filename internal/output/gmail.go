package output

import (
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/jordan-wright/email"
	"github.com/mattn/go-runewidth"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
	"github.com/vanng822/go-premailer/premailer"
)

type Gmail struct {
	address      string
	password     string
	templateFile string
}

func GmailFlusher(c config.Config) (Flusher, error) {
	if c.GmailAddress.Address == "" || c.GmailAppPassword == "" {
		return nil, fmt.Errorf("gmail address and password are required")
	}
	return &Gmail{
		address:      c.GmailAddress.Address,
		password:     c.GmailAppPassword,
		templateFile: c.GmailTemplateFile,
	}, nil
}

func (p *Gmail) Flush(timeline plago.Timeline) error {
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

var (
	reEmptyLines = regexp.MustCompile(`[\p{Z}\s]*\n[\p{Z}\s]*\n`)
	reAmazon     = regexp.MustCompile(`https?://(www\.amazon\.[a-z.]{2,6})/(?:[^\s]*?/)?(?:dp|gp/product|exec/obidos)/([A-Z0-9]{10})(?:[/?][^\s]*)?`)
	reUtmTracker = regexp.MustCompile(`[?&]utm_[a-zA-Z0-9_]+=[^&]*`)
)

func (p *Gmail) body(timeline plago.Timeline) (string, error) {
	var buff strings.Builder
	if err := p.template().Execute(&buff, timeline); err != nil {
		return "", err
	}

	inliner, err := premailer.NewPremailerFromString(buff.String(), premailer.NewOptions(
		premailer.WithRemoveClasses(true),
	))
	if err != nil {
		return "", err
	}

	return inliner.Transform()
}

//go:embed gmail_template.tmpl
var defaultTemplate string

func (p *Gmail) template() *template.Template {
	funcs := template.FuncMap{
		"max": func(max int, text string) string {
			return runewidth.Truncate(text, max, "…")
		},
		"compact": func(text string) string {
			text = reEmptyLines.ReplaceAllString(text, "\n")
			text = strings.TrimRightFunc(text, unicode.IsSpace)

			text = reAmazon.ReplaceAllString(text, `https://$1/dp/$2/`)
			text = reUtmTracker.ReplaceAllString(text, "?") // lazy
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
	}
	if file := p.templateFile; file != "" {
		base := filepath.Base(file)
		return template.Must(template.New(base).Funcs(funcs).ParseFiles(p.templateFile))
	} else {
		return template.Must(template.New("body").Funcs(funcs).Parse(defaultTemplate))
	}
}
