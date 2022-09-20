package input

import (
	"time"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Fetcher interface {
	Fetch() (*message.Timeline, error)
}

type Dummy struct {
	Timeline *message.Timeline
}

func DummyFetcher(config.Config) Fetcher {
	var ts = time.Date(2022, 9, 1, 12, 02, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60))
	return &Dummy{
		Timeline: &message.Timeline{
			Source:  "dummy",
			Subject: "test",
			Messages: []message.Message{
				{
					Timestamp: ts,
					URL:       "https://example.com/1",
					UserName:  "user1",
					Text:      "あああああああ改行\nあああああ",
				},
				{
					Section:   "main",
					URL:       "https://example.com/2",
					Timestamp: ts.Add(2 * time.Minute),
					UserName:  "user2",
					Text:      "いいいいいいい<script>いいいいいいいいいいいいいい",
				},
				{
					Section:   "sub",
					URL:       "https://example.com/3",
					Timestamp: ts.Add(1 * time.Minute),
					UserName:  "user3",
					Text:      "えええええ",
					Attachments: []message.Message{
						{
							Type: "image",
							URL:  "https://www.gravatar.com/avatar/f5d789b9076fd42eaabee3b2941b74db?s=50",
						},
						{
							Type: "image",
							URL:  "https://www.gravatar.com/avatar/f5d789b9076fd42eaabee3b2941b74db?s=50",
						},
						{
							Type: "text",
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
					URL:       "https://example.com/3",
					UserName:  "user2",
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
}

func (p Dummy) Fetch() (*message.Timeline, error) {
	p.Timeline.Sort()
	return p.Timeline, nil
}
