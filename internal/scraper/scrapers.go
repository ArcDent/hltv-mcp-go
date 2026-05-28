package scraper

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
)

type ResultsScraper struct{ cli *client.HltvClient }

func NewResultsScraper(cli *client.HltvClient) *ResultsScraper { return &ResultsScraper{cli: cli} }

func (s *ResultsScraper) GetResults(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/results", "results")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}

type MatchesScraper struct{ cli *client.HltvClient }

func NewMatchesScraper(cli *client.HltvClient) *MatchesScraper { return &MatchesScraper{cli: cli} }

func (s *MatchesScraper) GetUpcoming(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/matches", "matches_upcoming")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}

type NewsScraper struct{ cli *client.HltvClient }

func NewNewsScraper(cli *client.HltvClient) *NewsScraper { return &NewsScraper{cli: cli} }

func (s *NewsScraper) GetNews(ctx context.Context, year int, month string) (*goquery.Document, error) {
	path := "/news/archive"
	if year > 0 && month != "" {
		path = fmt.Sprintf("/news/archive/%d/%s", year, month)
	}
	body, err := s.cli.FetchHTML(ctx, path, "news_archive")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}

type RealtimeNewsScraper struct{ cli *client.HltvClient }

func NewRealtimeNewsScraper(cli *client.HltvClient) *RealtimeNewsScraper {
	return &RealtimeNewsScraper{cli: cli}
}

func (s *RealtimeNewsScraper) GetRealtimeNews(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/", "realtime_news")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}

// shared helpers
func cleanText(s string) string { return strings.TrimSpace(s) }
