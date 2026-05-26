package normalizer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizeTeamProfile extracts team profile data from HLTV team page HTML
func NormalizeTeamProfile(doc *goquery.Document, fallback types.ResolvedTeam) types.TeamProfile {
	p := types.TeamProfile{
		ID:      fallback.ID,
		Name:    fallback.Name,
		Slug:    fallback.Slug,
		Country: fallback.Country,
		Rank:    fallback.Rank,
	}
	if name := strings.TrimSpace(doc.Find(".team-name, .profile-team-name, h1").First().Text()); name != "" {
		p.Name = name
	}
	return p
}
