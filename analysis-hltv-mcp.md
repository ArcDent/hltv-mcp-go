# HLTV MCP Service 源码分析

## 项目概述

**hltv-mcp-service** (v0.3.0) 是一个面向 OpenCode 的本地 MCP（Model Context Protocol）stdio 服务，封装了 HLTV（Counter-Strike 电竞赛事数据平台）的搜索、赛程、赛果、新闻和选手/队伍数据查询能力。项目采用 **TypeScript MCP 服务器 + Python Flask 上游爬虫 API** 双层架构。

---

## 技术栈

### 前端层（MCP 服务）

| 组件 | 技术 | 版本 |
|------|------|------|
| 运行时 | Node.js | >=18.17 |
| 语言 | TypeScript | 5.7.2 |
| 模块系统 | ESM (`"type": "module"`) | - |
| 编译目标 | ES2022 / NodeNext | - |
| MCP SDK | @modelcontextprotocol/sdk | ^1.12.0 |
| 验证 | Zod | ^3.23.8 |
| 开发执行 | tsx | ^4.19.2 |
| 构建 | tsc | - |
| 测试 | Node.js native test runner + tsx | - |
| 包管理 | npm | - |

### 后端层（上游爬虫 API）

| 组件 | 技术 |
|------|------|
| 语言 | Python |
| Web 框架 | Flask (>=3.1.1) |
| API 文档 | Flasgger (Swagger) |
| 爬虫引擎 | Scrapy (基于 hltv_scraper) |
| 限流 | Flask-Limiter |
| 测试 | pytest |
| 构建 | Makefile |

---

## 目录结构

