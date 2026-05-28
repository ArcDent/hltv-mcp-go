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

// NormalizeNewsArticle extracts plain text from a news article page
func NormalizeNewsArticle(doc *goquery.Document, link string) types.NewsArticle {
	a := types.NewsArticle{Link: link}

	titleEl := doc.Find(".news-headline, .article-title, h1").First()
	a.Title = strings.TrimSpace(titleEl.Text())

	dateEl := doc.Find(".news-date, .article-date, .date").First()
	a.PublishedAt = strings.TrimSpace(dateEl.Text())

	authorEl := doc.Find(".news-author, .author-name").First()
	a.Author = strings.TrimSpace(authorEl.Text())

	var paragraphs []string
	doc.Find(".news-block p, .news-body p, article p, .body p").Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" {
			paragraphs = append(paragraphs, t)
		}
	})
	if len(paragraphs) == 0 {
		doc.Find(".content p, .main-content p, .article-content p").Each(func(_ int, s *goquery.Selection) {
			t := strings.TrimSpace(s.Text())
			if t != "" {
				paragraphs = append(paragraphs, t)
			}
		})
	}
	a.BodyText = strings.Join(paragraphs, "\n\n")

	return a
}
