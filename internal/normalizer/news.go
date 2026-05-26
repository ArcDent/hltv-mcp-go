package normalizer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizeNews parses archive news items from HLTV news page HTML
func NormalizeNews(doc *goquery.Document) []types.NewsItem {
	var items []types.NewsItem
	doc.Find(".news-item, article, .newstext").Each(func(_ int, s *goquery.Selection) {
		a := s.Find("a").First()
		title := strings.TrimSpace(a.Text())
		if title == "" {
			return
		}
		link, _ := a.Attr("href")
		items = append(items, types.NewsItem{
			Title:       title,
			Link:        link,
			PublishedAt: strings.TrimSpace(s.Find(".news-date, time, .date").First().Text()),
			Tag:         strings.TrimSpace(s.Find(".news-tag, .tag, .category").First().Text()),
		})
	})
	return items
}

// NormalizeRealtimeNews parses realtime news items from HLTV homepage HTML
func NormalizeRealtimeNews(doc *goquery.Document) []types.RealtimeNewsItem {
	var items []types.RealtimeNewsItem
	doc.Find(".news-item, article, .realtime-news-item, .con-news").Each(func(_ int, s *goquery.Selection) {
		a := s.Find("a").First()
		title := strings.TrimSpace(a.Text())
		if title == "" {
			return
		}
		link, _ := a.Attr("href")
		items = append(items, types.RealtimeNewsItem{
			Section:      "latest",
			Title:        title,
			Link:         link,
			RelativeTime: strings.TrimSpace(s.Find(".time, .relative-time, .news-time").First().Text()),
			Comments:     strings.TrimSpace(s.Find(".comments, .comment-count").First().Text()),
		})
	})
	return items
}