```
hltv-mcp/
├── package.json              # Node 项目配置
├── tsconfig.json             # TypeScript 编译配置
├── AGENTS.md                 # AI Agent 交接文件
├── README.md                 # 项目文档（中文）
│
├── src/                      # TypeScript MCP 服务源码
│   ├── index.ts              # 唯一运行时入口
│   ├── cache/
│   │   ├── memoryCache.ts        # 内存缓存（容量上限 + stale 窗口 + 并发合并）
│   │   └── memoryCache.test.ts
│   ├── clients/
│   │   └── hltvApiClient.ts      # 上游 HTTP 客户端（多 baseUrl 重试/故障切换）
│   ├── commands/
│   │   └── commandHandlers.ts    # 命令处理器（仅类型导出）
│   ├── config/
│   │   ├── env.ts                # 环境变量加载和配置构建（核心配置模块）
│   │   └── env.test.ts
│   ├── doctor/
│   │   ├── opencodeDoctor.ts         # OpenCode 诊断工具
│   │   ├── opencodeDoctorCli.ts      # CLI 诊断入口
│   │   └── *.test.ts
│   ├── errors/
│   │   └── appError.ts          # 统一应用错误类型（11种错误码）
│   ├── mcp/
│   │   ├── schemas.ts           # Zod 工具参数 schema 定义
│   │   └── server.ts            # MCP 工具注册 + stdio 传输
│   ├── renderers/
│   │   └── chineseRenderer.ts   # 中文渲染输出（格式化+原因说明+分页信息）
│   ├── resolvers/
│   │   ├── entityIdentity.ts    # 实体身份核心（名称标准化、别名目录、查询变体生成）
│   │   ├── teamResolver.ts      # 队伍解析器
│   │   ├── playerResolver.ts    # 选手解析器（含别名词典）
│   │   └── teamResolver.test.ts
│   ├── services/
│   │   ├── hltvFacade.ts        # 核心编排层（缓存使用、查询标准化、工具行为）
│   │   ├── hltvNormalizer.ts    # 数据标准化（上游原始 JSON → 规范化结构）
│   │   ├── matchCommandParser.ts    # /match 命令参数解析
│   │   ├── summaryService.ts    # 中文自然语言摘要生成
│   │   ├── upcomingMatchesQuery.ts  # 未来赛程查询过滤（去除兜底占位符）
│   │   └── *.test.ts
│   ├── types/
│   │   ├── common.ts            # 通用类型（ToolResponse, ToolMeta, ToolError）
│   │   └── hltv.ts              # HLTV 域类型（实体、比赛、新闻、查询接口）
│   ├── upstream/
│   │   ├── managedUpstream.ts   # 受管上游启动/停止管理
│   │   ├── healthcheck.ts       # 上游健康检查（轮询 + token 验证）
│   │   ├── port.ts              # 端口可用性检查 + baseUrl 构建
│   │   ├── processManager.ts    # 子进程生成管理
│   │   ├── pythonLocator.ts     # Python 解释器路径解析
│   │   ├── startupError.ts      # 启动错误类型
│   │   ├── types.ts             # 上游类型定义
│   │   └── *.test.ts
│   ├── utils/
│   │   ├── localizedNames.ts    # 本地化名称系统（70+队伍+20+赛事中英文映射）
│   │   ├── teamAliasCatalog.ts  # 队伍别名目录（26支重点队伍的规范化别名）
│   │   ├── strings.ts           # 字符串工具（slugify, 字符清理, HTML实体链接解析）
│   │   ├── object.ts            # 对象工具（安全字段提取、数组包装）
│   │   ├── time.ts              # 时间工具（固定 Asia/Shanghai 时区）
│   │   └── *.test.ts
│   └── matchCommandFlow.test.ts   # 端到端流程测试（/match 命令）
│
├── hltv-api-fixed/            # Python 上游爬虫 API（独立子项目）
│   ├── app.py                    # Flask 应用入口
│   ├── config.py                 # 应用配置
│   ├── Makefile                  # 独立构建/测试命令
│   ├── requirements.txt          # Python 依赖
│   ├── Dockerfile
│   ├── routes/                   # Flask 蓝图路由
│   │   ├── teams.py              # /api/v1/teams/*
│   │   ├── players.py            # /api/v1/players/*
│   │   ├── matches.py            # /api/v1/matches/*
│   │   ├── results.py            # /api/v1/results/*
│   │   └── news.py               # /api/v1/news/*
│   ├── hltv_scraper/             # Scrapy 爬虫核心
│   │   └── hltv_scraper/spiders/ # 14个爬虫（队伍/选手/比赛/新闻/实时新闻等）
│   ├── tests/                    # pytest 测试
│   └── env/                      # Python 虚拟环境（默认路径）
│
├── scripts/
│   └── run-tests.mjs             # 测试运行脚本
├── docs/
│   ├── templates/                # OpenCode 命令模板（team/player/result/match/news）
│   └── superpowers/              # 设计文档和计划
└── examples/
    └── opencode-project/         # OpenCode 完整示例配置
```

---

## 系统架构

### 双层架构图

