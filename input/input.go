package input

import (
	"time"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Fetcher interface {
	Fetch() (message.Timeline, error)
}

type Dummy struct {
	Timeline message.Timeline
}

func DummyFetcher(config.Config) (Fetcher, error) {
	var ts = time.Date(2022, 9, 1, 12, 02, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60))
	p := &Dummy{
		Timeline: message.Timeline{
			Source:  "dummy",
			Subject: "test",
			Messages: []message.Message{
				{
					Timestamp: ts,
					Permalink: "https://example.com/1",
					Lead:      "user1",
					Text:      "あああああああ改行\nあああああ",
				},
				{
					Section:   "main",
					Permalink: "https://example.com/2",
					Timestamp: ts.Add(2 * time.Minute),
					Lead:      "user2",
					Text:      "いいいいいいい<script>いいいいいいいいいいいいいい",
				},
				{
					Section:   "sub",
					Permalink: "https://example.com/3",
					Timestamp: ts.Add(1 * time.Minute),
					Lead:      "user3",
					Text:      "えええええ",
					Attachments: []message.Message{
						{
							Type:      message.TypeImage,
							Permalink: "https://www.gravatar.com/avatar/f5d789b9076fd42eaabee3b2941b74db?s=50",
						},
						{
							Type:      message.TypeImage,
							Permalink: "https://www.gravatar.com/avatar/f5d789b9076fd42eaabee3b2941b74db?s=50",
						},
						{
							Type: message.TypeText,
							Text: "引用\n引用",
						},
					},
				},
				{
					Section:   "sub",
					Timestamp: ts.Add(3 * time.Minute),
					Text:      "おお",
				},
				{
					Section:   "main",
					Permalink: "https://example.com/3",
					Lead:      "user2",
					Timestamp: ts.Add(3 * time.Minute),
					Text:      "ううう",
					Attachments: []message.Message{
						{
							Text: "ううううう",
						},
					},
				},
			},
		},
	}

	return p, nil
}

func (p Dummy) Fetch() (message.Timeline, error) {
	return p.Timeline.Sorted(), nil
}
