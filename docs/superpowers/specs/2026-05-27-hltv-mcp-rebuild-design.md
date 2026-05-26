# HLTV MCP Go 全栈重建设计

## 目标

将现有 TypeScript + Python 双层架构的 HLTV MCP 服务，重建为 Go 单一二进制全栈应用，
保持全部 10 个 MCP 工具功能不变，并增加基于 React 的 Web 管理面板和数据浏览器。

## 核心决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 语言 | Go | 单二进制部署、goroutine 天然适配 MCP stdio + HTTP 并发、embed 内嵌前端、低内存快启动 |
| MCP SDK | mark3labs/mcp-go | 社区最活跃 Go MCP SDK，API 风格接近原 TS SDK |
| HTTP 路由 | chi | 标准库兼容、middleware 链清晰、URL 参数提取优于 net/http |
| 前端 | React + Vite | 交互丰富的管理面板，同语言共享类型不可用（Go），但 SPA 体验好 |
| 前后端通信 | REST only | 管理面板数据变化非高频，轮询足够；YAGNI |
| 爬虫 | net/http 优先 / chromedp fallback | 轻量优先，反爬时自动降级 |
| 前端 CSS | Tailwind CSS | 原子类组件内联，Vite 热更新，shadcn/ui 等组件库默认支持 |

## 架构

