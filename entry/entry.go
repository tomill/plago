package entry

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/tomill/centre/config"
)

type Timeline struct {
	Source  string  `json:"source"`
	Subject string  `json:"subject"`
	RefID   string  `json:"refid,omitempty"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	Section     string    `json:"section,omitempty"`
	Channel     string    `json:"channel,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitzero"`
	URL         string    `json:"url,omitempty"`
	User        string    `json:"user,omitempty"`
	Reply       bool      `json:"reply,omitempty"`
	Text        string    `json:"text,omitempty"`
	Images      []string  `json:"images,omitempty"`
	Attachments []*Entry  `json:"attachments,omitempty"`
}

func NewTimeline(c config.ExecParams) Timeline {
	return Timeline{
		Source:  c.Input,
		Subject: c.Since.Format(time.DateOnly),
		RefID:   c.RefID,
	}
}

func (t *Timeline) Append(entry *Entry) {
	if entry != nil {
		t.Entries = append(t.Entries, *entry)
	}
}

func (t *Timeline) AppendError(err error) {
	_, file, line, _ := runtime.Caller(1)

	t.Entries = append(t.Entries, Entry{
		Channel:     "Error",
		Timestamp:   time.Now(),
		Text:        err.Error(),
		Attachments: []*Entry{{Text: fmt.Sprintf("%s:%d", file, line)}},
	})
}

func (t *Timeline) Sorted() Timeline {
	sort.Slice(t.Entries, func(i, j int) bool {
		e1, e2 := t.Entries[i], t.Entries[j]

		if e1.Section != e2.Section {
			return e1.Section < e2.Section
		}
		if e1.Channel != e2.Channel {
			return e1.Channel < e2.Channel
		}
		if e1.Timestamp.Equal(e2.Timestamp) {
			return e1.URL < e2.URL
		}

		return e1.Timestamp.Before(e2.Timestamp)
	})

	return *t
}

func (e *Entry) AddImage(url string) {
	e.Images = append(e.Images, url)
}

func (e *Entry) AddAttachment(a Entry) *Entry {
	e.Attachments = append(e.Attachments, &a)
	return &a
}
