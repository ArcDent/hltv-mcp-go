package scraper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/types"
)

type PlayerScraper struct{ cli *client.HltvClient }

func NewPlayerScraper(cli *client.HltvClient) *PlayerScraper { return &PlayerScraper{cli: cli} }

func (s *PlayerScraper) Search(ctx context.Context, name string) ([]types.ResolvedPlayer, error) {
	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(name))
	body, err := s.cli.FetchHTML(ctx, path, "player_search")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytesReader(body))
	if err != nil {
		return nil, err
	}
	var players []types.ResolvedPlayer
	doc.Find(".player-search-result, .player-col, table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		name := cleanText(sel.Find(".player-name, a, td").First().Text())
		if name == "" {
			return
		}
		players = append(players, types.ResolvedPlayer{Type: "player", Name: name, Slug: slugify(name)})
	})
	return players, nil
}

func (s *PlayerScraper) GetPlayer(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/player/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "player_detail")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}

func (s *PlayerScraper) GetPlayerOverview(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/stats/players/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "player_stats")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}
