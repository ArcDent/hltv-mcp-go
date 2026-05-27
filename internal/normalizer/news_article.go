package normalizer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

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
		// fallback: try broader selectors
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
