package scraper

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
)

type searchHit struct {
	id   int
	name string
	slug string
}

func searchHLTV(ctx context.Context, cli *client.HltvClient, query, entity, label string) ([]searchHit, error) {
	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(query))
	body, err := cli.FetchHTML(ctx, path, label)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(fmt.Sprintf(`/%s/(\d+)/(.+)`, entity))
	selPattern := fmt.Sprintf("a[href*='/%s/']", entity)
	var hits []searchHit
	doc.Find("table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		link := sel.Find(selPattern)
		if link.Length() == 0 {
			return
		}
		href, _ := link.Attr("href")
		m := re.FindStringSubmatch(href)
		if len(m) < 3 {
			return
		}
		id, _ := strconv.Atoi(m[1])
		name := strings.TrimSpace(link.Text())
		if name == "" || id == 0 {
			return
		}
		hits = append(hits, searchHit{id: id, name: name, slug: m[2]})
	})
	return hits, nil
}

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
