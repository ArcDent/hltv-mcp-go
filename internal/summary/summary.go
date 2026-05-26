package summary

import (
	"fmt"
	"strings"

	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/localization"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// Service generates Chinese natural-language summaries from structured data
type Service struct {
	mode config.SummaryMode
}

func New(mode config.SummaryMode) *Service { return &Service{mode: mode} }

func (s *Service) SummarizeTeam(data *types.TeamRecentData) string {
	if s.mode == config.SummaryRaw {
		return "已启用 raw 模式，当前未生成自然语言摘要。"
	}
	if data == nil {
		return "无法生成队伍摘要。"
	}
	name := localization.FormatTeamDisplay(data.Profile.Name)
	rank := "排名未知"
	if data.Profile.Rank > 0 {
		rank = fmt.Sprintf("排名约 #%d", data.Profile.Rank)
	}
	record := data.SummaryStats.RecentRecord
	nextMatch := ""
	if len(data.UpcomingMatches) > 0 {
		m := data.UpcomingMatches[0]
		opp := m.Opponent
		if opp == "" {
			opp = m.Team2
		}
		if opp != "" {
			nextMatch = fmt.Sprintf("，下一场对阵 %s", localization.FormatTeamDisplay(opp))
		}
	}
	return fmt.Sprintf("%s %s，近况 %s%s。", name, rank, record, nextMatch)
}

func (s *Service) SummarizePlayer(data *types.PlayerRecentData) string {
	if s.mode == config.SummaryRaw {
		return "已启用 raw 模式，当前未生成自然语言摘要。"
	}
	if data == nil {
		return "无法生成选手摘要。"
	}
	team := localization.FormatTeamDisplay(data.Profile.Team)
	return fmt.Sprintf("%s（%s）近期状态概览。", data.Profile.Name, team)
}

func (s *Service) SummarizeMatches(items []types.NormalizedMatch, todayOnly bool) string {
	if s.mode == config.SummaryRaw {
		return "已启用 raw 模式。"
	}
	if len(items) == 0 {
		return "暂无比赛数据。"
	}
	var parts []string
	for i, m := range items {
		if i >= 2 {
			break
		}
		parts = append(parts, fmt.Sprintf("%s vs %s",
			localization.FormatTeamDisplay(m.Team1),
			localization.FormatTeamDisplay(m.Team2)))
	}
	prefix := "赛程"
	if todayOnly {
		prefix = "今日赛程"
	}
	return fmt.Sprintf("%s重点：%s。", prefix, strings.Join(parts, "；"))
}

func (s *Service) SummarizeNews(items []types.NewsItem) string {
	if s.mode == config.SummaryRaw {
		return "已启用 raw 模式。"
	}
	if len(items) == 0 {
		return "暂无新闻。"
	}
	var parts []string
	for i, item := range items {
		if i >= 3 {
			break
		}
		parts = append(parts, item.Title)
	}
	return fmt.Sprintf("重点新闻：%s。", strings.Join(parts, "；"))
}

func (s *Service) SummarizeRealtimeNews(items []types.RealtimeNewsItem) string {
	if s.mode == config.SummaryRaw {
		return "已启用 raw 模式。"
	}
	if len(items) == 0 {
		return "暂无实时新闻。"
	}
	var parts []string
	for i, item := range items {
		if i >= 3 {
			break
		}
		parts = append(parts, item.Title)
	}
	return fmt.Sprintf("实时新闻：%s。", strings.Join(parts, "；"))
}
