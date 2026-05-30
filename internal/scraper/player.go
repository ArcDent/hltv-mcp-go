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
	hits, err := searchHLTV(ctx, s.cli, name, "player", "player_search")
	if err != nil {
		return nil, err
	}
	players := make([]types.ResolvedPlayer, len(hits))
	for i, h := range hits {
		players[i] = types.ResolvedPlayer{Type: "player", ID: h.id, Name: h.name, Slug: h.slug}
	}
	return players, nil
}

func (s *PlayerScraper) GetPlayer(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, fmt.Sprintf("/player/%d/%s", id, url.PathEscape(slug)), "player_detail")
}

func (s *PlayerScraper) GetPlayerOverview(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, fmt.Sprintf("/stats/players/%d/%s", id, url.PathEscape(slug)), "player_stats")
}
