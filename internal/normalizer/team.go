package normalizer

import (
	"strconv"
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

// NormalizeTeamDetail extracts full team detail from the team page HTML
func NormalizeTeamDetail(doc *goquery.Document) types.TeamDetail {
	td := types.TeamDetail{}

	// Ranking
	rankEl := doc.Find(".profile-team-stat .value, .world-rank, .rank-value").First()
	if rankText := cleanText(rankEl.Text()); rankText != "" {
		td.Ranking.WorldRank, _ = strconv.Atoi(strings.TrimPrefix(rankText, "#"))
	}
	pointsEl := doc.Find(".profile-team-stat .description:contains('points'), .points-value").First()
	if pointsText := cleanText(pointsEl.Text()); pointsText != "" {
		parts := strings.Fields(pointsText)
		for _, p := range parts {
			if n, err := strconv.Atoi(strings.Trim(p, "()")); err == nil {
				td.Ranking.Points = n
				break
			}
		}
	}

	// Achievements
	doc.Find(".trophy, .achievement, .honor-item").Each(func(_ int, s *goquery.Selection) {
		label := cleanText(s.Find(".trophy-name, .achievement-label, .label").First().Text())
		countText := cleanText(s.Find(".trophy-count, .achievement-count, .count").First().Text())
		count, _ := strconv.Atoi(countText)
		if label == "" {
			return
		}
		tier := "a"
		lower := strings.ToLower(label)
		if strings.Contains(lower, "major") {
			tier = "major"
		} else if strings.Contains(lower, "s-tier") || strings.Contains(lower, "intel") || strings.Contains(lower, "esl pro league") || strings.Contains(lower, "blast") {
			tier = "s"
		}
		if strings.Contains(lower, "win streak") || strings.Contains(lower, "连胜") {
			tier = "streak"
		}
		td.Achievements = append(td.Achievements, types.TeamAchievement{
			Label: label, Count: count, Tier: tier,
		})
	})

	// Roster
	doc.Find(".player-card, .teammate, .player-holder").Each(func(_ int, s *goquery.Selection) {
		nameEl := s.Find(".player-name, .name, a[href*='/player/']").First()
		name := cleanText(nameEl.Text())
		if name == "" {
			return
		}
		p := types.TeamRosterPlayer{Name: name}
		href, exists := nameEl.Attr("href")
		if exists && strings.Contains(href, "/player/") {
			parts := strings.Split(strings.Trim(href, "/"), "/")
			for i, part := range parts {
				if part == "player" && i+1 < len(parts) {
					p.ID, _ = strconv.Atoi(parts[i+1])
				}
				if i+2 < len(parts) && part == "player" {
					p.Slug = parts[i+2]
				}
			}
		}
		ratingEl := s.Find(".rating, .player-rating, .stat-rating").First()
		p.Rating, _ = strconv.ParseFloat(cleanText(ratingEl.Text()), 64)
		countryEl := s.Find(".flag, .country, .player-country").First()
		if alt, ok := countryEl.Attr("alt"); ok {
			p.Country = alt
		} else {
			p.Country = cleanText(countryEl.Text())
		}
		td.Roster = append(td.Roster, p)
	})

	return td
}
