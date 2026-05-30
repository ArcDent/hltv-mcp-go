package normalizer

import (
	"fmt"
	"regexp"
	"strconv"
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
	if team := strings.TrimSpace(doc.Find(".playerTeam a[itemprop=\"text\"], .player-team a[itemprop=\"text\"]").First().Text()); team != "" {
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

// NormalizePlayerDetail extracts full player detail from HLTV player page 
func NormalizePlayerDetail(doc *goquery.Document) types.PlayerDetail {
	pd := types.PlayerDetail{}

	pd.Profile.Name = cleanText(doc.Find(".playerNickname").First().Text())
	if pd.Profile.Name == "" { return pd }

	pd.Profile.RealName = cleanText(doc.Find(".playerRealname").First().Text())
	pd.Profile.Team = cleanText(doc.Find(".playerTeam a[itemprop=\"text\"]").First().Text())
	pd.Profile.Country, _ = doc.Find("img.flag").First().Attr("title")
	pd.Profile.Slug = cleanText(strings.ReplaceAll(strings.ToLower(pd.Profile.Name), " ", "-"))

	// Age from .playerInfoRow.playerAge
	doc.Find(".playerInfoRow").Each(func(_ int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if strings.Contains(t, "Age") && pd.Profile.Age == 0 {
			ageText := cleanText(s.Find(".listRight").Text())
			fmt.Sscanf(ageText, "%d", &pd.Profile.Age)
		}
	})

	// Rating from player-stat > Rating 3.0
	doc.Find(".player-stat").Each(func(_ int, s *goquery.Selection) {
		t := s.Text()
		if (strings.Contains(t, "Rating") || strings.Contains(t, "Rating 2") || strings.Contains(t, "Rating 3")) && pd.Rating.Value == 0 {
			pd.Rating.Value, _ = strconv.ParseFloat(cleanText(s.Find(".statsVal p").First().Text()), 64)
		}
	})

	// Maps count from .stats-window "(Past 3 months • 47 maps)"
	if t := cleanText(doc.Find(".stats-window").Text()); t != "" {
		t = strings.Trim(t, "()")
		re := regexp.MustCompile(`(\d+)\s*maps?`)
		if m := re.FindStringSubmatch(t); len(m) > 1 {
			pd.Rating.Maps, _ = strconv.Atoi(m[1])
		}
	}

	// Abilities — value is inside <b> within <p> within .statsVal (e.g. "100/100")
	abilityDefs := []struct{ key, en, zh string; isRating bool }{
		{"rating","Rating","综合",true},{"firepower","Firepower","火力",false},
		{"opening","Opening","突破",false},{"clutching","Clutching","残局",false},
		{"sniping","Sniping","狙击",false},{"entrying","Entrying","进点",false},
		{"trading","Trading","补枪",false},{"utility","Utility","道具",false},
	}
	for _, def := range abilityDefs {
		ab := types.PlayerAbility{Key:def.key, LabelEn:def.en, LabelZh:def.zh, Max:100}
		if def.isRating {
			ab.Value = pd.Rating.Value; ab.Max = 0; ab.Format = "decimal"
		} else {
			doc.Find(".player-stat").Each(func(_ int, s *goquery.Selection) {
				if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(def.en)) && ab.Value == 0 {
					// Extract <b> value from <p><b>100</b><span>/100</span></p>
					ab.Value, _ = strconv.ParseFloat(cleanText(s.Find(".statsVal p b").First().Text()), 64)
				}
			})
		}
		pd.Abilities = append(pd.Abilities, ab)
	}

	// All-time stats from .all-time-stat (old layout): Matches, K/D, Win rate, Headshots, Win streak
	doc.Find(".all-time-stat").Each(func(_ int, s *goquery.Selection) {
		valText := cleanText(s.Find(".stat").Text())
		desc := cleanText(s.Find(".description").Text())
		switch {
		case strings.Contains(desc, "K/D") || strings.Contains(desc, "KDR"): pd.Career.KD, _ = strconv.ParseFloat(valText, 64)
		case strings.Contains(desc, "Win rate") && !strings.Contains(desc, "Streak"): pd.Career.WinRate = valText
		case strings.Contains(desc, "Headshots"): pd.Career.HeadshotPct = valText
		case strings.Contains(desc, "Win streak"): pd.Career.WinStreak, _ = strconv.Atoi(valText)
		case desc == "Matches": pd.Career.Matches, _ = strconv.Atoi(valText)
		}
	})

	// Summary stats from .highlighted-stat (available on both old and new layout)
	var sum types.PlayerSummary
	doc.Find(".highlighted-stat").Each(func(_ int, s *goquery.Selection) {
		valText := cleanText(s.Find(".stat").Text())
		desc := cleanText(s.Find(".description").Text())
		if desc == "" {
			desc = cleanText(strings.Replace(s.Text(), valText, "", 1))
		}
		v, _ := strconv.Atoi(strings.ReplaceAll(valText, ",", ""))
		switch {
		case desc == "Teams": sum.Teams = v
		case desc == "Days in current team": sum.DaysInTeam = v
		case desc == "Days in teams": sum.DaysInTeams = v
		case desc == "Major won": sum.MajorWon = v
		case desc == "Majors played": sum.MajorsPlayed = v
		case desc == "LANs won": sum.LANsWon = v
		case desc == "LANs played": sum.LANsPlayed = v
		case desc == "Major trophies": sum.MajorTrophies = v
		case desc == "Notable trophies": sum.NotableTrophies = v
		case desc == "Major MVPs": sum.MajorMVPs = v
		case desc == "Total MVPs": sum.TotalMVPs = v
		case desc == "Major EVPs": sum.MajorEVPs = v
		case desc == "Total EVPs": sum.TotalEVPs = v
		}
	})
	if sum.Teams > 0 || sum.MajorWon > 0 || sum.LANsWon > 0 || sum.MajorTrophies > 0 {
		pd.Summary = &sum
	}

	// Top 20 — from .playerInfoRow.playerTop20
	doc.Find(".playerTop20, .playerInfo, .profile-summary, .playerSummaryContainer").Each(func(_ int, s *goquery.Selection) {
		if len(pd.Top20Ranks) > 0 { return }
		t := cleanText(s.Text())
		re := regexp.MustCompile(`#(\d+)\s*\('?(\d{2,4})'?\)?`)
		matches := re.FindAllStringSubmatch(t, -1)
		if len(matches) > 0 {
			pd.Top20Ranks = make(map[string]int)
			for _, m := range matches {
				rank, _ := strconv.Atoi(m[1])
				year := m[2]
				if len(year) == 2 { year = "20" + year }
				pd.Top20Ranks[year] = rank
			}
		}
	})

	// Honors — from .majorSection and .trophySection
	doc.Find(".majorWinner").Each(func(_ int, s *goquery.Selection) {
		v, _ := strconv.Atoi(cleanText(s.Find("b").First().Text()))
		if v > 0 {
			pd.Honors = append(pd.Honors, types.PlayerHonor{Label:"Major 冠军", Value:v})
		}
	})
	doc.Find(".majorMVP").Each(func(_ int, s *goquery.Selection) {
		v, _ := strconv.Atoi(cleanText(s.Find("b").First().Text()))
		if v > 0 {
			pd.Honors = append(pd.Honors, types.PlayerHonor{Label:"Major MVP", Value:v})
		}
	})
	doc.Find(".mvp-count").Each(func(_ int, s *goquery.Selection) {
		v, _ := strconv.Atoi(cleanText(s.Text()))
		if v > 0 {
			pd.Honors = append(pd.Honors, types.PlayerHonor{Label:"总 MVP", Value:v})
		}
	})

	// Recent matches — from .playerpage-matchbox
	doc.Find(".playerpage-matchbox").Each(func(i int, s *goquery.Selection) {
		if i >= 7 { return }
		m := types.PlayerRecentMatch{Result: "scheduled"}

		// Extract match ID and full path from href
		if href, ok := s.Attr("href"); ok {
			// Try stats match URL first: /stats/matches/126993/spirit-vs-falcons
			if re := regexp.MustCompile(`/stats/matches/(\d+)/([^"\s]+)`); re != nil {
				if mid := re.FindStringSubmatch(href); len(mid) > 1 {
					m.MatchID, _ = strconv.Atoi(mid[1])
					m.MatchSlug = mid[2]
				}
			}
			// Fallback: try regular match URL pattern
			if m.MatchID == 0 {
				if re := regexp.MustCompile(`/matches/(\d+)/([^"\s]+)`); re != nil {
					if mid := re.FindStringSubmatch(href); len(mid) > 1 {
						m.MatchID, _ = strconv.Atoi(mid[1])
						m.MatchSlug = mid[2]
					}
				}
			}
		}

		// Determine result from class
		if s.HasClass("won-matchbox") {
			m.Result = "win"
		} else if s.HasClass("lost-matchbox") {
			m.Result = "loss"
		}

		m.Opponent = cleanText(s.Find(".playerpage-matchbox-team .text-ellipsis").First().Text())
		m.Event = cleanText(s.Find(".playerpage-matchbox-bottom").First().Text())
		m.Date = cleanText(s.Find(".playerpage-match-date").First().Text())
		m.Score = strings.ReplaceAll(cleanText(s.Find(".playerpage-match-result").First().Text()), " ", "")
		m.Score = normalizeBO1Score(m.Score)
		if m.Opponent == "" {
			m.Opponent = cleanText(s.Find(".playerpage-matchbox-team").First().Text())
		}
		m.Opponent = translatePlaceholder(m.Opponent)
		m.Team = pd.Profile.Team
		m.Team = translatePlaceholder(m.Team)

		pd.RecentMatches = append(pd.RecentMatches, m)
	})

	// Also try .result-con for match data if playerpage-matchbox found nothing
	if len(pd.RecentMatches) == 0 {
		doc.Find(".result-con").Each(func(i int, s *goquery.Selection) {
			if i >= 7 { return }
			m := types.PlayerRecentMatch{Result: "scheduled"}
			m.Team = cleanText(s.Find(".team1 .team, .line-align.team1 .team").First().Text())
			if m.Team == "" { m.Team = pd.Profile.Team }
			m.Team = translatePlaceholder(m.Team)
			m.Opponent = cleanText(s.Find(".team2 .team, .line-align.team2 .team").First().Text())
			m.Opponent = translatePlaceholder(m.Opponent)
			m.Score = strings.ReplaceAll(cleanText(s.Find(".result-score").First().Text()), " ", "")
			m.Event = cleanText(s.Find(".event-name").First().Text())
			if m.Score != "" { m.Result = "loss"; if strings.Count(m.Score, "2") >= 1 { m.Result = "win" } }
			m.Score = normalizeBO1Score(m.Score)
			pd.RecentMatches = append(pd.RecentMatches, m)
		})
	}

	return pd
}

// normalizeBO1Score converts BO1 match scores (e.g. "13:5") to "1:0"/"0:1"
// If both sides < 13, returns the original score unchanged.
func normalizeBO1Score(score string) string {
	parts := strings.SplitN(score, ":", 2)
	if len(parts) != 2 {
		return score
	}
	a, errA := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, errB := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errA != nil || errB != nil {
		return score
	}
	if a >= 13 || b >= 13 {
		if a > b {
			return "1:0"
		}
		if b > a {
			return "0:1"
		}
		return "平局"
	}
	return score
}

func CollectRecentHighlights(doc *goquery.Document) []string {
	var highlights []string
	doc.Find(".achievement, .highlight, .trophy").Each(func(_ int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			highlights = append(highlights, text)
		}
	})
	return highlights
}
