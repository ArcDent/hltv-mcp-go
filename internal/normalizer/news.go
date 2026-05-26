package normalizer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizeNews parses archive news items from HLTV news page HTML
// HLTV structure: .newstext div contains the title text directly (no <a> inside)
// The link is in a parent <a> tag wrapping the entire news block
func NormalizeNews(doc *goquery.Document) []types.NewsItem {
	var items []types.NewsItem
	doc.Find(".newstext").Each(func(_ int, s *goquery.Selection) {
		title := cleanText(s.Text())
		if title == "" {
			return
		}
		// Look for link in ancestors or sibling containers
		link, _ := s.Parent().Find("a").First().Attr("href")
		if link == "" {
			// Try climbing up to the parent container
			link, _ = s.Parent().Parent().Find("a").Attr("href")
		}
		// Date is in sibling .newstc > .newsrecent
		date := cleanText(s.Parent().Find(".newsrecent").First().Text())
		items = append(items, types.NewsItem{
			Title:       title,
			Link:        link,
			PublishedAt: date,
		})
	})
	return items
}

// NormalizeRealtimeNews parses realtime news items from HLTV homepage HTML
func NormalizeRealtimeNews(doc *goquery.Document) []types.RealtimeNewsItem {
	var items []types.RealtimeNewsItem
	// HLTV homepage has news links in various containers
	doc.Find("a[href*='/news/']").Each(func(_ int, s *goquery.Selection) {
		title := cleanText(s.Text())
		title = strings.Join(strings.Fields(title), " ")
		if title == "" {
			return
		}
		link, _ := s.Attr("href")
		// Only include actual news links (not navigation)
		if !strings.HasPrefix(link, "/news/") {
			return
		}
		items = append(items, types.RealtimeNewsItem{
			Section: "latest",
			Title:   title,
			Link:    link,
		})
	})
	return items
}
