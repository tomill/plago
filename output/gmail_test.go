package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomill/plago/entry"
)

func TestGmailTemplate(t *testing.T) {
	timeline := entry.Timeline{
		Entries: []entry.Entry{
			{
				User: `成歩堂龍一 / Ryuichi Naruhodo`,
				Text: `1行目
　
2行目↓


https://www.amazon.co.jp/%E6%AD%A3%E8%A6%8F%E8%A1%A8%E7%8F%BE-%E7%AC%AC3%E7%89%88-Jeffrey-F-Friedl/dp/4873113598/ref=sr_1_1?__mk_ja_JP=%E3%82%AB%E3%82%BF%E3%82%AB%E3%83%8A&crid=8JXLWOQ7WFRS&dib=eyJ2IjoiMSJ9.3LZA8ZPZ4KJ-M8dQ55SVaHtRIQ8vyjubLZVj4yInmq2uxOy4LftAa8szvaXL051iHK6PkfGwh87w8pF9lyQwkUOXQrgP7ynEnrFoE31OjyybbtTmXLBizSSKjFvq2YFZEc7ISP3gzqKDg5FqJ8_pCqd9fHF6_3wES0ocRsgkAmLG1ZafS6rQ8M795QKbkjWrLuE7PlJl2GTgE_c4SjnP2QnvMGa7Y4LC_7lhjxaqxGyxJ5J-8l3UvSy13jN076bGb42CWek84u-7BEpe3JzhS1jmoM7lgDfvr42K4xKfIZc.UOvLZid5F91PwJ3_WYIFp1DVvVYy-65-c48wbOTTKBc&dib_tag=se&keywords=%E6%AD%A3%E8%A6%8F%E8%A1%A8%E7%8F%BE&qid=1784873931&sprefix=%E6%AD%A3%E8%A6%8F%E8%A1%A8%E7%8F%BE%2Caps%2C215&sr=8-1
https://example.com/?utm_source=mail&utm_medium=email&utm_campaign=mail_signature&refid=foo`,
			},
		},
	}

	res, err := (&Gmail{}).body(timeline)
	assert.NoError(t, err)
	assert.Contains(t, res, `<div style="margin:0 0 0.4em;color:#222">
		成歩堂龍一 / Ryui… `+"\u00a0"+`1行目<br/>
2行目↓<br/>
https://www.amazon.co.jp/dp/4873113598/<br/>
https://example.com/???&amp;refid=foo
	</div>`)
}
