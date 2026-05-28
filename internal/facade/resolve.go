package facade

import (
	"context"
	"fmt"

	"github.com/arcdent/hltv-mcp/internal/errors"
	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// ResolveTeam resolves a team name to HLTV identity candidates
func (f *HltvFacade) ResolveTeam(query types.ResolveQuery) *types.ToolResponse {
	q := map[string]any{"name": query.Name, "exact": query.Exact}
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	key := fmt.Sprintf("resolve_team:%s:%v:%d", query.Name, query.Exact, query.Limit)
	ttl := f.cfg.CacheTTLEntity

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		items, err := f.ts.Search(context.Background(), query.Name)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, errors.New(errors.CodeEntityNotFound,
				fmt.Sprintf("No team matched '%s'", query.Name), false, q)
		}
		if len(items) > query.Limit {
			items = items[:query.Limit]
		}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

// ResolvePlayer resolves a player name to HLTV identity candidates
func (f *HltvFacade) ResolvePlayer(query types.ResolveQuery) *types.ToolResponse {
	q := map[string]any{"name": query.Name, "exact": query.Exact}
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	key := fmt.Sprintf("resolve_player:%s:%v:%d", query.Name, query.Exact, query.Limit)
	ttl := f.cfg.CacheTTLEntity

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		items, err := f.ps.Search(context.Background(), query.Name)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, errors.New(errors.CodeEntityNotFound,
				fmt.Sprintf("No player matched '%s'", query.Name), false, q)
		}
		if len(items) > query.Limit {
			items = items[:query.Limit]
		}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

// GetTeamRecent fetches a team's recent results, upcoming matches, and stats
func (f *HltvFacade) GetTeamRecent(query types.TeamRecentQuery) *types.ToolResponse {
	if query.TeamID == 0 && query.TeamName == "" {
		return &types.ToolResponse{Error: &types.ToolError{
			Code: "INVALID_ARGUMENT", Message: "team_id or team_name required", Retryable: false,
		}}
	}
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	teamID := query.TeamID
	teamName := query.TeamName
	q := map[string]any{"team_id": teamID, "team_name": teamName}
	key := fmt.Sprintf("team_recent:%d:%s", teamID, teamName)
	ttl := f.cfg.CacheTTLTeam

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		teamSearch := f.ResolveTeam(types.ResolveQuery{Name: teamName, Limit: 1})
		if teamSearch.Error != nil {
			return nil, fmt.Errorf("%s", teamSearch.Error.Message)
		}
		teams, _ := teamSearch.Items.([]types.ResolvedTeam)
		if len(teams) == 0 {
			return nil, fmt.Errorf("team not found: %s", teamName)
		}
		team := teams[0]

		doc, err := f.ts.GetTeam(context.Background(), team.ID, team.Slug)
		if err != nil {
			return nil, err
		}
		profile := normalizer.NormalizeTeamProfile(doc, team)

		matchDoc, err := f.ts.GetTeamMatches(context.Background(), team.ID)
		if err != nil {
			return nil, err
		}
		matchList := normalizer.NormalizeMatches(matchDoc, profile.Name)
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
				case types.OutcomeWin: wins++
				case types.OutcomeLoss: losses++
				case types.OutcomeDraw: draws++
				}
			}
			record := fmt.Sprintf("%dW-%dL", wins, losses)
			if draws > 0 { record += fmt.Sprintf("-%dD", draws) }
			data := types.TeamRecentData{
				Profile: profile, RecentResults: recent, UpcomingMatches: upcoming,
				SummaryStats: types.TeamSummaryStats{
					Wins: wins, Losses: losses, Draws: draws, RecentRecord: record,
				},
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Data: data, ResolvedEntity: team, Meta: meta}, nil
	})
}

// GetPlayerRecent fetches a player's profile, stats, and highlights
func (f *HltvFacade) GetPlayerRecent(query types.PlayerRecentQuery) *types.ToolResponse {
	if query.PlayerID == 0 && query.PlayerName == "" {
		return &types.ToolResponse{Error: &types.ToolError{
			Code: "INVALID_ARGUMENT", Message: "player_id or player_name required", Retryable: false,
		}}
	}
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	playerID := query.PlayerID
	playerName := query.PlayerName
	q := map[string]any{"player_id": playerID, "player_name": playerName}
	key := fmt.Sprintf("player_recent:%d:%s", playerID, playerName)
	ttl := f.cfg.CacheTTLPlayer

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		playerSearch := f.ResolvePlayer(types.ResolveQuery{Name: playerName, Limit: 1})
		if playerSearch.Error != nil {
			return nil, fmt.Errorf("%s", playerSearch.Error.Message)
		}
		players, _ := playerSearch.Items.([]types.ResolvedPlayer)
		if len(players) == 0 {
			return nil, fmt.Errorf("player not found: %s", playerName)
		}
		player := players[0]

		doc, err := f.ps.GetPlayer(context.Background(), player.ID, player.Slug)
		if err != nil {
			return nil, err
		}
		profile := normalizer.NormalizePlayerProfile(doc, player)

		overviewDoc, err := f.ps.GetPlayerOverview(context.Background(), player.ID, player.Slug)
		if err != nil {
			return nil, err
		}
		overview := normalizer.NormalizeOverview(overviewDoc)

		recentMatches := normalizer.NormalizeMatches(doc, profile.Name)
		normalizer.SortByPlayedAtDesc(recentMatches)
		if query.Limit > 0 && len(recentMatches) > query.Limit {
			recentMatches = recentMatches[:query.Limit]
		}

		highlights := normalizer.CollectRecentHighlights(doc)
		if len(highlights) > query.Limit {
			highlights = highlights[:query.Limit]
		}

		data := types.PlayerRecentData{
			Profile:          profile,
			Overview:         overview,
			RecentMatches:    recentMatches,
			RecentHighlights: highlights,
		}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Data: data, ResolvedEntity: player, Meta: meta}, nil
	})
}

