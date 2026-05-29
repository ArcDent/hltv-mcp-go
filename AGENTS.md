# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈
- 目标：Go 单二进制，MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, React 18, Vite, Tailwind CSS v4
- 远端仓库：[ArcDent/HLTV-data](https://github.com/ArcDent/HLTV-data)

## 项目静态结构
```
├── main.go                  # MCP stdio + HTTP :8082 双 goroutine
├── Dockerfile               # 三阶段：frontend → Go → alpine
├── internal/
│   ├── types/               # 共享类型 + ToolError
│   ├── config/              # 环境变量配置
│   ├── crypto/              # AES-256-GCM 加解密
│   ├── cache/               # TTL + stale + 并发合并
│   ├── client/              # HTTP + Firecrawl 客户端
│   ├── scraper/             # fetchDoc 共享 + 7 爬虫模块
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
│       ├── components/      # Modal, Details, SearchableList
│       └── pages/           # 5 页面
```

## 最近操作
- 2026-05-29：依赖收敛 — 删除 chromedp（115行+依赖）、`internal/errors` 包、4跳死参数链、空目录和过期文件；Docker 基础镜像从 headless-shell → alpine；scraper 提取共享 fetchDoc；normalizer 内联薄包装；ToolError 实现 error 接口
- 2026-05-29：搜索页面切换 bug 修复 — SearchableList 添加 `key={type}`；embed 指令 `dist/*` → `dist`
- 2026-05-29：Firecrawl 集成 — MatchesScraper.GetUpcoming 403 时回退到 Firecrawl；重写 NormalizeUpcomingMatches
- 2026-05-29：HLTV CF 封锁修复 — handler 超时；nil pointer panic 修复

## 关键发现

### HLTV HTML 选择器（核心参考）
- **选手页**: `.playerNickname` / `.playerRealname` / `.playerTeam a[itemprop="text"]` / `.player-stat` > `.statsVal p b`(能力值) / `.stats-window`(maps数) / `.playerpage-matchbox`(近期比赛) / `.playerpage-match-result`(比分) / `.playerpage-match-date` / `.majorWinner b`(Major冠军数) / `.mvp-count`(MVP数) / `.all-time-stat` > `.stat` + `.description`(生涯战斗统计，旧版) / `.highlighted-stat` > `.stat` + `.description`(生涯概览，新版通用)
- **队伍页**: `h1.profile-team-name` / `.value.h-rank` / `.bodyshot-team a[href*='/player/']`(队员) / `.trophySection .trophyDescription[title]` / `.highlighted-stat`(胜率/连胜)
- **比赛链接**: `.playerpage-matchbox[href]` 正则 `/stats/matches/(\d+)/([^"\s]+)`
- **赛果**: `.result-con` > `.line-align.team1 .team` / `.result-score` / `.event-name`
- **赛程**: `.matches-list-section` > `.matches-list-headline`(日期) + `.match`(比赛)
- **新闻**: `.newstext` / `.news-block p`(正文) — 不可用 `.Text()` 取整个容器
- **搜索**: `table tbody tr > a[href*='/player/']` 正则 `/player/(\d+)/(.+)`

### CF 分层
- **HTTP 直连可用**：`/player/`、`/team/`、`/search`、`/news/`
- **Firecrawl 回退**：`/matches`（HTTP 403 时自动回退，需 `FIRECRAWL_API_KEY`）
- **被 Cloudflare 封锁 (403)**：`/matches`、`/results`、`/`

### 缓存模式
- `PlayerDetail` 不走 `withCache`，直接用 `cache.Get/Set`，key 格式 `player_detail:<id>`
- `sync/atomic.Int64` 计数器，与 `sync.RWMutex` 无锁竞争

### nickname 覆盖层
- `internal/localization/overrides.go`：线程安全内存缓存 + JSON 持久化
- 空值语义 = 删除条目；写操作先更新内存再写磁盘
- API：`PUT /api/nicknames/team` 先解析 canonical 名；`PUT /api/nicknames/player` 直接存储

### React Router 路由切换
- 不同路由渲染同一组件类型时，React reconciliation 复用实例不重新挂载，内部 state 保留
- 修复方式：给组件添加 `key` 区分不同路由实例

### Go embed + Vite
- `//go:embed dist` 递归包含整个 dist 目录（含 `dist/assets/`）
- Vite build 将 JS/CSS 放在 `dist/assets/`，index.html 引用它们

### 错误处理
- `ToolError` 直接实现 `error` 接口，删除独立的 `internal/errors` 包
- 错误创建直接 `&types.ToolError{Code: "...", Message: "..."}`

### 部署
- Docker 三阶段构建 → `ghcr.io/arcdent/hltv-data:latest`（alpine 基础镜像，~15MB binary）
- CI/CD：push main → GitHub Actions 自动构建
- 前端变更需 `vite build` + `go build` + 重启服务

## 下一步
- 考虑为 /results 页面也添加 Firecrawl 回退
- 监控 Firecrawl API 配额消耗

## 进行中
- 无（2026-05-29 依赖收敛已完成）

