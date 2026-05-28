package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/facade"
	"github.com/arcdent/hltv-mcp/internal/renderer"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// CreateServer registers all 9 MCP tools and returns the configured server
func CreateServer(cfg *config.Config, f *facade.HltvFacade, r *renderer.Renderer) *server.MCPServer {
	s := server.NewMCPServer(cfg.MCPServerName, cfg.MCPServerVersion)

	// 1. resolve_team
	s.AddTool(mcp.NewTool("resolve_team",
		mcp.WithDescription("Resolve a team name to stable HLTV identity candidates."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Team name to search")),
		mcp.WithBoolean("exact", mcp.Description("Exact match only")),
		mcp.WithNumber("limit", mcp.Description("Max results (1-10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.ResolveQuery{
			Name:  req.GetString("name", ""),
			Exact: req.GetBool("exact", false),
			Limit: req.GetInt("limit", 0),
		}
		resp := f.ResolveTeam(q)
		return toolResult(r.RenderResolveResult("队伍候选", resp)), nil
	})

	// 2. resolve_player
	s.AddTool(mcp.NewTool("resolve_player",
		mcp.WithDescription("Resolve a player name to stable HLTV identity candidates."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Player name to search")),
		mcp.WithBoolean("exact", mcp.Description("Exact match only")),
		mcp.WithNumber("limit", mcp.Description("Max results (1-10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.ResolveQuery{
			Name:  req.GetString("name", ""),
			Exact: req.GetBool("exact", false),
			Limit: req.GetInt("limit", 0),
		}
		resp := f.ResolvePlayer(q)
		return toolResult(r.RenderResolveResult("选手候选", resp)), nil
	})

	// 3. hltv_team_recent
	s.AddTool(mcp.NewTool("hltv_team_recent",
		mcp.WithDescription("Get recent state, recent results, and upcoming matches for one team."),
		mcp.WithNumber("team_id", mcp.Description("HLTV team id")),
		mcp.WithString("team_name", mcp.Description("Team name")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.TeamRecentQuery{
			TeamID:   req.GetInt("team_id", 0),
			TeamName: req.GetString("team_name", ""),
			Limit:    req.GetInt("limit", 0),
		}
		resp := f.GetTeamRecent(q)
		return toolResult(r.RenderTeamRecent(resp)), nil
	})

	// 4. hltv_player_recent
	s.AddTool(mcp.NewTool("hltv_player_recent",
		mcp.WithDescription("Get recent state and overview statistics for one player."),
		mcp.WithNumber("player_id", mcp.Description("HLTV player id")),
		mcp.WithString("player_name", mcp.Description("Player name")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.PlayerRecentQuery{
			PlayerID:   req.GetInt("player_id", 0),
			PlayerName: req.GetString("player_name", ""),
			Limit:      req.GetInt("limit", 0),
		}
		resp := f.GetPlayerRecent(q)
		return toolResult(r.RenderPlayerRecent(resp)), nil
	})

	// 5. hltv_results_recent
	s.AddTool(mcp.NewTool("hltv_results_recent",
		mcp.WithDescription("Get recent HLTV results with optional team or event filters."),
		mcp.WithString("team", mcp.Description("Team name filter")),
		mcp.WithString("event", mcp.Description("Event name filter")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-20)")),
		mcp.WithNumber("days", mcp.Description("Time window in days (1-30)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.ResultsRecentQuery{
			Team:  req.GetString("team", ""),
			Event: req.GetString("event", ""),
			Limit: req.GetInt("limit", 0),
			Days:  req.GetInt("days", 0),
		}
		resp := f.GetResultsRecent(q)
		return toolResult(r.RenderMatches(resp)), nil
	})

	// 6. hltv_matches_upcoming
	s.AddTool(mcp.NewTool("hltv_matches_upcoming",
		mcp.WithDescription("Get upcoming HLTV matches for explicit filters."),
		mcp.WithNumber("team_id", mcp.Description("HLTV team id")),
		mcp.WithString("team", mcp.Description("Team name filter — omit for generic requests")),
		mcp.WithString("event", mcp.Description("Event name filter — omit for generic requests")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-20)")),
		mcp.WithNumber("days", mcp.Description("Time window in days (1-30)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.UpcomingMatchesQuery{
			TeamID: req.GetInt("team_id", 0),
			Team:   req.GetString("team", ""),
			Event:  req.GetString("event", ""),
			Limit:  req.GetInt("limit", 0),
			Days:   req.GetInt("days", 0),
		}
		resp := f.GetUpcomingMatches(q)
		return toolResult(r.RenderMatches(resp)), nil
	})

	// 7. hltv_matches_today
	s.AddTool(mcp.NewTool("hltv_matches_today",
		mcp.WithDescription("Get today's HLTV matches in fixed Asia/Shanghai time."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp := f.GetTodayMatches()
		return toolResult(r.RenderMatches(resp)), nil
	})

// 8. hltv_realtime_news
	s.AddTool(mcp.NewTool("hltv_realtime_news",
		mcp.WithDescription("Get realtime/latest HLTV news."),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-50, default 25)")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("offset", mcp.Description("Zero-based offset")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.RealtimeNewsQuery{
			Limit:  req.GetInt("limit", 0),
			Page:   req.GetInt("page", 0),
			Offset: req.GetInt("offset", 0),
		}
		resp := f.GetRealtimeNews(q)
		return toolResult(r.RenderRealtimeNews(resp)), nil
	})

// 9. hltv_news_digest
	s.AddTool(mcp.NewTool("hltv_news_digest",
		mcp.WithDescription("Get HLTV monthly archive news."),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-50)")),
		mcp.WithString("tag", mcp.Description("Archive title/topic filter")),
		mcp.WithNumber("year", mcp.Description("Year")),
		mcp.WithString("month", mcp.Description("Month name or number")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("offset", mcp.Description("Zero-based offset")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.NewsDigestQuery{
			Limit:  req.GetInt("limit", 0),
			Tag:    req.GetString("tag", ""),
			Year:   req.GetInt("year", 0),
			Month:  req.GetString("month", ""),
			Page:   req.GetInt("page", 0),
			Offset: req.GetInt("offset", 0),
		}
		resp := f.GetNewsDigest(q)
		return toolResult(r.RenderNews(resp)), nil
	})

	return s
}

func toolResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(text)},
	}
}

