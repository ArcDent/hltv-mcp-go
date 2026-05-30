package scraper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/types"
)

type TeamScraper struct{ cli *client.HltvClient }

func NewTeamScraper(cli *client.HltvClient) *TeamScraper { return &TeamScraper{cli: cli} }

func (s *TeamScraper) Search(ctx context.Context, name string) ([]types.ResolvedTeam, error) {
	hits, err := searchHLTV(ctx, s.cli, name, "team", "team_search")
	if err != nil {
		return nil, err
	}
	teams := make([]types.ResolvedTeam, len(hits))
	for i, h := range hits {
		teams[i] = types.ResolvedTeam{Type: "team", ID: h.id, Name: h.name, Slug: h.slug}
	}
	return teams, nil
}

func (s *TeamScraper) GetTeam(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, fmt.Sprintf("/team/%d/%s", id, url.PathEscape(slug)), "team_detail")
}

func (s *TeamScraper) GetTeamMatches(ctx context.Context, id int) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, fmt.Sprintf("/team/%d/matches", id), "team_matches")
}
