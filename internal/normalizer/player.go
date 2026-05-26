package normalizer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizePlayerProfile extracts player profile data from HLTV player page HTML
func NormalizePlayerProfile(doc *goquery.Document, fallback types.ResolvedPlayer) types.PlayerProfile {
	p := types.PlayerProfile{
		ID:      fallback.ID,
		Name:    fallback.Name,
		Slug:    fallback.Slug,
		Team:    fallback.Team,
		Country: fallback.Country,
	}
	if name := strings.TrimSpace(doc.Find(".playerNickname, .player-nickname, h1").First().Text()); name != "" {
		p.Name = name
	}
	if team := strings.TrimSpace(doc.Find(".playerTeam a, .player-team a").First().Text()); team != "" {
		p.Team = team
	}
	return p
}

// NormalizeOverview extracts player stats from one or more stat page documents
func NormalizeOverview(docs ...*goquery.Document) map[string]any {
	overview := make(map[string]any)
	for _, doc := range docs {
		doc.Find(".stats-row, .stat, .summaryStatBreakdownRow, .stat-row").Each(func(_ int, s *goquery.Selection) {
			label := strings.TrimSpace(s.Find(".stat-label, .summaryStatBreakdownRowLabel, .label").First().Text())
			value := strings.TrimSpace(s.Find(".stat-value, .summaryStatBreakdownRowValue, .value").First().Text())
			if label != "" && value != "" {
				key := strings.ToLower(label)
				key = strings.ReplaceAll(key, " ", "_")
				overview[key] = value
			}
		})
	}
	return overview
}

// CollectRecentHighlights extracts highlight/achievement text from player doc
func CollectRecentHighlights(doc *goquery.Document) []string {
	var highlights []string
	doc.Find(".achievement, .highlight, .trophy").Each(func(_ int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			highlights = append(highlights, text)
		}
	})
	return highlights
}
