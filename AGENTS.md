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
│   ├── http/                # chi router + REST API + SSE + SPA fallback
│   └── storage/             # SQLite 持久化（migration + Store + CRUD）
├── frontend/                # React + Vite + Tailwind
│   └── src/
│       ├── api/client.ts    # API 客户端
│       ├── components/      # Modal, Details, SearchableList
│       └── pages/           # 5 页面
```

## 最近操作
- 2026-05-30：前端 SSE 集成 — `useSSE` hook（模块级单例 EventSource）+ 4 页面自动刷新（Matches/TeamDetail/PlayerDetail/News）；构建验证通过
- 2026-05-30：长期化存储全部实现完成 — 7 Group（16 任务）编译通过 + 12 测试套件通过 + 端到端验证（health/SSE/SQLite）；6 次 commit
- 2026-05-29：Group D facade + router 集成 — Type A/B 三层回退、withCacheOrStore、SSE 路由注册
- 2026-05-29：Group B storage 包 — 6 文件（migration + Store + 4 CRUD）
- 2026-05-29：Group C SSE hub — SSEHub + SSEHandler
- 2026-05-29：赛程覆盖面修复 — `.match-wrapper` + `data-match-id` 去重

## 关键发现

### HLTV HTML 选择器（核心参考）
- **选手页**: `.playerNickname` / `.playerRealname` / `.playerTeam a[itemprop="text"]` / `.player-stat` > `.statsVal p b`(能力值) / `.stats-window`(maps数) / `.playerpage-matchbox`(近期比赛) / `.playerpage-match-result`(比分) / `.playerpage-match-date` / `.majorWinner b`(Major冠军数) / `.mvp-count`(MVP数) / `.all-time-stat` > `.stat` + `.description`(生涯战斗统计，旧版) / `.highlighted-stat` > `.stat` + `.description`(生涯概览，新版通用)
- **队伍页**: `h1.profile-team-name` / `.value.h-rank` / `.bodyshot-team a[href*='/player/']`(队员) / `.trophySection .trophyDescription[title]` / `.highlighted-stat`(胜率/连胜)
- **比赛链接**: `.playerpage-matchbox[href]` 正则 `/stats/matches/(\d+)/([^"\s]+)`
- **赛果**: `.result-con` > `.line-align.team1 .team` / `.result-score` / `.event-name`
- **赛程**: `.matches-list-section` > `.match-wrapper`(比赛容器，每场比赛唯一) > `.match`(可能嵌套两层) / `.match-team.team1/team2 .match-teamname`(队名) / `.match-event`(赛事名) / `.match-info`(时间/boN) / `.match-no-info`(无队伍时的占位描述)；`data-match-id` 属性获取比赛 ID；`.match-wrapper` 的 `team1`/`team2` 属性在队伍未定时为空
- **新闻**: `.newstext` / `.news-block p`(正文) — 不可用 `.Text()` 取整个容器
- **搜索**: `table tbody tr > a[href*='/player/']` 正则 `/player/(\d+)/(.+)`

### CF 分层
- **HTTP 直连可用**：`/player/`、`/team/`、`/search`、`/news/`
- **Firecrawl 回退**：`/matches`（HTTP 403 时自动回退，需 `FIRECRAWL_API_KEY`）
- **被 Cloudflare 封锁 (403)**：`/matches`、`/results`、`/`

### 三层回退 (Cache -> SQLite -> HLTV)
- **Type A**（player/team/news article detail）：`GetXxxCached` 方法内联三层逻辑，Tier 2 命中后后台 goroutine `refreshXxx` 更新缓存，调用 `broadcast` 推送 SSE 事件
- **Type B**（matches/events/news lists）：通过 `withCacheOrStore` 方法，Tier 1 检查缓存（含 stale），Tier 2 查询 SQLite（命中则回缓存 + 后台 `RunOnce` 刷新），Tier 3 直接爬取并存库
- `scrapeXxx` 辅助方法执行实际抓取并写入 SQLite（nil-store 安全）
- `store *storage.Store` 为 nil 时自动降级为 Cache-only 模式
- `notify` 回调桥接 facade -> SSEHub.Broadcast，用于前端实时刷新

### 缓存模式
- `PlayerDetail`/`TeamDetail`/`NewsArticle` 走三层回退 Type A，`withCacheOrStore` 用于 Type B
- `sync/atomic.Int64` 计数器，与 `sync.RWMutex` 无锁竞争

### nickname 覆盖层
- `internal/localization/overrides.go`：线程安全内存缓存 + JSON 持久化
- 空值语义 = 删除条目；写操作先更新内存再写磁盘
- API：`PUT /api/nicknames/team` 尝试解析 canonical（目录内队伍），否则直接按原始名称存储；`PUT /api/nicknames/player` 直接存储
- `BuildFullDict` 对目录队伍做别名展开 + 所有 override 添加直接 key-value 映射（确保非目录队伍昵称出现在赛程页面）

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
- `docker pull + run` 部署新镜像（需要 FIRECRAWL_API_KEY）
- 考虑为 /results 页面也添加 Firecrawl 回退

## 进行中
- 无

