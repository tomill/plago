package input

import (
	"encoding/json"
	"strings"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Fetcher interface {
	Fetch() (message.Timeline, error)
}

type Raw struct {
	data string
}

func (p Raw) Fetch() (message.Timeline, error) {
	var timeline message.Timeline
	err := json.NewDecoder(strings.NewReader(p.data)).Decode(&timeline)

	return timeline, err
}

func RawFetcher(c config.Config) (Fetcher, error) {
	return &Raw{data: c.RawData}, nil
}

func DummyFetcher(config.Config) (Fetcher, error) {
	data := `
{
  "Source": "dummy",
  "Subject": "test",
  "RefID": "",
  "Messages": [
    {
      "Permalink": "https://example.com/1",
      "Timestamp": "2022-09-01T12:02:00+09:00",
      "Lead": "user1",
      "Text": "あああああああ改行\nあああああ"
    },
    {
      "Section": "main",
      "Permalink": "https://example.com/2",
      "Timestamp": "2022-09-01T12:04:00+09:00",
      "Lead": "user3",
      "Text": "いいいいいいい<script>いいいいいいいいいいいいいい"
    },
    {
      "Section": "main",
      "Permalink": "https://example.com/3",
      "Timestamp": "2022-09-01T12:05:00+09:00",
      "Lead": "user2",
      "Text": "ううう",
      "Attachments": [
        {
          "Timestamp": "0001-01-01T00:00:00Z",
          "Text": "ううううう"
        }
      ]
    },
    {
      "Section": "sub",
      "Permalink": "https://example.com/3",
      "Timestamp": "2022-09-01T12:03:00+09:00",
      "Lead": "user3",
      "Text": "えええええ",
      "Attachments": [
        {
          "Type": "image",
          "Permalink": "https://www.gravatar.com/avatar/f5d789b9076fd42eaabee3b2941b74db?s=50",
          "Timestamp": "0001-01-01T00:00:00Z"
        },
        {
          "Type": "image",
          "Permalink": "https://www.gravatar.com/avatar/f5d789b9076fd42eaabee3b2941b74db?s=50",
          "Timestamp": "0001-01-01T00:00:00Z"
        },
        {
          "Timestamp": "0001-01-01T00:00:00Z",
          "Text": "引用\n引用"
        }
      ]
    },
    {
      "Section": "sub",
      "Timestamp": "2022-09-01T12:05:00+09:00",
      "Text": "おお"
    }
  ]
}`

	return &Raw{data: data}, nil
}
