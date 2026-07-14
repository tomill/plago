package input

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/entry"
	"resty.dev/v3"
)

type Feed struct {
	config.ExecParams
	client *resty.Client
}

func FeedFetcher(c config.Config) (Fetcher, error) {
	p := &Feed{
		ExecParams: c.ExecParams,
		client: resty.NewWithClient(httpClient).
			SetBaseURL("https://theoldreader.com/").
			SetHeader("Authorization", "GoogleLogin auth="+c.FeedReaderToken),
	}

	return p, nil
}

type feedItems struct {
	ItemRefs []struct {
		Id string `json:"id"`
	} `json:"itemRefs"`
}

func (res feedItems) AsContentsQuery() url.Values {
	v := url.Values{}
	for _, id := range res.ItemRefs {
		v.Add("i", "tag:google.com,2005:reader/item/"+id.Id) // id
	}
	return v
}

func (p Feed) Fetch() (entry.Timeline, error) {
	timeline := entry.NewTimeline(p.ExecParams)

	var list feedItems
	{
		res, err := p.client.R().
			SetQueryParams(map[string]string{
				"output": "json",
				"s":      "user/-/state/com.google/reading-list", // subscription
				"xt":     "user/-/state/com.google/read",         // exclude
				"n":      "1000",                                 // numbers
			}).
			SetResult(&list).
			Get("reader/api/0/stream/items/ids")

		if err != nil || !res.IsStatusSuccess() {
			return timeline, fmt.Errorf("get unread items error: %w (status: %s)", err, res.Status())
		}
	}

	var contentsResponse struct {
		Items []struct {
			Title      string   `json:"title"`
			Published  int64    `json:"published"`
			Categories []string `json:"categories"`
			Canonical  []struct {
				Href string `json:"href"`
			} `json:"canonical"`
			Origin struct {
				Title string `json:"title"`
				URL   string `json:"htmlUrl"`
			} `json:"origin"`
			Summary struct {
				Content string `json:"content"`
			} `json:"summary"`
			Enclosure []struct {
				URL  string `json:"href"`
				Type string `json:"type"`
			} `json:"enclosure"`
		} `json:"items"`
	}
	{
		res, err := p.client.R().
			SetQueryParam("output", "json").
			SetFormDataFromValues(list.AsContentsQuery()).
			SetResult(&contentsResponse).
			Post("reader/api/0/stream/items/contents")
		if err != nil || !res.IsStatusSuccess() {
			return timeline, fmt.Errorf("get item contents error: %w (status: %s)", err, res.Status())
		}
	}

	for _, item := range contentsResponse.Items {
		e := &entry.Entry{
			User: item.Origin.Title,
			URL:  item.Canonical[0].Href,
			Text: item.Title,
		}
		for _, tag := range item.Categories {
			if strings.HasPrefix(tag, "user/-/label/") {
				e.Channel = strings.Replace(tag, "user/-/label/", "", 1)
			}
		}

		if m := reFirstImageURL.FindStringSubmatch(item.Summary.Content); len(m) > 0 {
			e.AddImage(`https://` + m[1])
		}

		for _, enclosure := range item.Enclosure {
			if strings.HasPrefix(enclosure.Type, "image/") {
				e.AddImage(enclosure.URL)
			}
		}

		timeline.Append(e)
	}

	if len(timeline.Entries) > 0 {
		query := list.AsContentsQuery()
		query.Set("a", "user/-/state/com.google/read") // action
		res, err := p.client.R().
			SetFormDataFromValues(query).
			Post("reader/api/0/edit-tag")
		if err != nil || !res.IsStatusSuccess() {
			return timeline, fmt.Errorf("mark as read error: %w (status: %s)", err, res.Status())
		}
	}

	return timeline.Sorted(), nil
}

var (
	reFirstImageURL = regexp.MustCompile(`<img\s+[^>]*src="https?://([^"]+)"`)
)
