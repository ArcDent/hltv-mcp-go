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
	"github.com/arcdent/hltv-mcp/internal/types"
)

var teamLinkRE = regexp.MustCompile(`/team/(\d+)/(.+)`)

type TeamScraper struct{ cli *client.HltvClient }

func NewTeamScraper(cli *client.HltvClient) *TeamScraper { return &TeamScraper{cli: cli} }

func (s *TeamScraper) Search(ctx context.Context, name string) ([]types.ResolvedTeam, error) {
	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(name))
	body, err := s.cli.FetchHTML(ctx, path, "team_search")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var teams []types.ResolvedTeam
	doc.Find("table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		link := sel.Find("a[href*='/team/']")
		if link.Length() == 0 {
			return
		}
		href, _ := link.Attr("href")
		m := teamLinkRE.FindStringSubmatch(href)
		if len(m) < 3 {
			return
		}
		id, _ := strconv.Atoi(m[1])
		slug := m[2]
		name := strings.TrimSpace(link.Text())
		if name == "" || id == 0 {
			return
		}
		teams = append(teams, types.ResolvedTeam{
			Type: "team", ID: id, Name: name, Slug: slug,
		})
	})
	return teams, nil
}

func (s *TeamScraper) GetTeam(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, fmt.Sprintf("/team/%d/%s", id, url.PathEscape(slug)), "team_detail")
}

func (s *TeamScraper) GetTeamMatches(ctx context.Context, id int) (*goquery.Document, error) {
	return fetchDoc(s.cli, ctx, fmt.Sprintf("/team/%d/matches", id), "team_matches")
}
