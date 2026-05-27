package facade

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// Thin wrappers so the facade can call normalizer functions without importing normalizer in every file

func normalizeMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	return normalizer.NormalizeMatches(doc, perspective)
}

func normalizeTeamProfile(doc *goquery.Document, fallback types.ResolvedTeam) types.TeamProfile {
	return normalizer.NormalizeTeamProfile(doc, fallback)
}

func normalizeTeamDetail(doc *goquery.Document) types.TeamDetail {
	return normalizer.NormalizeTeamDetail(doc)
}

func normalizePlayerProfile(doc *goquery.Document, fallback types.ResolvedPlayer) types.PlayerProfile {
	return normalizer.NormalizePlayerProfile(doc, fallback)
}

func normalizeOverview(docs ...*goquery.Document) map[string]any {
	return normalizer.NormalizeOverview(docs...)
}

func collectRecentHighlights(doc *goquery.Document) []string {
	return normalizer.CollectRecentHighlights(doc)
}

func sortByPlayedAtDesc(matches []types.NormalizedMatch) {
	normalizer.SortByPlayedAtDesc(matches)
}

func sortByScheduledAtAsc(matches []types.NormalizedMatch) {
	normalizer.SortByScheduledAtAsc(matches)
}

func splitTeamMatches(matches []types.NormalizedMatch) ([]types.NormalizedMatch, []types.NormalizedMatch) {
	return normalizer.SplitTeamMatches(matches)
}

func buildTeamRecentResponse(
	q map[string]any,
	query types.TeamRecentQuery,
	ttl int,
	f *HltvFacade,
	profile types.TeamProfile,
	matchList []types.NormalizedMatch,
	team types.ResolvedTeam,
) (*types.ToolResponse, error) {
	recent, upcoming := normalizer.SplitTeamMatches(matchList)
	normalizer.SortByPlayedAtDesc(recent)
	normalizer.SortByScheduledAtAsc(upcoming)

	if query.Limit > 0 && len(recent) > query.Limit {
		recent = recent[:query.Limit]
	}
	if query.Limit > 0 && len(upcoming) > query.Limit {
		upcoming = upcoming[:query.Limit]
	}

	wins, losses, draws := 0, 0, 0
	for _, m := range recent {
		switch m.Result {
		case types.OutcomeWin:
			wins++
		case types.OutcomeLoss:
			losses++
		case types.OutcomeDraw:
			draws++
		}
	}
	record := fmt.Sprintf("%dW-%dL", wins, losses)
	if draws > 0 {
		record += fmt.Sprintf("-%dD", draws)
	}

	data := types.TeamRecentData{
		Profile:         profile,
		RecentResults:   recent,
		UpcomingMatches: upcoming,
		SummaryStats: types.TeamSummaryStats{
			Wins: wins, Losses: losses, Draws: draws, RecentRecord: record,
		},
	}
	meta := f.createMeta(ttl)
	return &types.ToolResponse{Query: q, Data: data, ResolvedEntity: team, Meta: meta}, nil
}
