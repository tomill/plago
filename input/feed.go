package input

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

// Feed https://github.com/theoldreader/api
type Feed struct {
	since  time.Time
	client *sling.Sling
}

func FeedFetcher(c config.Config) (Fetcher, error) {
	p := &Feed{
		since: c.Since,
		client: sling.New().
			Base("https://theoldreader.com/").
			Set("Authorization", "GoogleLogin auth="+c.TheOldReaderToken),
	}

	return p, nil
}

func (p Feed) Fetch() (message.Timeline, error) {
	timeline := message.Timeline{
		Source:  "feed",
		Subject: p.since.Format(time.DateOnly),
	}

	var list feedItems
	{
		query := struct {
			Subscription string `url:"s,omitempty"`
			Exclude      string `url:"xt,omitempty"`
			Numbers      int    `url:"n,omitempty"`
		}{
			Subscription: "user/-/state/com.google/reading-list",
			Exclude:      "user/-/state/com.google/read",
			Numbers:      1000,
		}

		req := p.client.New().Get("reader/api/0/stream/items/ids?output=json").QueryStruct(query)
		if err := p.call(req, &list); err != nil {
			return timeline, fmt.Errorf("theoldreader get unread items error: %w", err)
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
		} `json:"items"`
	}
	{
		req := p.client.New().Post("reader/api/0/stream/items/contents?output=json").BodyForm(list.AsContentsQuery())
		if err := p.call(req, &contentsResponse); err != nil {
			return timeline, fmt.Errorf("theoldreader get item contents error: %w", err)
		}
	}

	for _, item := range contentsResponse.Items {
		msg := message.Message{
			UserName:  item.Origin.Title,
			Permalink: item.Canonical[0].Href,
			Text:      item.Title,
		}
		for _, tag := range item.Categories {
			if strings.HasPrefix(tag, "user/-/label/") {
				msg.Channel = strings.Replace(tag, "user/-/label/", "", 1)
			}
		}

		timeline.Append(msg)
	}

	if len(timeline.Messages) > 0 {
		query := list.AsContentsQuery()
		query.Action = "user/-/state/com.google/read"
		req := p.client.New().Post("reader/api/0/edit-tag").BodyForm(query)
		if err := p.call(req, nil); err != nil {
			return timeline, fmt.Errorf("theoldreader mark as read error: %w", err)
		}
	}

	return timeline.Sorted(), nil
}

type feedItems struct {
	ItemRefs []struct {
		Id string `json:"id"`
	} `json:"itemRefs"`
}

func (res feedItems) AsContentsQuery() struct {
	Action string   `url:"a,omitempty"`
	Ids    []string `url:"i,omitempty"`
} {
	var ids []string
	for _, id := range res.ItemRefs {
		ids = append(ids, "tag:google.com,2005:reader/item/"+id.Id)
	}

	return struct {
		Action string   `url:"a,omitempty"`
		Ids    []string `url:"i,omitempty"`
	}{
		Ids: ids,
	}
}

func (p Feed) call(req *sling.Sling, v any) error {
	var res *http.Response
	var err error
	if v == nil {
		r, _ := req.Request()
		res, err = http.DefaultClient.Do(r)
	} else {
		res, err = req.ReceiveSuccess(&v)
	}

	if err != nil {
		return fmt.Errorf("theoldreader http call error: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request error: %s - %s", res.Request.URL.Path, res.Status)
	}

	return nil
}