```
┌──────────────────────────────────────────────────────┐
│  OpenCode / MCP Client (stdio)                       │
│  ┌──────────────────────────────────────────────┐    │
│  │  Slash Commands: /team /player /match /news  │    │
│  │  MCP Tools: hltv_local_*                     │    │
│  └──────────────────────────────────────────────┘    │
└──────────────────────┬───────────────────────────────┘
                       │ stdio (JSON-RPC)
┌──────────────────────┴───────────────────────────────┐
│  TypeScript MCP Server (src/)                        │
│  ┌─────────────────────────────────────────────┐     │
│  │  index.ts                                    │     │
│  │  ├── loadConfig()      配置加载              │     │
│  │  ├── startManagedUpstream()  上游自动启动    │     │
│  │  ├── HltvApiClient     HTTP 客户端           │     │
│  │  ├── TeamResolver / PlayerResolver  解析器   │     │
│  │  ├── MemoryCache       内存缓存              │     │
│  │  ├── HltvFacade        业务编排层            │     │
│  │  ├── ChineseRenderer   中文渲染              │     │
│  │  └── McpServer (stdio transport)             │     │
│  └─────────────────────────────────────────────┘     │
└──────────────────────┬───────────────────────────────┘
                       │ HTTP (localhost:18020)
┌──────────────────────┴───────────────────────────────┐
│  Python Flask API (hltv-api-fixed/)                   │
│  ┌─────────────────────────────────────────────┐     │
│  │  app.py (Flask + Swagger)                    │     │
│  │  ├── /api/v1/teams/*     队伍搜索/详情/比赛  │     │
│  │  ├── /api/v1/players/*   选手搜索/详情/统计  │     │
│  │  ├── /api/v1/matches/*   未来赛程            │     │
│  │  ├── /api/v1/results/*   近期赛果            │     │
│  │  ├── /api/v1/news/*      归档新闻/实时新闻   │     │
│  │  └── /healthz            健康检查            │     │
│  └─────────────────────────────────────────────┘     │
└──────────────────────┬───────────────────────────────┘
                       │ Scrapy 爬虫
┌──────────────────────┴───────────────────────────────┐
│  HLTV.org (CS 电竞数据平台)                           │
└──────────────────────────────────────────────────────┘
```

### 关键数据流

```
1. MCP 客户端调用 hltv_local_resolve_team({name: "Spirit"})
2. → McpServer 路由到 server.tool("resolve_team", ...)
3. → HltvFacade.resolveTeam(query)
4. → 检查 MemoryCache.get("resolve_team:{...}")
5. → 缓存未命中时调用 TeamResolver.resolve(name)
6. → TeamResolver 构建查询变体（别名展开）
7. → HltvApiClient.searchTeams(query) 发送 HTTP GET
8. → Python Flask API /api/v1/teams/search/:name
9. → HLTVScraper 执行 Scrapy 爬虫爬取 HLTV.org
10. → 返回 JSON → TS 端 normalize 为 ResolvedTeamEntity[]
11. → 结果写入 Cache 并返回给 MCP Client
```

---

## 核心模块详解

### 1. 入口 — `src/index.ts`

**唯一运行时入口**，负责：

1. **配置加载**：`loadConfig()` 从环境变量构建 `AppConfig`
2. **受管上游启动**（默认启用）：`startManagedUpstream()` 在 MCP 启动前拉起 `hltv-api-fixed/app.py` Python 进程
3. **依赖注入组装**：
   ```
   MemoryCache → HltvApiClient → TeamResolver/PlayerResolver
   → HltvFacade → SummaryService → ChineseRenderer
   → createMcpServer → startMcpServer (stdio)
   ```
4. **优雅关闭**：注册 `beforeExit`/`SIGINT`/`SIGTERM`/`stdin end` 信号处理器

**关键行为**：
- Python 解释器路径默认 `hltv-api-fixed/env/bin/python`，缺失则 fail fast
- 通过 `AbortController` 实现 startup 取消信号
- `stopManagedUpstreamOnce()` 使用 Promise 缓存，确保仅执行一次

### 2. 配置系统 — `src/config/env.ts`

**环境变量驱动的配置工厂**。

支持环境变量（含默认值）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HLTV_UPSTREAM_MANAGED` | `true` | 是否自动启动上游 |
| `HLTV_UPSTREAM_PORT` | `18020` | 上游端口 |
| `HLTV_UPSTREAM_PYTHON_PATH` | `hltv-api-fixed/env/bin/python` | Python 路径 |
| `HLTV_UPSTREAM_START_TIMEOUT_MS` | `15000` | 启动超时 |
| `HLTV_API_BASE_URL` | `http://127.0.0.1:8020` | 上游 API 基础 URL |
| `HLTV_API_FALLBACK_BASE_URL` | - | 备用 URL |
| `HLTV_API_TIMEOUT_MS` | `8000` | HTTP 超时 |
| `DEFAULT_RESULT_LIMIT` | `5` | 默认结果数 |
| `SUMMARY_MODE` | `template` | 摘要模式 |

