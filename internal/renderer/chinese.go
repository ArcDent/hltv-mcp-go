package renderer

import (
	"fmt"
	"strings"

	"github.com/arcdent/hltv-mcp/internal/localization"
	"github.com/arcdent/hltv-mcp/internal/summary"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// Renderer produces Chinese-formatted text output for MCP tool responses
type Renderer struct{}

func New() *Renderer { return &Renderer{} }

func (r *Renderer) RenderTeamRecent(resp *types.ToolResponse) string {
	if resp.Error != nil {
		return r.renderError("队伍近况", resp)
	}
	data := resp.Data.(*types.TeamRecentData)
	s := summary.SummarizeTeam(data)
	var b strings.Builder
	fmt.Fprintf(&b, "【队伍近况】%s\n\n", localization.FormatTeamDisplay(data.Profile.Name))
	fmt.Fprintf(&b, "【关键事实】\n排名：#%d  近况：%s\n", data.Profile.Rank, data.SummaryStats.RecentRecord)
	for _, m := range data.RecentResults {
		result := "未知"
		switch m.Result {
		case types.OutcomeWin:
			result = "胜"
		case types.OutcomeLoss:
			result = "负"
		}
		fmt.Fprintf(&b, "- %s %s %s\n", result, localization.FormatTeamDisplay(m.Opponent), m.Score)
	}
	for _, m := range data.UpcomingMatches {
		fmt.Fprintf(&b, "- vs %s %s\n", localization.FormatTeamDisplay(m.Opponent), m.ScheduledAt)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n", s)
	fmt.Fprintf(&b, "\n【更新时间】%s\n【来源】%s\n", resp.Meta.FetchedAt, resp.Meta.Source)
	return b.String()
}

func (r *Renderer) RenderPlayerRecent(resp *types.ToolResponse) string {
	if resp.Error != nil {
		return r.renderError("选手近况", resp)
	}
	data := resp.Data.(*types.PlayerRecentData)
	s := summary.SummarizePlayer(data)
	var b strings.Builder
	fmt.Fprintf(&b, "【选手近况】%s\n\n", data.Profile.Name)
	fmt.Fprintf(&b, "【关键事实】\n所属队伍：%s  国家：%s\n",
		localization.FormatTeamDisplay(data.Profile.Team), data.Profile.Country)
	for k, v := range data.Overview {
		fmt.Fprintf(&b, "- %s: %v\n", k, v)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", s, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderMatches(resp *types.ToolResponse) string {
	if resp.Error != nil {
		return r.renderError("比赛", resp)
	}
	items := resp.Items.([]types.NormalizedMatch)
	title := "未来比赛"
	if q, ok := resp.Query["today_only"].(bool); ok && q {
		title = "今日比赛"
	}
	s := summary.SummarizeMatches(items, title == "今日比赛")
	var b strings.Builder
	fmt.Fprintf(&b, "【%s】\n\n", title)
	for i, m := range items {
		fmt.Fprintf(&b, "%d. %s vs %s", i+1,
			localization.FormatTeamDisplay(m.Team1),
			localization.FormatTeamDisplay(m.Team2))
		if m.Score != "" {
			fmt.Fprintf(&b, " — %s", m.Score)
		}
		if m.Event != "" {
			fmt.Fprintf(&b, " — %s", m.Event)
		}
		fmt.Fprintln(&b)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", s, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderNews(resp *types.ToolResponse) string {
	if resp.Error != nil {
		return r.renderError("新闻", resp)
	}
	items := resp.Items.([]types.NewsItem)
	s := summary.SummarizeNews(items)
	var b strings.Builder
	fmt.Fprintf(&b, "【新闻集合】\n\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s — %s\n", i+1, item.Title, item.PublishedAt)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", s, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderRealtimeNews(resp *types.ToolResponse) string {
	if resp.Error != nil {
		return r.renderError("实时新闻", resp)
	}
	items := resp.Items.([]types.RealtimeNewsItem)
	s := summary.SummarizeRealtimeNews(items)
	var b strings.Builder
	fmt.Fprintf(&b, "【实时新闻】\n\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. [%s] %s — %s\n", i+1, item.Section, item.Title, item.RelativeTime)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", s, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderResolveResult(title string, resp *types.ToolResponse) string {
	if resp.Error != nil {
		return r.renderError(title, resp)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "【%s】\n\n", title)
	switch items := resp.Items.(type) {
	case []types.ResolvedTeam:
		for i, item := range items {
			fmt.Fprintf(&b, "%d. %s (id=%d)\n", i+1, item.Name, item.ID)
		}
	case []types.ResolvedPlayer:
		for i, item := range items {
			fmt.Fprintf(&b, "%d. %s (id=%d)\n", i+1, item.Name, item.ID)
		}
	}
	fmt.Fprintf(&b, "\n【更新时间】%s\n【来源】%s\n", resp.Meta.FetchedAt, resp.Meta.Source)
	return b.String()
}

func (r *Renderer) renderError(title string, resp *types.ToolResponse) string {
	return fmt.Sprintf("【%s】\n请求失败：%s\n%s\n", title, resp.Error.Code, resp.Error.Message)
}
