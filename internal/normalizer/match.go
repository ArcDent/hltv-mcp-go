package normalizer

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizeMatches parses HLTV HTML result/match rows into NormalizedMatch slices.
// Use ".result-con" for results, ".upcoming-match" or ".match-box" for upcoming.
func NormalizeMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	return normalizeResultsCon(doc, perspective)
}

// normalizeResultsCon handles the "/results" page structure
func normalizeResultsCon(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	doc.Find(".result-con").Each(func(_ int, s *goquery.Selection) {
		m := types.NormalizedMatch{Result: types.OutcomeUnknown}

		// Extract team names from dedicated .team divs within .line-align containers
		m.Team1 = cleanText(s.Find(".line-align.team1 .team").First().Text())
		m.Team2 = cleanText(s.Find(".line-align.team2 .team").First().Text())

		// Score from .result-score
		if score := cleanText(s.Find(".result-score").First().Text()); score != "" {
			m.Score = score
		}

		// Event name from the event-cell or map-text
		m.Event = cleanText(s.Find(".event-name, .map-text, .stars").First().Text())

		// Match link with ID
		if href, ok := s.Find("a.a-reset").First().Attr("href"); ok && href != "" {
			if id := parseMatchID(href); id > 0 {
				m.MatchID = id
			}
		}

		// Time/date
		if t := cleanText(s.Find(".time, .date").First().Text()); t != "" {
			if m.Score != "" {
				m.PlayedAt = t
			} else {
				m.ScheduledAt = t
			}
		}

		if perspective != "" {
			if m.Team1 == perspective {
				m.Opponent = m.Team2
			} else if m.Team2 == perspective {
				m.Opponent = m.Team1
			}
		}

		// Only include if we have at least one team identified
		if m.Team1 != "" || m.Team2 != "" {
			matches = append(matches, m)
		}
	})
	return matches
}

// NormalizeUpcomingMatches handles the "/matches" page structure (after Cloudflare bypass)
func NormalizeUpcomingMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	doc.Find(".upcoming-match, .match-box, .matchCard").Each(func(_ int, s *goquery.Selection) {
		m := types.NormalizedMatch{Result: types.OutcomeScheduled}

		m.Team1 = cleanText(s.Find(".team1 .team, .matchTeam1 .team, .team-1").First().Text())
		m.Team2 = cleanText(s.Find(".team2 .team, .matchTeam2 .team, .team-2").First().Text())
		m.Event = cleanText(s.Find(".matchEventName, .event-name, .event").First().Text())
		m.ScheduledAt = cleanText(s.Find(".matchTime, .time, .date").First().Text())

		if href, ok := s.Find("a").First().Attr("href"); ok {
			if id := parseMatchID(href); id > 0 {
				m.MatchID = id
			}
		}

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

// SortByScheduledAtAsc sorts matches in ascending order by scheduled_at
func SortByScheduledAtAsc(matches []types.NormalizedMatch) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ScheduledAt < matches[j].ScheduledAt
	})
}