**WSL 特殊处理**：
- 在 WSL 环境下运行时，自动检测 `/etc/resolv.conf` 中的 nameserver IP
- 当 `HLTV_API_BASE_URL` 是回环地址时，自动添加宿主机 IP 作为备用 URL

**缓存 TTL 配置**（秒）：
- 实体缓存: 3600, 队伍近况: 300, 选手近况: 300
- 赛果: 120, 赛程: 60, 新闻: 180, 实时新闻: 60

### 3. MCP 服务器 — `src/mcp/server.ts`

**工具注册和传输层**。

注册 10 个 MCP 工具：

| 工具名 | 功能 | Schema |
|--------|------|--------|
| `resolve_team` | 队伍名 → 实体候选 | `resolveEntitySchema` |
| `resolve_player` | 选手名 → 实体候选 | `resolveEntitySchema` |
| `hltv_team_recent` | 队伍近况（赛果+赛程+统计） | `teamRecentSchema` |
| `hltv_player_recent` | 选手近况（统计+亮点+比赛） | `playerRecentSchema` |
| `hltv_results_recent` | 近期赛果（可选队伍/赛事过滤） | `resultsSchema` |
| `hltv_matches_upcoming` | 未来赛程（可选过滤） | `matchesSchema` |
| `hltv_matches_today` | 今日赛程（无参数） | `matchesTodaySchema`（空） |
| `match_command_parse` | /match 参数解析 | `matchCommandParseSchema` |
| `hltv_realtime_news` | 实时新闻（分页） | `realtimeNewsSchema` |
| `hltv_news_digest` | 归档新闻（按年/月/标签） | `newsSchema` |

**传输层**：仅实现 `StdioServerTransport`，Streamable HTTP 和 SSE 未实现。

### 4. 业务编排 — `src/services/hltvFacade.ts`

**核心业务逻辑层**（~1560 行），包含：

#### 缓存策略
- 使用 `withCache()` 通用包装器
- 两级缓存命中：**精确命中**（未过期）→ **stale 回退**（过期但在保留窗口内）
- 并发请求合并：`cache.runOnce()` 确保相同 key 的并发 miss 只发起一次上游请求

#### 实体解析
- **ID-first 查询**：传了 `team_id`/`player_id` 时优先按 ID 解析 canonical slug
- **Slug 候选生成**：从搜索结果链接、名称 slugify、实体链接解析多个候选 slug，逐个尝试匹配
- **失败安全**：`resolveTeamById()` 和 `resolvePlayerById()` 尝试多个 slug 候选，逐个请求直到成功

#### 选手所属队伍推断
- **优先队列表**（PRIORITY_TEAM_QUERIES）：19 支关注的队伍
- **回退扫描**：从近期赛果和未来赛程提取活跃队伍，按出现频率排序
- **并发 workers**：可配置并发数（默认 6），最短路径优先
- **Roster 匹配**：从队伍详情提取成员列表，与选手别名做模糊匹配

#### 查询标准化
- `isLikelyAutofilledUpcomingQuery()` 检测 LLM 自动填充的占位值
- 泛化关键词剥离："today matches"、"今日赛程" 等
- 占位符检测：`x`, `y`, `z`, `?`, `-`, `n/a`, `null` 等

#### 时间窗口过滤
- 固定时区 `Asia/Shanghai`
- 赛果按 `played_at` 从新到旧排序
- 赛程按 `scheduled_at` 从早到晚排序
- 缺少时间的记录排在最后

### 5. 上游 API 客户端 — `src/clients/hltvApiClient.ts`

**HTTP 请求层**，封装对 Python Flask API 的调用：

