package input

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
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
			SetBaseURL(c.FeedReaderAPI).
			SetHeader("Authorization", "GoogleLogin auth="+c.FeedReaderToken),
	}

	return p, nil
}

func (p Feed) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)

	var itemIDs []string
	{
		var itemRefs struct {
			ItemRefs []struct {
				ID string `json:"id"`
			} `json:"itemRefs"`
		}
		res, err := p.client.R().
			SetQueryParam("s", "user/-/state/com.google/reading-list"). // subscription
			SetQueryParam("xt", "user/-/state/com.google/read").        // exclude
			SetQueryParam("n", "1000").                                 // numbers
			SetQueryParam("output", "json").
			SetResult(&itemRefs).
			Get("reader/api/0/stream/items/ids")
		if err != nil || !res.IsStatusSuccess() {
			return timeline, fmt.Errorf("get unread item ids error: %w (status: %s)", err, res.Status())
		}
		for _, v := range itemRefs.ItemRefs {
			itemIDs = append(itemIDs, v.ID)
		}
	}

	var items struct {
		Items []feedItem `json:"items"`
	}
	{
		res, err := p.client.R().
			SetQueryParam("output", "json").
			SetFormDataFromValues(url.Values{
				"i": lo.Map(itemIDs, func(v string, _ int) string {
					return "tag:google.com,2005:reader/item/" + v
				}),
			}).
			SetResult(&items).
			Post("reader/api/0/stream/items/contents")
		if err != nil || !res.IsStatusSuccess() {
			return timeline, fmt.Errorf("get feed items error: %w (status: %s)", err, res.Status())
		}
	}

	var readIDs []string
	for _, item := range items.Items {
		if entry := p.build(item); entry != nil {
			timeline.Append(entry)
			readIDs = append(readIDs, item.ID)
		}
	}

	if len(timeline.Entries) > 0 {
		res, err := p.client.R().
			SetQueryParam("a", "user/-/state/com.google/read"). // action = read
			SetFormDataFromValues(url.Values{"i": readIDs}).
			Post("reader/api/0/edit-tag")
		if err != nil || !res.IsStatusSuccess() {
			return timeline, fmt.Errorf("mark as read error: %w (status: %s)", err, res.Status())
		}
	}

	return timeline.Sorted(), nil
}

type feedItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Published     int64    `json:"published"`
	CrawlTimeMsec int64    `json:"crawlTimeMsec,string"`
	Categories    []string `json:"categories"`
	Canonical     []struct {
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
}

var (
	reFirstImageURL = regexp.MustCompile(`<img\s+[^>]*src="(https://[^"]+)"`)
)

func (p Feed) build(item feedItem) *plago.Entry {
	if !timeinrange(time.UnixMilli(item.CrawlTimeMsec), p.ExecParams) {
		return nil
	}

	var ch string
	for _, tag := range item.Categories {
		if after, ok := strings.CutPrefix(tag, "user/-/label/"); ok {
			ch = after
			break
		}
	}

	entry := &plago.Entry{
		Channel:   ch,
		Timestamp: time.Unix(item.Published, 0).In(tz),
		URL:       item.Canonical[0].Href,
		User:      item.Origin.Title,
		Text:      item.Title,
	}

	var img []string
	for _, enclosure := range item.Enclosure {
		if strings.HasPrefix(enclosure.Type, "image/") {
			img = append(img, enclosure.URL)
		}
	}
	if m := reFirstImageURL.FindStringSubmatch(item.Summary.Content); len(m) > 0 {
		img = append(img, m[1])
	}
	if len(img) > 0 {
		entry.AddAttachment(plago.Entry{Images: img})
	}

	return entry
}
