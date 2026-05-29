package normalizer

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

var monthMap = map[string]string{
	"January": "01", "February": "02", "March": "03", "April": "04",
	"May": "05", "June": "06", "July": "07", "August": "08",
	"September": "09", "October": "10", "November": "11", "December": "12",
}

var resultsDateRe = regexp.MustCompile("Results for (\\w+) (\\d+)(?:st|nd|rd|th)? (\\d{4})")

func parseResultsDate(headline string) string {
	m := resultsDateRe.FindStringSubmatch(headline)
	if len(m) != 4 {
		return ""
	}
	month, ok := monthMap[m[1]]
	if !ok {
		return ""
	}
	day := m[2]
	if len(day) == 1 {
		day = "0" + day
	}
	return m[3] + "-" + month + "-" + day
}

// NormalizeMatches parses HLTV HTML result/match rows into NormalizedMatch slices.
// Use ".result-con" for results, ".upcoming-match" or ".match-box" for upcoming.
func NormalizeMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	return normalizeResultsCon(doc, perspective)
}

// normalizeResultsCon handles the "/results" page structure
func normalizeResultsCon(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	doc.Find(".results-sublist").Each(func(_ int, sublist *goquery.Selection) {
		date := parseResultsDate(cleanText(sublist.Find(".standard-headline").First().Text()))
		sublist.Find(".result-con").Each(func(_ int, s *goquery.Selection) {
			m := types.NormalizedMatch{Result: types.OutcomeUnknown}

			m.Team1 = cleanText(s.Find(".line-align.team1 .team").First().Text())
			m.Team2 = cleanText(s.Find(".line-align.team2 .team").First().Text())

			if score := cleanText(s.Find(".result-score").First().Text()); score != "" {
				m.Score = score
			}

			m.Event = cleanText(s.Find(".event-name, .map-text, .stars").First().Text())

			if href, ok := s.Find("a.a-reset").First().Attr("href"); ok && href != "" {
				if id := parseMatchID(href); id > 0 {
					m.MatchID = id
				}
			}

			m.PlayedAt = date

			if perspective != "" {
				if m.Team1 == perspective {
					m.Opponent = m.Team2
				} else if m.Team2 == perspective {
					m.Opponent = m.Team1
				}
			}

			if m.Team1 != "" || m.Team2 != "" {
				matches = append(matches, m)
			}
		})
	})
	return matches
}

func NormalizeUpcomingMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	currentDate := strings.Split(time.Now().UTC().Format(time.RFC3339), "T")[0] // "2006-01-02"

	// Find parent container holding both headlines and match-wrappers
	headline := doc.Find(".matches-list-headline").First()
	var parent *goquery.Selection = doc.Find("body").First()
	if headline.Length() > 0 {
		if p := headline.Parent(); p != nil && p.Length() > 0 {
			parent = p
		}
	}

	parent.Children().Each(func(_ int, child *goquery.Selection) {
		if child.HasClass("matches-list-headline") {
			text := cleanText(child.Text())
			if idx := strings.LastIndex(text, "- "); idx >= 0 {
				currentDate = strings.TrimSpace(text[idx+2:])
			}
			return
		}
		child.Find(".match").Each(func(_ int, s *goquery.Selection) {
			m := types.NormalizedMatch{Result: types.OutcomeScheduled}

			m.Event = cleanText(s.Find(".match-top").First().Text())

			infoText := cleanText(s.Find(".match-info").First().Text())
			if idx := strings.Index(infoText, " "); idx > 0 {
				m.ScheduledAt = currentDate + " " + infoText[:idx]
				m.BestOf = cleanText(infoText[idx:])
			} else {
				m.ScheduledAt = currentDate + " " + infoText
			}

			teamsText := cleanText(s.Find(".match-teams").First().Text())
			teamsText = strings.ReplaceAll(teamsText, "\n", " ")
			teamsText = strings.ReplaceAll(teamsText, "  ", " ")
			if idx := strings.Index(teamsText, " vs "); idx > 0 {
				m.Team1 = cleanText(teamsText[:idx])
				m.Team2 = cleanText(teamsText[idx+4:])
			} else {
				parts := strings.Fields(teamsText)
				if len(parts) >= 2 {
					m.Team1 = parts[0]
					if strings.ToLower(parts[len(parts)-2]) == "vs" {
						m.Team2 = parts[len(parts)-1]
					} else {
						m.Team2 = parts[len(parts)-1]
					}
				}
			}

			s.Find("a").Each(func(_ int, a *goquery.Selection) {
				if href, ok := a.Attr("href"); ok {
					if id := parseMatchID(href); id > 0 {
						m.MatchID = id
					}
				}
			})

			if perspective != "" {
				if m.Team1 == perspective {
					m.Opponent = m.Team2
				} else if m.Team2 == perspective {
					m.Opponent = m.Team1
				}
			}
			m.Team1 = TranslatePlaceholder(m.Team1)
			m.Team2 = TranslatePlaceholder(m.Team2)
			m.Opponent = TranslatePlaceholder(m.Opponent)

			if m.Team1 != "" && m.Team2 != "" {
				matches = append(matches, m)
			}
		})
	})
	return matches
}

func cleanText(s string) string {
	return strings.TrimSpace(s)
}

func parseMatchID(href string) int {
	re := regexp.MustCompile(`/matches/(\d+)/`)
	if m := re.FindStringSubmatch(href); len(m) > 1 {
		if id, err := strconv.Atoi(m[1]); err == nil {
			return id
		}
	}
	return 0
}

// SplitTeamMatches separates matches into recent (played) and upcoming (scheduled)
func SplitTeamMatches(matches []types.NormalizedMatch) (recent, upcoming []types.NormalizedMatch) {
	for _, m := range matches {
		if m.Score != "" || m.PlayedAt != "" {
			recent = append(recent, m)
		}
		if m.ScheduledAt != "" {
			upcoming = append(upcoming, m)
		}
	}
	return
}

// SortByPlayedAtDesc sorts matches in descending order by played_at
func SortByPlayedAtDesc(matches []types.NormalizedMatch) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].PlayedAt > matches[j].PlayedAt
	})
}

// TranslatePlaceholder maps HLTV bracket placeholder team names to Chinese
func TranslatePlaceholder(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return s
	}
	if strings.Contains(lower, "winner") {
		return "胜者"
	}
	if strings.Contains(lower, "loser") {
		return "败者"
	}
	if strings.Contains(lower, "tbd") {
		return "待定"
	}
	return s
}

// SortByScheduledAtAsc sorts matches in ascending order by scheduled_at
func SortByScheduledAtAsc(matches []types.NormalizedMatch) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ScheduledAt < matches[j].ScheduledAt
	})
}
