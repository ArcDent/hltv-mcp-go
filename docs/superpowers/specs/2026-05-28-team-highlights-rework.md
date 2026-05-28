# 队伍详情 — 高亮数据重做

## 概述
用队伍页自带高亮统计区替代不可靠的 results 翻页方案：胜率、连胜、最近5场。

## 后端

### 新增类型 (`internal/types/types.go`)
- `TeamHighlightMatch` — `opponent` + `result`（won/lost）
- `TeamHighlights` — `win_rate`、`win_streak`、`recent_matches`（最多5场）
- `TeamDetail.Highlights *TeamHighlights`

### 新增 normalizer (`internal/normalizer/team.go`)
`NormalizeTeamHighlights(doc)` 从队伍页 HTML 提取：
- `.highlighted-stat` 含 "Win rate" → `win_rate`（如 76.2%）
- `.highlighted-stat` 含 "Current win streak" → `win_streak`
- `.last-5-matches` 内 `.highlighted-stat.text-ellipsis` → 对手名
- `.highlighted-match-status.match-won` / `.match-lost` → 胜负
- 最近5场去重（同一对手可能出现在不同 highlight 区域）

### 修改 `GetTeamDetailCached` (`internal/facade/facade.go`)
- 调用 `NormalizeTeamHighlights(doc)` 填入 `td.Highlights`
- 移除 results 翻页循环（不再需要 `GetResultsOffset` 调用）
- 保留 upcoming matches 获取（显示即将到来的比赛）
- W/L/D 统计改为从 highlights 的 recent_matches 推算

## 前端

### 修改 `TeamDetail.tsx`
- 统计栏：胜率显示 `{highlights.win_rate}`，连胜显示 `{highlights.win_streak} 连胜`
- 近期战绩区：显示最近5场 → 对手名 + W/L 标签
- fallback：若 highlights 为空，显示"暂无数据"