- **多 baseUrl 重试**：支持 primary + fallback 列表
- **智能路由**：记住最近成功的 baseUrl 索引，优先使用
- **超时控制**：使用 `AbortController` + `setTimeout`
- **错误分类**：
  - 404 → `UPSTREAM_NOT_FOUND`（不重试）
  - 5xx → `UPSTREAM_UNAVAILABLE`（可重试，切换到备用 URL）
  - 超时 → `UPSTREAM_TIMEOUT`（可重试）
- **URL 解析**：正确处理 path prefix（如 `/some-prefix/`）

API 端点映射：

| 方法 | 端点 |
|------|------|
| `searchTeams` | `GET /api/v1/teams/search/:name` |
| `getTeam` | `GET /api/v1/teams/:id/:slug` |
| `getTeamMatches` | `GET /api/v1/teams/:id/matches[/:offset]` |
| `searchPlayers` | `GET /api/v1/players/search/:name` |
| `getPlayer` | `GET /api/v1/players/:id/:slug` |
| `getPlayerOverview` | `GET /api/v1/players/stats/overview/:id/:slug` |
| `getRecentResults` | `GET /api/v1/results/` |
| `getUpcomingMatches` | `GET /api/v1/matches/upcoming` |
| `getNews` | `GET /api/v1/news[/:year/:month]` |
| `getRealtimeNews` | `GET /api/v1/news/realtime` |

### 6. 内存缓存 — `src/cache/memoryCache.ts`

**轻量级内存缓存**，关键特性：

- **容量上限**：默认 500 条目，FIFO 淘汰
- **TTL 过期**：每个条目独立过期时间
- **Stale 窗口**：过期后仍保留一段时间（默认 3600s），过期窗口内可提供 stale 数据兜底
- **并发合并**：`runOnce()` 实现相同 key 的并发 miss 合并为单次计算
- **双层读取**：
  - `get()` → 仅在未过期时返回
  - `getStale()` → 在 stale 窗口内也返回
  - `getStaleWithMeta()` → 额外返回 `staleAgeSec` 元信息

### 7. 实体解析器 — `src/resolvers/`

#### EntityDirectory（`entityIdentity.ts`）
通用双索引目录：
- `byId`: `Map<number, T>` — 按 ID 查找
- `aliasIds`: `Map<string, Set<number>>` — 按别名查找

关键工具函数：
- `normalizeLookupName()`: 多级名称标准化（`strict` / `loose` / `slug` / `tokens`）
- `buildQueryVariants()`: 从别名词典生成搜索查询变体
- `buildSlugCandidates()`: 从多个来源生成 slug 候选项

#### TeamResolver 和 PlayerResolver
- 均继承相同的 resolve 模式：先查本地缓存 → 多搜索词变体并发查询上游 → 评分排序
- 支持 `exact` 精确匹配过滤
- 记忆化：`remember()` 将解析结果写入目录供后续使用

### 8. 数据标准化 — `src/services/hltvNormalizer.ts`

**上游原始 JSON → 规范化类型**的转换层（~740 行）。

输出标准化类型：

```
NormalizedMatch    — 比赛（含 team1/2_id, opponent, event, result, score, 时间等）
TeamProfile        — 队伍档案（id, name, slug, country, rank）
PlayerProfile      — 选手档案（id, name, slug, team, country）
NewsItem           — 新闻条目（title, link, published_at, tag）
RealtimeNewsItem   — 实时新闻条目（section, category, title, relative_time）
```

关键设计：
- **宽容解析**：多种上游字段名映射（如 `team1` / `team1_name` / `team1Name`）
- **字符清理**：处理 HLTV 的编码损坏（波兰/土耳其等特殊字符的 mojibake）
- **嵌套展开**：自动递归展开嵌套的 matches/results 数组
- **Time 标准化**：所有时间输出统一为 ISO 8601，时区固定 `Asia/Shanghai`

### 9. 中文渲染 — `src/renderers/chineseRenderer.ts`

**面向最终用户的中文格式化输出**（~320 行）。

