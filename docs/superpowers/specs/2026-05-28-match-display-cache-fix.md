# 赛程日期修复 & UI 调整

## 概述
修复4个 bug：队伍详情缺少历史比赛、赛事名称截断、赛果无日期、赛程无日期。

## 后端改动

### 1. results 页面日期提取（修复2）
**文件:** `internal/normalizer/match.go`

- results 页面结构：`.results-all > .results-sublist`，每段含 `.standard-headline`（"Results for May 28th 2026"）+ 若干 `.result-con`
- 新增字段或处理逻辑：遍历 `.results-sublist`，从 headline 用正则 `Results for (\w+) (\d+)(?:st|nd|rd|th)? (\d{4})` 提取日期，转为 `YYYY-MM-DD`
- 将该日期赋给该 sublist 下所有 `.result-con` 比赛的 `played_at`

### 2. matches 页面日期提取（修复4）
**文件:** `internal/normalizer/match.go`

- matches 页面结构：日期标题 `.matches-list-headline`（"Thursday - 2026-05-28"）+ 下方 `.match-wrapper` 列表
- 遍历页面：遇到 `.matches-list-headline` 记录当前日期，后续 `.match` 都关联该日期
- 从 "Thursday - 2026-05-28" 提取 `YYYY-MM-DD` 部分
- `scheduled_at` 改为 `YYYY-MM-DD HH:MM` 完整格式

### 3. 翻页获取队伍历史比赛（修复1a）
**文件:** `internal/normalizer/match.go`、`internal/facade/facade.go`、`internal/scraper/scrapers.go`（如需要公开 `GetResultsOffset`）

- 同时创建 `NormalizeResultsSublist(doc, teamName)` 函数：遍历 `.results-sublist`，提取 section 日期 + `.result-con` 比赛，按队名过滤
- `GetTeamDetailCached` 中翻页：循环 offset 0/100/200，调用 `GetResultsOffset`，用新 normalizer 提取带日期的比赛
- 每页结果立即按队名过滤并追加到 `allMatches`
- 上限 3 页或已凑够 10 场

### 4. 修复 results page `GetResultsOffset` 可访问性
**文件:** `internal/scraper/scrapers.go`

- `ResultsScraper.GetResultsOffset(ctx, offset)` 方法已存在，确认其可用即可

## 前端改动

### 5. 赛事名放宽（修复1b）
**文件:** `frontend/src/components/TeamDetail.tsx:157`

- 赛事名 `maxWidth:80` → `maxWidth:140`，或直接删除 maxWidth 约束，用 flex 自然分配

### 6. 对阵居中（修复3）
**文件:** `frontend/src/components/TeamDetail.tsx:146-159`

- match row 布局改为居中：`justifyContent:'center'`，`textAlign:'center'`
- W/L badge + 队名 + 比分 + 对手名 显示在同一行

### 7. 弹窗时间日期显示（修复4，去除 BO3 后时间）
**文件:** `frontend/src/pages/Matches.tsx:137-178`

- scheduled 比赛：中央显示时间 `HH:MM`（大字号），下方日期 `MM/DD`（小字号，始终显示，不仅限于 >24h）
- 右侧栏：只显示 BO3，不带日期

### 8. 赛事卡片日期（修复4）
**文件:** `frontend/src/pages/Matches.tsx:99-101`

- 赛事卡片日期从 `date_start`（现已是 `YYYY-MM-DD` 格式）格式化为 `MM/DD ~ MM/DD`

## 约束
- 不引入新依赖
- 不修改 types 结构体的 JSON 字段名
- `PlayedAt` 格式为 `YYYY-MM-DD`（results 页无精确时间）
- `ScheduledAt` 格式改为 `YYYY-MM-DD HH:MM`
