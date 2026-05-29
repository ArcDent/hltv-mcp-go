package scraper

import (
	"bytes"
	"context"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
)

func fetchDoc(cli *client.HltvClient, ctx context.Context, path, key string) (*goquery.Document, error) {
	body, err := cli.FetchHTML(ctx, path, key)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}

type ResultsScraper struct{ cli *client.HltvClient }

func NewResultsScraper(cli *client.HltvClient) *ResultsScraper { return &ResultsScraper{cli: cli} }

func (s *ResultsScraper) GetResults(ctx context.Context) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, "/results", "results")
}

type MatchesScraper struct{ cli *client.HltvClient }

func NewMatchesScraper(cli *client.HltvClient) *MatchesScraper { return &MatchesScraper{cli: cli} }

func (s *MatchesScraper) GetUpcoming(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/matches", "matches_upcoming")
	if err != nil {
		body, err = s.cli.FetchViaFirecrawl(ctx, "/matches")
		if err != nil {
			return nil, err
		}
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}

type NewsScraper struct{ cli *client.HltvClient }

func NewNewsScraper(cli *client.HltvClient) *NewsScraper { return &NewsScraper{cli: cli} }

func (s *NewsScraper) GetNews(ctx context.Context, year int, month string) (*goquery.Document, error) {
	path := "/news/archive"
	if year > 0 && month != "" {
		path = fmt.Sprintf("/news/archive/%d/%s", year, month)
	}
	return fetchDoc(s.cli, ctx, path, "news_archive")
}

type RealtimeNewsScraper struct{ cli *client.HltvClient }

func NewRealtimeNewsScraper(cli *client.HltvClient) *RealtimeNewsScraper {
	return &RealtimeNewsScraper{cli: cli}
}

func (s *RealtimeNewsScraper) GetRealtimeNews(ctx context.Context) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, "/", "realtime_news")
}

type NewsArticleScraper struct{ cli *client.HltvClient }

func NewNewsArticleScraper(cli *client.HltvClient) *NewsArticleScraper {
	return &NewsArticleScraper{cli: cli}
}

func (s *NewsArticleScraper) GetArticle(ctx context.Context, url string) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, url, "news_article")
}
