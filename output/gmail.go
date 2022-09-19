package output

import (
	"fmt"
	"net/smtp"
	"net/textproto"

	"github.com/aymerick/douceur/inliner"
	"github.com/jordan-wright/email"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

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

func (p Gmail) Flush(timeline *message.Timeline) error {
	css := `<style>
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
</style>`

	body, err := inliner.Inline(css + timeline.HTML())
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s+%s@gmail.com", p.account, timeline.Source)
	msg := &email.Email{
		To:   []string{addr},
		From: fmt.Sprintf("%s <%s>", timeline.Source, addr),
		Headers: textproto.MIMEHeader{
			"References": []string{fmt.Sprintf("<%s-%s>", timeline.Subject, addr)},
		},
		Subject: timeline.Subject,
		HTML:    []byte(body),
	}

	err = msg.Send(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", p.account+"@gmail.com", p.appPassword, "smtp.gmail.com"),
	)
	if err != nil {
		return err
	}

	return nil
}