每个渲染方法输出固定格式：
```
【标题】
【关键事实】      ← 结构化事实列表
【中文总结】      ← SummaryService 生成的自然语言摘要
【原因说明】      ← 如有部分数据或异常
【更新时间】
【来源】          ← 含缓存状态
【分页】          ← 新闻类额外包含
```

### 10. 中文摘要 — `src/services/summaryService.ts`

**模板化自然语言摘要生成**（~130 行），支持 `template` 和 `raw` 两种模式：
- `template` 模式生成面向用户的中文自然语言摘要
- `raw` 模式返回占位文本，由 LLM 在最终回答中自行生成摘要

### 11. 本地化名称 — `src/utils/localizedNames.ts`

**名称本地化基础设施**（~760 行）。

维护 70+ 支队伍和 20+ 个赛事的中英文映射：

格式：`英文原名 / 中文官方译名 / 民间翻译`

示例：
- `Vitality / Vitality战队 / 小蜜蜂`
- `IEM Rio / IEM 里约站 / 里约IEM`

赛事名称支持地理位置自动推导（如 `{series} {location}` → 中文站名）。

### 12. 受管上游 — `src/upstream/managedUpstream.ts`

**Python 进程生命周期管理**：

启动流程：
1. 解析 Python 解释器路径（支持 `python3` 命令查找）
2. 确保端口可用
3. 生成 `instanceToken`（UUID）
4. `spawn` Python 子进程（传递环境变量 `HLTV_UPSTREAM_PORT`、`HLTV_UPSTREAM_HEALTH_PATH`、`HLTV_UPSTREAM_INSTANCE_TOKEN`）
5. 轮询 healthcheck 端点（验证状态码 200 + instance_token 匹配）
6. 超时或进程异常退出则 fail fast

关闭流程：
- `SIGTERM` 发送，1 秒后 `SIGKILL` 强制终止
- 资源清理在 `beforeExit`/`SIGINT`/`SIGTERM`/`stdin end` 时触发

---

## 类型系统

### ToolResponse 泛型设计

```typescript
ToolResponse<TData, TItem, TResolvedEntity> {
  query: Record<string, unknown>;           // 实际执行的查询参数
  resolved_entity?: TResolvedEntity;        // 解析后的实体（如队伍/选手）
  data?: TData;                             // 主数据（如 TeamRecentData）
  items?: TItem[];                          // 列表数据（如 NormalizedMatch[]）
  meta: ToolMeta;                           // 元数据（来源/时间/缓存/分页）
  error: ToolError | null;                  // 错误信息
}
```

### AppError 错误码体系

```
INVALID_ARGUMENT      — 参数错误
ENTITY_NOT_FOUND      — 实体未找到
ENTITY_AMBIGUOUS      — 实体歧义（预留）
UPSTREAM_TIMEOUT      — 上游超时（可重试）
UPSTREAM_NOT_FOUND    — 上游 404（不可重试）
UPSTREAM_UNAVAILABLE  — 上游不可用（可重试）
UPSTREAM_BAD_DATA     — 上游数据格式异常
RATE_LIMITED          — 频率限制
LLM_SUMMARY_FAILED    — LLM 摘要失败（预留）
PARTIAL_DATA          — 部分数据
INTERNAL_ERROR        — 内部错误
```

---

## Python 上游 API 结构

### Flask 路由

| Blueprint | 端点前缀 | 路由 |
|-----------|---------|------|
| teams | `/api/v1/teams` | `/rankings[/:type[/:year/:month/:day]]`, `/search/:name`, `/:id/matches[/:offset]`, `/:id/:team_name` |
| players | `/api/v1/players` | `/search/:name`, `/stats/overview/:id/:player_name`, `/:id/:player_name` |
| matches | `/api/v1/matches` | `/upcoming` |
| results | `/api/v1/results` | `/[/<offset>]` |
| news | `/api/v1/news` | `/realtime`, `[/:year/:month]/` |