```
┌─────────────────────────────────────────────────────┐
│  hltv-mcp (single Go binary)                        │
│                                                      │
│  ┌──────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │  stdio   │  │  HTTP (:8082)│  │  //go:embed   │  │
│  │  MCP     │  │  chi router  │  │  dist/ (React) │  │
│  │  Server  │  │              │  │               │  │
│  └────┬─────┘  └──────┬───────┘  └───────────────┘  │
│       │               │                              │
│       └───────┬───────┘                              │
│               │                                      │
│       ┌───────┴───────┐                              │
│       │   HltvFacade  │  编排层                      │
│       └───────┬───────┘                              │
│               │                                      │
│       ┌───────┴───────┐                              │
│       │  HltvClient   │  爬虫客户端                  │
│       │  (net/http    │  chromedp fallback            │
│       │   + chromedp) │                              │
│       └───────┬───────┘                              │
│               │                                      │
│       ┌───────┴───────┐                              │
│       │  MemoryCache  │  TTL + stale 窗口            │
│       └───────────────┘                              │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │  React App (embedded via //go:embed)           │  │
│  │  ├── Dashboard (状态/缓存/爬虫)                │  │
│  │  ├── Teams (搜索+详情)                         │  │
│  │  ├── Players (搜索+详情)                       │  │
│  │  ├── Matches (今日/赛程/赛果)                  │  │
│  │  ├── News (实时/归档)                          │  │
│  │  └── Cache Management                          │  │
│  └────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

Go 进程启动后同时运行两个 goroutine：
- MCP stdio server（通过 `mark3labs/mcp-go` 注册 10 个 tool，stdio transport）
- HTTP server（`chi` router，端口可配，默认 8082），同时 serve REST API 和嵌入的 React 前端

所有组件共享同一个内存实例：`HltvFacade`、`MemoryCache`、`HltvClient`。

## 目录结构

```
hltv-mcp-fully-rebuild/
├── main.go                    # 入口：解析 flag、启动 MCP goroutine + HTTP goroutine
├── go.mod / go.sum
│
├── internal/
│   ├── config/
│   │   └── config.go          # 环境变量加载 → Config struct
│   │
│   ├── mcp/
│   │   ├── server.go          # 10 个工具注册及 handler
│   │   └── transport.go       # stdio 传输启动
│   │
│   ├── http/
│   │   ├── router.go          # chi 路由（API + SPA fallback）
│   │   ├── middleware.go      # CORS / 日志 / recovery
│   │   └── handlers/
│   │       ├── status.go      # GET  /api/status
│   │       ├── cache.go       # GET|DELETE /api/cache
│   │       ├── search.go      # GET  /api/search
│   │       ├── matches.go     # GET  /api/matches、/api/matches/today、/api/results
│   │       └── news.go        # GET  /api/news/realtime、/api/news
│   │
│   ├── facade/
│   │   ├── facade.go          # HltvFacade struct + withCache 通用包装
│   │   ├── resolve.go         # resolveTeam / resolvePlayer
│   │   ├── matches.go         # getTodayMatches / getUpcomingMatches / getResultsRecent
│   │   └── news.go            # getRealtimeNews / getNewsDigest
│   │
│   ├── client/
│   │   ├── client.go          # HTTP 请求（多 baseUrl 重试）
│   │   └── chromedp.go        # chromedp fallback 抓取
│   │
│   ├── scraper/
│   │   ├── team.go            # 队伍搜索 + 详情
│   │   ├── player.go          # 选手搜索 + 详情 + 统计概览
│   │   ├── results.go         # 近期赛果
│   │   ├── matches.go         # 未来赛程
│   │   ├── news.go            # 归档新闻
│   │   └── realtime_news.go   # 实时新闻
│   │
│   ├── cache/
│   │   └── cache.go           # 内存缓存（TTL + stale 窗口 + 并发合并）
│   │
│   ├── normalizer/
│   │   ├── match.go           # NormalizedMatch 标准化
│   │   ├── team.go            # TeamProfile 标准化
│   │   ├── player.go          # PlayerProfile 标准化
│   │   └── news.go            # NewsItem 标准化
│   │
│   ├── renderer/
│   │   └── chinese.go         # 中文渲染（MCP 工具文本输出）
│   │
│   ├── summary/
│   │   └── summary.go         # 中文摘要（template/raw 模式）
│   │
│   ├── localization/
│   │   ├── catalog.go         # 70+ 队伍中英文映射
│   │   └── events.go          # 20+ 赛事中英文映射
│   │
│   ├── types/
│   │   └── types.go           # 全部共享类型定义
│   │
│   └── errors/
│       └── errors.go          # AppError + 错误码
│
├── frontend/                   # React + Vite（独立子项目）
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── index.html
│   └── src/
│       ├── App.tsx
│       ├── main.tsx
│       ├── api/               # REST API 调用封装（fetch wrapper）
│       ├── pages/
│       │   ├── Dashboard.tsx   # 服务状态 + 缓存统计 + 爬虫状态
│       │   ├── Matches.tsx     # 今日/未来赛程/近期赛果 + 筛选
│       │   ├── Teams.tsx       # 队伍搜索 + 详情
│       │   ├── Players.tsx     # 选手搜索 + 详情
│       │   ├── News.tsx        # 实时新闻 / 归档新闻
│       │   └── Cache.tsx       # 缓存条目浏览 + 清除
│       └── components/         # 共享 UI 组件
│
└── dist/                       # 前端构建产物（gitignore，go:embed 嵌入）
```

## REST API

所有端点返回统一 JSON 格式：

```json
{
  "data": {},
  "meta": { "fetched_at": "...", "cache_hit": true, "timezone": "Asia/Shanghai" },
  "error": null
}
```

### 端点列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 `{"status":"ok"}` |
| GET | `/api/status` | 服务状态（uptime, version, cache_entries, scraper_state） |
| GET | `/api/cache` | 缓存统计（entries, hits, misses, stale_hits, memory_approx） |
| DELETE | `/api/cache` | 清除全部缓存 |
| GET | `/api/search?q=&type=team\|player` | 实体搜索 |
| GET | `/api/teams/:id` | 队伍详情（profile + recent_results + upcoming_matches） |
| GET | `/api/players/:id` | 选手详情（profile + overview + recent_highlights） |
| GET | `/api/matches?team_id=&team=&event=&days=&limit=` | 未来赛程 |
| GET | `/api/matches/today` | 今日赛程 |
| GET | `/api/results?team_id=&team=&event=&days=&limit=` | 近期赛果 |
| GET | `/api/news/realtime?limit=&offset=` | 实时新闻 |
| GET | `/api/news?year=&month=&tag=&limit=&offset=` | 归档新闻 |

### SPA 路由

所有非 `/api/*` 的 GET 请求返回 `index.html`，由 React Router 处理客户端路由。

## MCP 工具（10 个，与原项目功能等价）

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `resolve_team` | 队伍名解析为 HLTV 实体候选 | `name`(必填), `exact?`, `limit?` |
| `resolve_player` | 选手名解析为 HLTV 实体候选 | `name`(必填), `exact?`, `limit?` |
| `hltv_team_recent` | 队伍近况（赛果 + 赛程 + 统计） | `team_id?`, `team_name?`, `limit?`, `include_upcoming?`, `include_recent_results?`, `detail?`, `exact?` |
| `hltv_player_recent` | 选手近况（统计 + 亮点） | `player_id?`, `player_name?`, `limit?`, `detail?`, `exact?` |
| `hltv_results_recent` | 近期赛果 | `team_id?`, `team?`, `event?`, `limit?`, `days?` |
| `hltv_matches_upcoming` | 未来赛程 | `team_id?`, `team?`, `event?`, `limit?`, `days?` |
| `hltv_matches_today` | 今日赛程（UTC+8 日界线） | 无参数 |
| `match_command_parse` | /match 命令参数解析 | `raw_args?` |
| `hltv_realtime_news` | 实时新闻 | `limit?`, `page?`, `offset?` |
| `hltv_news_digest` | 归档新闻 | `limit?`, `tag?`, `year?`, `month?`, `page?`, `offset?` |

### 不可回归的行为约定

- `/match` 裸调用 → `hltv_matches_today({})`；任何非空参数 → 直接拒绝
- `HltvFacade.getTodayMatches()` 委托为 `getUpcomingMatches({})`
- 泛化占位符自动剥离（"today matches"、"今日赛程"、`x`、`y`、`n/a` 等）
- 时区固定 `Asia/Shanghai`，所有时间计算、日界线判断使用该时区
- 缓存 stale 窗口兜底：上游请求失败时返回过期缓存而非直接报错

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MCP_SERVER_NAME` | `hltv-mcp-service` | MCP server name |
| `MCP_SERVER_VERSION` | `1.0.0` | MCP server version |
| `HTTP_PORT` | `8082` | HTTP 监听端口 |
| `HTTP_HOST` | `0.0.0.0` | HTTP 监听地址 |
| `HLTV_DATA_SOURCE` | `auto` | 爬虫数据源：`auto`（先直接 HTTP 再 chromedp）、`direct`（仅 HTTP）、`chromedp`（仅 chromedp） |
| `HLTV_CHROME_PATH` | 空（自动查找） | chromedp 使用的 Chrome/Chromium 可执行文件路径 |
| `HLTV_HTTP_TIMEOUT_MS` | `8000` | 单次 HTTP 请求超时（毫秒） |
| `HLTV_RETRY_COUNT` | `2` | HTTP 请求重试次数 |
| `CACHE_TTL_ENTITY_SEC` | `3600` | 实体缓存 TTL |
| `CACHE_TTL_TEAM_SEC` | `300` | 队伍近况缓存 TTL |
| `CACHE_TTL_PLAYER_SEC` | `300` | 选手近况缓存 TTL |
| `CACHE_TTL_RESULTS_SEC` | `120` | 赛果缓存 TTL |
| `CACHE_TTL_MATCHES_SEC` | `60` | 赛程缓存 TTL |
| `CACHE_TTL_NEWS_SEC` | `180` | 新闻缓存 TTL |
| `CACHE_TTL_REALTIME_NEWS_SEC` | `60` | 实时新闻缓存 TTL |
| `CACHE_MAX_ENTRIES` | `500` | 最大缓存条目数 |
| `CACHE_STALE_WINDOW_SEC` | `3600` | Stale 缓存保留窗口 |
| `DEFAULT_RESULT_LIMIT` | `5` | 默认查询结果数 |
| `SUMMARY_MODE` | `template` | 摘要模式：`template` 或 `raw` |

> 以下原项目变量在 Go 版本中**移除**（不再需要外部 Python 上游）：
> `HLTV_UPSTREAM_MANAGED`、`HLTV_UPSTREAM_PYTHON_PATH`、`HLTV_UPSTREAM_WORKDIR`、
> `HLTV_UPSTREAM_PORT`、`HLTV_UPSTREAM_HEALTH_PATH`、`HLTV_UPSTREAM_START_TIMEOUT_MS`、
> `HLTV_API_BASE_URL`、`HLTV_API_FALLBACK_BASE_URL`

## 爬虫策略

### 数据源

6 个爬虫模块，对应 HLTV.org 的 6 类页面：

| 模块 | 目标 | 缓存 TTL |
|------|------|----------|
| `scraper/team.go` | 队伍搜索 + 详情 | entity (3600s) |
| `scraper/player.go` | 选手搜索 + 详情 + 统计 | entity (3600s) |
| `scraper/results.go` | 近期赛果列表 | results (120s) |
| `scraper/matches.go` | 未来赛程列表 | matches (60s) |
| `scraper/news.go` | 月度归档新闻 | news (180s) |
| `scraper/realtime_news.go` | 首页实时新闻 feed | realtime_news (60s) |

### 抓取流程

1. 首先尝试 `net/http` 直接请求 HLTV.org 的 HTML 页面
2. 解析 HTML（`goquery`）提取结构化数据
3. 如遇 Cloudflare 拦截（503 / 标题包含 "Just a moment" / 无预期数据），fallback 到 `chromedp`
4. 结果存入缓存，TTL 内不再重复请求

### chromedp fallback 记忆

- 记录每个端点最近一次 HTTP 失败的时间戳
- 同一端点在 **5 分钟内**曾触发 chromedp fallback 的，后续请求直接走 chromedp，跳过 HTTP 尝试
- 5 分钟窗口过后恢复 HTTP 优先，重新探测

### chromedp Chrome 生命周期

- Go 启动时通过 `chromedp.ExecAllocator` 启动一个 headless Chrome 子进程（`chrome-headless-shell` 或系统 Chrome）
- 该 Chrome 实例作为常驻后台进程，chromedp 所有请求复用它
- 进程退出时 chromedp 负责关闭 Chrome 子进程

### 手动编译时的 Chrome 依赖

- Go 启动时检查 Chrome/Chromium 是否可用（`HLTV_CHROME_PATH` 配置路径或系统 PATH 自动查找）
- 如 chromedp 不可用：打印 warning 日志，自动降级为 `direct` 模式（`HLTV_DATA_SOURCE=direct`）
- 降级后 HTTP 请求失败直接返回错误，不再尝试 chromedp
- `GET /api/status` 的 `scraper_state` 字段反映当前降级状态

### 不做预抓取

不实现定时预抓取或启动预热——全部按需触发。HLTV MCP 是本地工具，QPS 极低。

## 前端

### 技术栈

- React 18+ / TypeScript
- Vite（构建工具）
- Tailwind CSS（样式）
- React Router（客户端路由）
- 图表库待定（Dashboard 需要缓存命中率趋势图时再确定）

### 页面

| 页面 | 路由 | 功能 |
|------|------|------|
| Dashboard | `/` | 服务状态卡片（uptime、Go 版本、内存）、缓存统计（条目数、命中率、stale 数）、爬虫状态、清除缓存按钮 |
| Matches | `/matches` | 标签切换（今日/未来赛程/近期赛果），队伍名、赛事名、天数筛选，列表展示 |
| Teams | `/teams` | 搜索框（英文名/中文名/昵称），结果列表 → 详情页（排名、国家、赛果表格、战绩 W/L/D） |
| Players | `/players` | 同上搜索 → 详情页（队伍、国家、统计指标、近期亮点） |
| News | `/news` | 标签切换（实时/归档），归档可选年月标签，分页浏览 |
| Cache | `/cache` | 缓存条目表格（key 预览、类型、过期时间），清除全部缓存按钮 |

## 构建与部署

### 手动编译

```bash
cd frontend && npm install && npm run build    # → ../dist/
go build -o hltv-mcp .                          # → hltv-mcp 二进制（内嵌 dist/）
./hltv-mcp                                       # 启动，MCP stdio + HTTP :8082
```

### Docker 多阶段构建

```dockerfile
# Stage 1: 构建前端
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/ .
RUN npm ci && npm run build

# Stage 2: 编译 Go
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./dist/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o hltv-mcp .

# Stage 3: 运行时
# chrome-headless-shell:stable 提供 chrome-headless-shell 二进制
# Go 通过 chromedp.ExecAllocator 启动该二进制作为 headless Chrome 实例
FROM chrome-headless-shell:stable
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/usr/bin/chrome-headless-shell
ENTRYPOINT ["/hltv-mcp"]
```

最终镜像大小约 300MB（主要是 Chromium），一键启动即可。

### Docker 运行

```bash
docker build -t hltv-mcp .
docker run --rm -p 8082:8082 hltv-mcp
# 前端: http://localhost:8082
# MCP: 通过 stdio 注册到 OpenCode
```

## 测试策略

| 包 | 测试内容 |
|----|----------|
| `internal/cache/` | TTL 过期、stale 窗口保留、并发合并、容量上限 FIFO 淘汰 |
| `internal/normalizer/` | 上游 HTML → NormalizedMatch/TeamProfile/PlayerProfile/NewsItem 转换 |
| `internal/facade/` | 查询标准化、占位符剥离、排序、时间窗口过滤 |
| `internal/localization/` | 名称映射覆盖率（70+ 队伍 + 20+ 赛事） |
| `internal/client/` | HTTP 重试逻辑、chromedp fallback 触发条件 |
| `internal/mcp/` | 工具注册正确性、schema 参数验证 |
| 前端 | Vitest + React Testing Library（后续补充，不作为 MVP 交付条件） |

Docker 验证：
```bash
docker build -t hltv-mcp .
docker run --rm -p 8082:8082 hltv-mcp &
curl http://localhost:8082/api/health   # → {"status":"ok"}
```

## 与原项目的差异

| 方面 | 原项目 | 新项目 |
|------|--------|--------|
| MCP 服务 | TypeScript + MCP SDK | Go + mark3labs/mcp-go |
| 爬虫 | Python Flask + Scrapy（独立进程） | Go net/http + chromedp（进程内） |
| 上游管理 | managed/external 两模式 | 无外部依赖 |
| 传输层 | 仅 stdio | stdio + HTTP REST |
| 前端 | 无（依赖 OpenCode UI） | React SPA 管理面板 + 数据浏览器 |
| 部署 | npm + Python venv | 单二进制 或 Docker 一键 |
| MCP 工具 | 10 个 | 10 个（功能等价） |
| 中文输出 | ChineseRenderer + SummaryService | 同逻辑 Go 实现 |
| 名称本地化 | 70+ 队伍 + 20+ 赛事 | 同数据 Go 实现 |
| 时区 | Asia/Shanghai | 同 |
| 缓存 | MemoryCache（TS） | 同逻辑 Go 实现 |
