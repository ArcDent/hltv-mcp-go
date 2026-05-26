package scraper

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/types"
)

var playerLinkRE = regexp.MustCompile(`/player/(\d+)/(.+)`)

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
	doc.Find("table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		link := sel.Find("a[href*='/player/']")
		if link.Length() == 0 {
			return
		}
		href, _ := link.Attr("href")
		m := playerLinkRE.FindStringSubmatch(href)
		if len(m) < 3 {
			return
		}
		id, _ := strconv.Atoi(m[1])
		slug := m[2]
		name := cleanText(link.Text())
		if name == "" || id == 0 {
			return
		}
		players = append(players, types.ResolvedPlayer{
			Type: "player", ID: id, Name: name, Slug: slug,
			Aliases: []string{name, slug},
		})
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
