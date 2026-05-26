package normalizer

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizeMatches parses goquery selections into NormalizedMatch slices
func NormalizeMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	doc.Find(".result-con, .match-box, .upcoming-match, .matches .match, table tbody tr").Each(func(_ int, s *goquery.Selection) {
		m := types.NormalizedMatch{Result: types.OutcomeScheduled}
		s.Find(".team, .team-cell, .team-name").Each(func(i int, team *goquery.Selection) {
			name := strings.TrimSpace(team.Text())
			if i == 0 {
				m.Team1 = name
			} else if i == 1 {
				m.Team2 = name
			}
		})
		if score := strings.TrimSpace(s.Find(".result-score, .score, .score-cell").Text()); score != "" {
			m.Score = score
			m.Result = types.OutcomeUnknown
		}
		m.Event = strings.TrimSpace(s.Find(".event-name, .event, .event-cell").Text())
		if t := strings.TrimSpace(s.Find(".time, .date, .match-time").Text()); t != "" {
			if m.Result == types.OutcomeScheduled {
				m.ScheduledAt = t
			} else {
				m.PlayedAt = t
			}
		}
		if href, ok := s.Find("a").Attr("href"); ok {
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
		matches = append(matches, m)
	})
	return matches
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
