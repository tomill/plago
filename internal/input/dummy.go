package input

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type Dummy struct {
	config.ExecParams
	data string
}

func DummyFetcher(c config.Config) (Fetcher, error) {
	return &Dummy{
		ExecParams: c.ExecParams,
		data: `
{
  "subject": "test",
  "entries": [
    {
      "section": "2022-09-01 12:00",
      "timestamp": "2022-09-01T12:02:00+09:00",
      "url": "https://example.com/1",
      "user": "成歩堂龍一 / Phoenix Wright",
      "text": "あああああああ\n\n　\n\n　┌○┐ 　　  ／\n　│  勝｜ﾊ,,ﾊ 　／\n　│  　｜ﾟωﾟ )  ／\n　│  訴 |　／/ \n　└○┘ (⌒) 　／\n　　　　し⌒"
    },
    {
      "section": "2022-09-01 12:00",
      "channel": "#main",
      "timestamp": "2022-09-01T12:04:00+09:00",
      "url": "https://example.com/2",
      "user": "user",
      "text": "いいいいいいい<tag></tag>いいいいいいいいいいいいいい"
    },
    null,
    {},
    {
      "section": "2022-09-01 13:00",
      "channel": "#main",
      "timestamp": "2022-09-01T13:05:00+09:00",
      "url": "https://example.com/3",
      "user": "user2",
      "text": "ううううう is:reply, 画像付き、quoteあり画像あり",
      "reply": true,
      "images": [
      	"https://placehold.co/400x300.png",
       	"https://placehold.co/300x400.png"
      ],
      "attachments": [
        {
          "text": "ううううう",
          "images": [
         	"https://placehold.co/400x300.png",
          	"https://placehold.co/300x400.png"
          ]
        }
      ]
    },
    {
      "section": "2022-09-01 13:00",
      "channel": "#sub",
      "timestamp": "2022-09-01T13:03:00+09:00",
      "url": "https://example.com/3",
      "user": "user3",
      "text": "ええええ引用ふたつ",
      "attachments": [
        {
          "text": "引用\n引用",
          "url": "https://example.com/3-1"
        },
        {
          "text": "引用\n引用"
        }
      ]
    },
    {
      "section": "2022-09-01 13:00",
      "channel": "#sub",
      "timestamp": "2022-09-01T13:05:00+09:00",
      "text": "おお"
    }
  ]
}`,
	}, nil
}

func (p *Dummy) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)
	if err := json.NewDecoder(strings.NewReader(p.data)).Decode(&timeline); err != nil {
		return timeline, err
	}

	timeline.AppendError(fmt.Errorf("this is dummy error"))
	return timeline, nil
}
