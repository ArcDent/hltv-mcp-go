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
	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(name))
	body, err := s.cli.FetchHTML(ctx, path, "team_search")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytesReader(body))
	if err != nil {
		return nil, err
	}
	var teams []types.ResolvedTeam
	doc.Find(".team-search-result, .team-col, table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		link, _ := sel.Find("a").Attr("href")
		name := sel.Find(".team-name, a, td").First().Text()
		name = cleanText(name)
		if name == "" {
			return
		}
		t := types.ResolvedTeam{Type: "team", Name: name, Slug: slugify(name)}
		if link != "" {
			_ = link // extract ID from link if possible
		}
		teams = append(teams, t)
	})
	return teams, nil
}

func (s *TeamScraper) GetTeam(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/team/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "team_detail")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}

func (s *TeamScraper) GetTeamMatches(ctx context.Context, id int) (*goquery.Document, error) {
	path := fmt.Sprintf("/team/%d/matches", id)
	body, err := s.cli.FetchHTML(ctx, path, "team_matches")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}
