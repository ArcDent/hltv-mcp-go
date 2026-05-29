package facade

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// Strip generic placeholder values (spec: non-regression behavior)
var genericPatterns = regexp.MustCompile(`^(?:today|matches?|schedule|fixtures?|全部|所有|比赛|赛程|今日赛程|今日比赛|今天比赛|今天赛程|未来赛程|未来比赛)$`)

func isGenericFilterText(s string) bool { return genericPatterns.MatchString(strings.TrimSpace(s)) }

func isPlaceholderText(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "" || s == "x" || s == "y" || s == "z" || s == "?" ||
		s == "-" || s == "n/a" || s == "null" || s == "undefined" || s == "tbd" || s == "none"
}

func stripGenericFilter(s string) string {
	s = strings.TrimSpace(s)
	if isGenericFilterText(s) || isPlaceholderText(s) {
		return ""
	}
	s = regexp.MustCompile(`^(?:today|upcoming|未来|即将)?\s*`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\s*(?:matches?|match|schedule|比赛|赛程)?\s*$`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// EventGroup holds matches grouped under one event name
type EventGroup struct {
	Name       string                  `json:"name"`
	DateStart  string                  `json:"date_start"`
	DateEnd    string                  `json:"date_end"`
	MatchCount int                     `json:"match_count"`
	Matches    []types.NormalizedMatch `json:"matches"`
}

// EventsResponse is the response for /api/events
type EventsResponse struct {
	Events []EventGroup            `json:"events"`
	Other  []types.NormalizedMatch `json:"other,omitempty"`
}

// GetEvents fetches matches and groups them by event name
func (f *HltvFacade) GetEvents(matchType string, limit int) *types.ToolResponse {
	q := map[string]any{"type": matchType, "limit": limit}
	key := fmt.Sprintf("events:%s:%d", matchType, limit)
	ttl := f.cfg.CacheTTLMatches

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		var items []types.NormalizedMatch
		switch matchType {
		case "today", "upcoming":
			doc, err := f.ms.GetUpcoming(context.Background())
			if err != nil {
				return nil, err
			}
			items = normalizer.NormalizeUpcomingMatches(doc, "")
			if matchType == "today" {
				items = filterToday(items)
			}
		case "results":
			doc, err := f.rs.GetResults(context.Background())
			if err != nil {
				return nil, err
			}
			items = normalizer.NormalizeMatches(doc, "")
			normalizer.SortByPlayedAtDesc(items)
		default:
			return nil, fmt.Errorf("invalid type: %s", matchType)
		}
		if len(items) > limit {
			items = items[:limit]
		}

		resp := groupByEvent(items)
		meta := f.createMeta(ttl)
		r := &types.ToolResponse{Query: q, Meta: meta}
		r.Data = resp
		return r, nil
	})
}

func groupByEvent(matches []types.NormalizedMatch) EventsResponse {
	groups := make(map[string]*EventGroup)
	var other []types.NormalizedMatch

	for _, m := range matches {
		event := strings.TrimSpace(m.Event)
		if event == "" {
			other = append(other, m)
			continue
		}
		g, exists := groups[event]
		if !exists {
			g = &EventGroup{Name: event}
			groups[event] = g
		}
		g.Matches = append(g.Matches, m)
		g.MatchCount++

		date := m.PlayedAt
		if date == "" {
			date = m.ScheduledAt
		}
		if date == "" {
			continue
		}
		date = strings.SplitN(date, " ", 2)[0]
		if g.DateStart == "" || date < g.DateStart {
			g.DateStart = date
		}
		if g.DateEnd == "" || date > g.DateEnd {
			g.DateEnd = date
		}
	}

	var events []EventGroup
	for _, g := range groups {
		events = append(events, *g)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].DateStart < events[j].DateStart
	})

	return EventsResponse{Events: events, Other: other}
}

// GetTodayMatches delegates to GetUpcomingMatches with empty query
func (f *HltvFacade) GetTodayMatches() *types.ToolResponse {
	return f.GetUpcomingMatches(types.UpcomingMatchesQuery{TodayOnly: true})
}

// GetUpcomingMatches fetches upcoming matches with optional filters
func (f *HltvFacade) GetUpcomingMatches(query types.UpcomingMatchesQuery) *types.ToolResponse {
	team := stripGenericFilter(query.Team)
	event := stripGenericFilter(query.Event)

	if isPlaceholderText(query.Team) && isPlaceholderText(query.Event) && query.Limit == 1 && query.Days == 1 {
		team = ""
		event = ""
	}
	todayOnly := query.TodayOnly
	userSetLimit := query.Limit // 0 means no explicit limit from user
	if query.Limit == 0 {
		query.Limit = 300 // internal max to prevent unbounded response
	}
	q := map[string]any{"team": team, "event": event, "today_only": todayOnly}
	key := fmt.Sprintf("matches_upcoming:%s:%s:%v", team, event, todayOnly)
	ttl := f.cfg.CacheTTLMatches

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.ms.GetUpcoming(context.Background())
		if err != nil {
			return nil, err
		}
		items := normalizer.NormalizeUpcomingMatches(doc, "")
		normalizer.SortByScheduledAtAsc(items)
		if todayOnly {
			items = filterToday(items)
		}
		if userSetLimit > 0 && len(items) > userSetLimit {
			items = items[:userSetLimit]
		}
		if !todayOnly && len(items) > query.Limit {
			items = items[:query.Limit] // hard cap at 300
		}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

// GetResultsRecent fetches recent results with optional filters
func (f *HltvFacade) GetResultsRecent(query types.ResultsRecentQuery) *types.ToolResponse {
	team := stripGenericFilter(query.Team)
	event := stripGenericFilter(query.Event)
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	if query.Days == 0 {
		query.Days = 7
	}
	q := map[string]any{"team": team, "event": event, "days": query.Days}
	key := fmt.Sprintf("results_recent:%s:%s:%d", team, event, query.Days)
	ttl := f.cfg.CacheTTLResults

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.rs.GetResults(context.Background())
		if err != nil {
			return nil, err
		}
		items := normalizer.NormalizeMatches(doc, "")
		normalizer.SortByPlayedAtDesc(items)
		if len(items) > query.Limit {
			items = items[:query.Limit]
		}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

func filterToday(matches []types.NormalizedMatch) []types.NormalizedMatch {
	today := time.Now().Format("2006-01-02")
	var result []types.NormalizedMatch
	for _, m := range matches {
		if strings.HasPrefix(m.ScheduledAt, today) {
			result = append(result, m)
		}
	}
	return result
}
