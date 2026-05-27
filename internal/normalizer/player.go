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

// NormalizePlayerDetail extracts full player detail from HLTV player page (chromedp)
func NormalizePlayerDetail(doc *goquery.Document) types.PlayerDetail {
	pd := types.PlayerDetail{}

	pd.Profile.Name = cleanText(doc.Find(".playerNickname").First().Text())
	if pd.Profile.Name == "" { return pd }

	pd.Profile.RealName = cleanText(doc.Find(".playerRealname").First().Text())
	pd.Profile.Team = cleanText(doc.Find(".playerTeam a").First().Text())
	pd.Profile.Country, _ = doc.Find("img.flag").First().Attr("title")
	pd.Profile.Slug = cleanText(strings.ReplaceAll(strings.ToLower(pd.Profile.Name), " ", "-"))

	doc.Find(".playerAge, .player-info span, .listRight, .playerInfo span").Each(func(_ int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if strings.Contains(t, "years") && pd.Profile.Age == 0 {
			fmt.Sscanf(t, "%d", &pd.Profile.Age)
		}
		if strings.HasPrefix(t, "$") && pd.Profile.PrizeMoney == "" {
			pd.Profile.PrizeMoney = t
		}
	})

	// Rating + maps
	doc.Find(".player-stat").Each(func(_ int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Rating") && pd.Rating.Value == 0 {
			pd.Rating.Value, _ = strconv.ParseFloat(cleanText(s.Find(".statsVal").Text()), 64)
		}
	})
	if t := cleanText(doc.Find(".stats-window").Text()); t != "" {
		fmt.Sscanf(t, "Past %d months • %d maps", new(int), &pd.Rating.Maps)
		if pd.Rating.Maps == 0 { fmt.Sscanf(t, "Past 3 months • %d maps", &pd.Rating.Maps) }
	}

	// Abilities
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
					ab.Value, _ = strconv.ParseFloat(cleanText(s.Find(".statsVal").Text()), 64)
				}
			})
		}
		pd.Abilities = append(pd.Abilities, ab)
	}

	// All-time stats
	doc.Find(".all-time-stat, .highlighted-stat").Each(func(_ int, s *goquery.Selection) {
		valText := cleanText(s.Find(".stat").Text())
		t := s.Text()
		switch {
		case strings.Contains(t, "KDR") && pd.Career.KD == 0: pd.Career.KD, _ = strconv.ParseFloat(valText, 64)
		case strings.Contains(t, "Win rate") && pd.Career.WinRate == "": pd.Career.WinRate = valText
		case strings.Contains(t, "Headshots") && pd.Career.HeadshotPct == "": pd.Career.HeadshotPct = valText
		case strings.Contains(t, "Win streak") && pd.Career.WinStreak == 0: pd.Career.WinStreak, _ = strconv.Atoi(valText)
		case strings.Contains(t, "Matches") && !strings.Contains(t, "Win") && pd.Career.Matches == 0: pd.Career.Matches, _ = strconv.Atoi(valText)
		}
	})

	// Career rating from somewhere in the page
	doc.Find(".player-stat, .statsVal, .all-time-stat").Each(func(_ int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Rating") && pd.Career.Rating == 0 {
			if v, err := strconv.ParseFloat(cleanText(s.Find(".statsVal, .stat").First().Text()), 64); err == nil {
				if pd.Career.Rating == 0 || v < pd.Rating.Value { pd.Career.Rating = v }
			}
		}
	})

	// Top 20
	profileText := cleanText(doc.Find(".playerInfo, .playerSummaryContainer, .profile-summary").Text())
	re := regexp.MustCompile(`#(\d)\s*\('(\d{2})\)`)
	matches := re.FindAllStringSubmatch(profileText, -1)
	if len(matches) > 0 {
		pd.Top20Ranks = make(map[string]int)
		for _, m := range matches {
			rank, _ := strconv.Atoi(m[1])
			pd.Top20Ranks["20"+m[2]] = rank
		}
	}

	// Honors
	doc.Find(".highlighted-stat").Each(func(_ int, s *goquery.Selection) {
		t := s.Text()
		valText := cleanText(s.Find(".stat").Text())
		if strings.Contains(t, "Majors won") { v,_:=strconv.Atoi(valText); pd.Honors=append(pd.Honors,types.PlayerHonor{Label:"Major 冠军",Value:v}) }
		if strings.Contains(t, "Total MVPs") { v,_:=strconv.Atoi(valText); pd.Honors=append(pd.Honors,types.PlayerHonor{Label:"总 MVP",Value:v}) }
		if strings.Contains(t, "Major MVP") || strings.Contains(t, "Major MVPs") { v,_:=strconv.Atoi(valText); pd.Honors=append(pd.Honors,types.PlayerHonor{Label:"Major MVP",Value:v}) }
	})

	// Recent matches (simplified — matches page structure varies)
	doc.Find(".recent-matches .match-row, .result-con, .match-box").Each(func(i int, s *goquery.Selection) {
		if i >= 7 { return }
		m := types.PlayerRecentMatch{Result: "scheduled"}
		m.Team = cleanText(s.Find(".team1 .team, .line-align.team1 .team").First().Text())
		m.Opponent = cleanText(s.Find(".team2 .team, .line-align.team2 .team").First().Text())
		m.Score = cleanText(s.Find(".result-score").First().Text())
		m.Event = cleanText(s.Find(".event-name").First().Text())
		if m.Score != "" { m.Result = "loss"; if strings.Contains(m.Score, "2:") { m.Result = "win" } }
		pd.RecentMatches = append(pd.RecentMatches, m)
	})

	return pd
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
