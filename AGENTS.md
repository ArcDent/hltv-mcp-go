# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈
- 目标：Go 单二进制，MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS v4
- 远端仓库：[ArcDent/HLTV-data](https://github.com/ArcDent/HLTV-data)

## 项目静态结构
```
├── main.go                  # MCP stdio + HTTP :8082 双 goroutine
├── Dockerfile               # 三阶段：frontend → Go → chromedp/headless-shell
├── internal/
│   ├── types/               # 共享类型
│   ├── errors/              # AppError（4 错误码）
│   ├── config/              # 环境变量配置
│   ├── crypto/              # AES-256-GCM 加解密
│   ├── cache/               # TTL + stale + 并发合并
│   ├── client/              # HTTP + chromedp 反CF
│   ├── scraper/             # team/player/results/matches/news
│   ├── localization/        # 26 队伍 + 98 选手中英映射
│   ├── normalizer/          # HLTV HTML → 标准化类型
│   ├── facade/              # 核心编排层
│   ├── summary/             # 中文摘要
│   ├── renderer/            # 中文格式化 MCP 输出
│   ├── mcp/                 # 9 MCP 工具
│   └── http/                # chi router + REST API + SPA fallback
├── frontend/                # React + Vite + Tailwind
│   └── src/
│       ├── api/client.ts    # API 客户端
│       ├── components/      # 6 组件
│       └── pages/           # 6 页面
```

## 最近操作
- 2026-05-29：深度收敛 — 删 9 个文件（.clinerules-*×5, .playwright-mcp/, SummaryHint 字段, PlayerNickname/TeamNickname 函数, NormalizeCareerFromOverview, cleanText 重复定义, 4 未用错误码, 5 未用类型字段）；前端 API 调用收敛（3 组件改回使用 api.* 方法）
- 2026-05-29：HLTV 选手页 A/B 改版适配 — 发现新版页面移除 `.all-time-stat`；前端修复 React 零值渲染陷阱
- 2026-05-28：前端别名编辑 UX 改进 + chromedp headless-shell flag 修复 + Docker SSL 修复 + 昵称字典后端迁移

## 关键发现

### HLTV HTML 选择器（核心参考）
- **选手页**: `.playerNickname` / `.playerRealname` / `.playerTeam a[itemprop="text"]` / `.player-stat` > `.statsVal p b`(能力值) / `.stats-window`(maps数) / `.playerpage-matchbox`(近期比赛) / `.playerpage-match-result`(比分) / `.playerpage-match-date` / `.majorWinner b`(Major冠军数) / `.mvp-count`(MVP数) / `.all-time-stat` > `.stat` + `.description`(生涯统计，旧版) / `.playerInfoRow.playerAge` / `.playerTop20`
- **队伍页**: `h1.profile-team-name` / `.value.h-rank` / `.bodyshot-team a[href*='/player/']`(队员) / `.trophySection .trophyDescription[title]` / `.highlighted-stat`(胜率/连胜)
- **比赛链接**: `.playerpage-matchbox[href]` 正则 `/stats/matches/(\d+)/([^"\s]+)`
- **赛果**: `.result-con` > `.line-align.team1 .team` / `.result-score` / `.event-name`
- **赛程**: `.matches-list-headline` + `.match-wrapper` > `.match-top`(赛事) + `.match-teams`(队伍) + `.match-info`(时间)
- **新闻**: `.newstext` / `.news-block p`(正文) — 不可用 `.Text()` 取整个容器
- **搜索**: `table tbody tr > a[href*='/player/']` 正则 `/player/(\d+)/(.+)`

### HLTV 选手页 A/B 改版
- **sh1ro 为新版**：无 `.all-time-stat`，生涯数据在 `/stats/players/` 页（`.stats-rows > .stats-row > span` 结构，headless-shell 无法通过 CF）
- **s1mple/ZywOo 为旧版**：`.all-time-stat` 仍存在，生涯数据正常提取

### CF 分层
- HTTP 直连可用：`/player/`、`/results`、`/matches`、`/team/`、`/news/`
- chromedp 可用：`/player/`、`/results`、`/matches`、`/team/`
- **被阻断**：`/stats/matches/`、`/stats/players/`（JS Challenge 在 headless-shell 中无法完成）

### chromedp 关键配置
- `chromedp.DefaultExecAllocatorOptions` 内置 `--headless`，headless-shell 需 `Flag("headless", false)` 覆盖
- 反 CF：`UserDataDir`(持久化 profile) + `--disable-blink-features=AutomationControlled` + Chrome UA
- 10s NewContext 超时保护

### 缓存模式
- `PlayerDetail` 不走 `withCache`，直接用 `cache.Get/Set`，key 格式 `player_detail:<id>`
- `sync/atomic.Int64` 计数器，与 `sync.RWMutex` 无锁竞争

### nickname 覆盖层
- `internal/localization/overrides.go`：线程安全内存缓存 + JSON 持久化
- 空值语义 = 删除条目；写操作先更新内存再写磁盘
- API：`PUT /api/nicknames/team` 先解析 canonical 名；`PUT /api/nicknames/player` 直接存储

### 部署
- Docker 三阶段构建 → `ghcr.io/arcdent/hltv-data:latest`
- CI/CD：push main → GitHub Actions 自动构建 + Watchtower 自动拉取
- 前端变更需 `vite build` + `go build` + 重启服务
