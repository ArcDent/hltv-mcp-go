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

	bodyEl := doc.Find(".news-block, .news-body, article, .body").First()
	if bodyEl.Length() == 0 {
		bodyEl = doc.Find(".content, .main-content, .article-content").First()
	}
	a.BodyText = strings.TrimSpace(bodyEl.Text())

	return a
}