### Scrapy 爬虫 (14个)

| 爬虫 | 目标 |
|------|------|
| `hltv_teams_search` | 队伍搜索 |
| `hltv_team` | 队伍详情 |
| `hltv_team_matches` | 队伍比赛 |
| `hltv_players_search` | 选手搜索 |
| `hltv_player` | 选手详情 |
| `hltv_player_stats_overview` | 选手统计概览 |
| `hltv_results` | 近期赛果 |
| `hltv_upcoming_matches` | 未来赛程 |
| `hltv_news` | 归档新闻 |
| `hltv_realtime_news` | 实时新闻 |
| `hltv_big_results` | 大赛结果 |
| `hltv_top30` | 排名 |
| `hltv_valve_ranking` | V社排名 |
| `hltv_match` | 单场比赛 |

---

## 关键行为约定（不可回归）

这些是 AGENTS.md 中明确标记为 "must not regress" 的行为：

1. **`/match` 命令仅限今日**：裸 `/match` → `hltv_matches_today({})`；任何非空参数 → 拒绝
2. **`HltvFacade.getTodayMatches()` 委托为 `getUpcomingMatches({})`**：空查询触发今日模式
3. **泛化占位符剥离**：`upcomingMatchesQuery.ts` 剔除 "today matches"、"今日赛程" 等自动填充占位符
4. **缓存 stale 窗口兜底**：上游请求失败时优先返回过期缓存而非直接报错
5. **时间固定 Asia/Shanghai**：所有时间计算、格式化、日界线判断均使用 `Asia/Shanghai`

---

## 测试策略

### TypeScript 测试

- **测试框架**：Node.js 原生 `node --test` + `tsx` loader
- **运行方式**：`npm test`（`node --import tsx --test src/**/*.test.ts`）
- **特殊测试**：`src/matchCommandFlow.test.ts` — 端到端流程测试（/match 命令 + MCP 工具注册 + 模板文件）

### Python 测试

- **测试框架**：pytest
- **运行方式**：`make test-unit` / `make test-integration` / `make test-fast` / `make test-slow`
- **测试内容**：`test_routes.py`（路由测试）、`test_integration_real.py`（真实请求集成测试）、解析器测试

---

## 部署模式

### 模式 1：Managed Upstream（默认）
- MCP 启动时自动拉起 `hltv-api-fixed/app.py`
- 无需手动启动上游
- 环境变量：`HLTV_UPSTREAM_MANAGED=true`

### 模式 2：External Upstream
- 需要手动部署 `hltv-scraper-api`
- 环境变量：`HLTV_UPSTREAM_MANAGED=false`，`HLTV_API_BASE_URL=http://...`

### OpenCode 集成
- **MCP 注册**：手动编辑 `opencode.jsonc`，添加 `hltv_local` local MCP
- **命令注册**：手动复制 `docs/templates/` 下的 `.md` 模板到 `.opencode/commands/`
- **工具名前缀**：`hltv_local_`（取决于 MCP 名称）
- **诊断命令**：`npm run doctor:opencode`

---

## 值得关注的设计决策

1. **本地化先行**：70+ 队伍 + 20+ 赛事的中文名称映射，支持英文原名、官方译名、民间昵称三层查找
2. **防御性占位符过滤**：LLM 可能自动填充 `{team: "x", event: "y", limit: 1, days: 1}`，系统主动识别并清除
3. **并发推断**：选手所属队伍的推断使用并发 worker pool + 优先队列 + 活跃度排序，避免全量扫描
4. **Slug 容错**：实体 slug 不再仅靠本地 `slugify()`，而是从上游链接、详情响应等多来源构建候选
5. **WSL 友好**：自动检测 WSL 环境并添加 Windows 宿主机 IP 作为备用 baseUrl
6. **Token 验证健康检查**：managed upstream 使用随机 UUID 验证健康的响应是当前实例而非残留
