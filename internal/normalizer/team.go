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

	// Profile: name from h1, country from profile container
	td.Profile.Name = cleanText(doc.Find("h1.profile-team-name, h1").First().Text())
	profileText := cleanText(doc.Find(".profile-team-container, .profile-team-info").First().Text())
	if profileText != "" {
		parts := strings.Fields(profileText)
		if len(parts) >= 1 {
			td.Profile.Country = parts[0]
		}
	}

	// Ranking: use .value.h-rank for HLTV world rank
	rankText := cleanText(doc.Find(".value.h-rank").First().Text())
	if rankText != "" {
		td.Ranking.WorldRank, _ = strconv.Atoi(strings.TrimPrefix(rankText, "#"))
	}
	// Also try .profile-team-stat for "World ranking#N"
	if td.Ranking.WorldRank == 0 {
		doc.Find(".profile-team-stat").Each(func(_ int, s *goquery.Selection) {
			t := cleanText(s.Text())
			if strings.HasPrefix(strings.ToLower(t), "world ranking") {
				td.Ranking.WorldRank, _ = strconv.Atoi(strings.TrimPrefix(strings.TrimPrefix(t, "World ranking"), "#"))
			}
		})
	}

	// Achievements: parse trophy links from .trophySection
	seenAchievements := make(map[string]bool)
	doc.Find(".trophySection .trophyDescription, .trophySection .trophy").Each(func(_ int, s *goquery.Selection) {
		label := cleanText(s.AttrOr("title", ""))
		if label == "" {
			label = cleanText(s.Text())
		}
		if label == "" || seenAchievements[label] {
			return
		}
		seenAchievements[label] = true

		tier := "a"
		lower := strings.ToLower(label)
		if strings.Contains(lower, "major") {
			tier = "major"
		} else if strings.Contains(lower, "iem") || strings.Contains(lower, "esl") || strings.Contains(lower, "blast") || strings.Contains(lower, "intel") {
			tier = "s"
		}
		if strings.Contains(lower, "win streak") || strings.Contains(lower, "连胜") || strings.Contains(lower, "streak") {
			tier = "streak"
		}
		td.Achievements = append(td.Achievements, types.TeamAchievement{
			Label: label, Count: 1, Tier: tier,
		})
	})

	// Roster: extract only from the current lineup grid (.bodyshot-team)
	seenIDs := make(map[int]bool)
	doc.Find(".bodyshot-team a[href*='/player/']").Each(func(_ int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if href == "" {
			return
		}
		parts := strings.Split(strings.Trim(href, "/"), "/")
		var pID int
		var pSlug string
		for i, part := range parts {
			if part == "player" && i+1 < len(parts) {
				pID, _ = strconv.Atoi(parts[i+1])
			}
			if i+2 < len(parts) && part == "player" {
				pSlug = parts[i+2]
			}
		}
		if pID == 0 || seenIDs[pID] {
			return
		}
		seenIDs[pID] = true

		p := types.TeamRosterPlayer{
			ID:   pID,
			Name: cleanText(s.Text()),
			Slug: pSlug,
		}
		td.Roster = append(td.Roster, p)
	})

	return td
}
