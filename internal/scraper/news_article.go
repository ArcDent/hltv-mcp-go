package scraper

import (
	"context"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
)

// NewsArticleScraper scrapes individual HLTV news article pages
type NewsArticleScraper struct{ cli *client.HltvClient }

// NewNewsArticleScraper creates a new NewsArticleScraper
func NewNewsArticleScraper(cli *client.HltvClient) *NewsArticleScraper {
	return &NewsArticleScraper{cli: cli}
}

// GetArticle fetches a news article page and returns the parsed document
func (s *NewsArticleScraper) GetArticle(ctx context.Context, url string) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, url, "news_article")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}
