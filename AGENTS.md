# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈重建
- 目标：Go 单二进制，同时运行 MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS v4
- 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API）
- 前端参考：[person-summon](https://github.com/arcdent/person-summon)（暗/亮双主题 CSS 变量体系 + Space Grotesk + 噪声纹理）
- 远端仓库：[ArcDent/hltv-mcp-go](https://github.com/ArcDent/hltv-mcp-go)

## 项目静态结构
```
├── main.go                    # 入口：MCP stdio + HTTP :8082 双 goroutine
├── Dockerfile                 # 三阶段：frontend → Go → chromedp/headless-shell
├── docker-compose.yml
├── internal/
│   ├── types/                 # 全部共享类型
│   ├── errors/                # AppError + 11 错误码
│   ├── config/                # 18 环境变量
│   ├── cache/                 # TTL + stale 窗口 + 并发合并
│   ├── client/                # HTTP + chromedp 反CF + fallback 记忆
│   ├── scraper/               # 7 爬虫（team/player/results/matches/news/realtime_news/news_article）
│   ├── localization/          # 26 队伍 + 3 赛事中英映射
│   ├── normalizer/            # HLTV HTML → 标准化类型
│   ├── facade/                # 核心编排层（withCache + resolve + matches + news）
│   ├── summary/               # 中文摘要生成
│   ├── renderer/              # 中文格式化 MCP 输出
│   ├── mcp/                   # 10 MCP 工具注册 + stdio 传输
│   └── http/                  # chi router + 12 REST API + SPA fallback
├── frontend/                  # React + Vite + Tailwind
│   └── src/
│       ├── api/client.ts      # 13 API 方法（含 getNewsArticle）
│       ├── components/        # NewsDetail, PlayerDetail, TeamDetail, SearchableList, TranslateProvider
│       └── pages/             # Dashboard, Matches, Teams, Players, News (集成 NewsDetail 点击弹窗), Cache
├── cmd/                       # 调试/验证工具（scrapercheck, antifp, e2e）
└── docs/superpowers/          # spec + plan
```

## 最近操作
- 2026-05-27：Task 14 完成 — 新建 NewsDetail React 组件（全屏模态框 + backdrop blur + 文章正文 pre-wrap 渲染 + 翻译按钮复用 TranslateProvider + localStorage 无限 TTL 翻译缓存 + "在 HLTV 阅读原文" 外链），集成到 News.tsx 中（点击新闻条目弹出详情），添加 getNewsArticle API 方法，构建通过并提交
- 2026-05-27：Task 13 完成 — 添加 /api/news/article?url= GET handler（URL 参数编码、调用 GetNewsArticleCached 模式、30s 超时 + 优雅降级），编译通过并提交
- 2026-05-27：Task 12 完成 — 添加 NewsArticleScraper（GetArticle 爬取单篇文章页面）、NormalizeNewsArticle（提取标题/日期/作者/正文）和 facade 方法（ScrapeNewsArticle + GetNewsArticleCached 用 MD5 URL hash 做无限期缓存），编译通过并提交
- 2026-05-27：Task 10 完成 — CF Smoke Test 确认 HLTV /news/ 页面 HTTP 直连返回 200，不被 CF Challenge 阻断（与 /stats/matches/ 不同），Feature 4 新闻文章抓取可正常推进
- 2026-05-27：Task 8 完成 — 添加 /api/events 后端 API（EventGroup + EventsResponse 类型、GetEvents + groupByEvent facade 方法、handler + 路由），编译通过并提交
- 2026-05-27：Task 7 完成 — 将 TeamDetail 集成到 SearchableList，队伍搜索结果可点击弹出 TeamDetail 模态框（新增 selectedTeamId 状态 + TeamDetail 导入 + onClick handler 分支 + cursor pointer）
- 2026-05-27：Task 6 完成 — 新建 TeamDetail React 组件，全屏模态框（backdrop blur + slideUp 动画），展示队伍排名/积分/战绩/成就/近期 10 场/队员阵容，队员可点击穿透至 PlayerDetail
- 2026-05-27：Task 5 完成 — 替换 GetTeam HTTP handler 存根为真实实现，新增 TeamDetail 类型/规格化/爬虫/facade（Task 1-4 前置），handler 模式镜像 GetPlayer 使用 GetTeamDetailCached + 30s 超时 + 优雅降级
- 2026-05-27：Chrome DevTools 全功能验证 — 占位符翻译 winner→胜者、BO1 归一化 0:1、缓存统计真实递增（6条目/3命中/7未命中）、选手详情缓存均已确认正常；发现 Windows localhost 端口转发会缓存旧响应，需用 WSL IP 直连
- 2026-05-27：全量中文化 + BO1 归一化 + 选手缓存 + 缓存统计修复（8 commits，7 tasks）— 赛程 winner/loser/tbd 映射为胜者/败者/待定、选手近期比赛 BO1 比分 13:5→1:0、选手详情 7 天 chromedp 缓存、修复缓存统计硬编码为 0
- 2026-05-27：修复队伍名显示 "Link" — `.playerTeam a` 选择器误抓合约来源链接，改用 `.playerTeam a[itemprop="text"]` 精确定位
- 2026-05-27：选手页面直接提取比赛比分 — `.playerpage-match-result` 元素包含比分（如 "2:0"），无需爬 `/stats/matches/`，同时清理空格格式

## 进行中
- 比赛个人 rating 数据获取 — 选手页 `.playerpage-match-rating` 始终为空，需要其他数据源（API 端点为 401 需认证）
- 选手队伍推断实现（参考原 hltv-mcp 的优先队列 + roster 扫描）

## 下一步
- Task 15：最终验证（全功能端到端检查，确认所有 15 个任务完整交付）
- 完整 70+ 队伍 localization 扩展
- OpenCode slash command 模板

## 关键发现

### 爬虫
- 5/6 端点 HTTP 直连，仅 `/matches` 需 chromedp
- chromedp 反 CF 关键：`UserDataDir`（持久化 profile）+ `--disable-blink-features=AutomationControlled` + Chrome 132 UA

### HLTV HTML 结构
- 赛果 `.result-con` > `.line-align.team1 .team` / `.result-score`
- 赛程 `.match` > `.match-top`(赛事) + `.match-teams`(队伍) + `.match-info`(时间)
- 搜索 `table tbody tr > a[href*='/team/']` 正则提取 ID
- 新闻 `.newstext` 文本在 div 内，链接需父级查找
- **选手页** `.playerNickname` / `.playerRealname` / `.playerTeam a[itemprop="text"]`(队伍，不可用裸 `a` 会抓到合约 Link) / `.player-stat` > `.statsVal p b`(能力值) / `.stats-window`(maps 数) / `.playerpage-matchbox`(近期比赛) / `.playerpage-match-result`(比分，格式 "2 : 0") / `.playerpage-match-rating`(个人 rating，始终为空) / `.majorSection` > `.majorWinner/.majorMVP`(荣誉) / `.mvp-count`(MVP 数) / `.all-time-stat` > `.stat` + `.description`(生涯) / `.playerInfoRow.playerAge` / `.playerTop20`
- **比赛链接** 从 `.playerpage-matchbox[href]` 正则 `/stats/matches/(\d+)/([^"\s]+)` 提取 match ID + slug
- **球员无队伍** `.playerTeam` 内 `<span itemprop="text">No team</span>` 表示无队伍，这是正常状态
- **球员有合约链接** 部分球员 .playerTeam 内包含 `<a class="contract-link">Link</a>`（合约来源链接），会被误抓为队伍名

### CF 阻断分层
- **HTTP 直连可用（无需 chromedp）**：`/player/`、`/results`、`/matches`、`/team/`、**`/news/`** — 无需反 CF 措施即可直连
- **可通过 chromedp**：`/player/`、`/results`、`/matches`、`/team/` — UserDataDir + anti-blink 有效
- **被 CF Challenge 阻断**：`/stats/matches/` — JS Challenge 无法在 headless 中完成，即使 20s 等待仍返回 "Just a moment..."

### 构建与部署
- `go build .` 因 `frontend/` 无 Go 文件失败 → 用 `go build github.com/arcdent/hltv-mcp`
- Docker: `GOTOOLCHAIN=auto` + `chromedp/headless-shell:latest` Chrome 路径 `/headless-shell/headless-shell`
- SPA fallback: `feFS.Open(path)` 必须 strip 前导 `/`

### 赛中文化与缓存模式
- HLTV bracket 占位符映射：winner/loser/tbd → 胜者/败者/待定，使用 `strings.Contains` 包含匹配（HLTV 实际文本可能是 "Winner of Group A" 格式）
- BO1 比分归一化：任一侧 >= 13 判定为 BO1，转为 1:0/0:1；`.result-con` 回退路径中归一化必须在胜负判断之后执行
- 选手详情缓存：`types.PlayerDetail` 不走 `withCache` 包装，直接用 `cache.Get/Set`（非 `*ToolResponse` 类型），key 格式 `player_detail:<id>`
- cache.GetStale 不计入 hits/misses（降级兜底不在统计范围）
- `sync/atomic.Int64` 用于 cache 计数器，与 `sync.RWMutex` 无锁竞争

### 测试与验证
- 通过 Chrome DevTools MCP 在 Windows Chrome 中直接测试前端页面效果
- **关键问题**：Windows `localhost:8082` 端口转发会缓存/代理旧 HTTP 响应（即使 WSL 进程已重启），导致前端显示过期数据。解决方案：使用 WSL IP（如 `172.21.32.31:8082`）直连，绕过 Windows 端口转发层
- **验证方法**：`curl` 从 WSL 内部调用 API 对照浏览器网络请求，若两者返回不同数据（不同 `fetched_at`），则问题在 Windows 端口转发层而非 Go 代码
- `strings` 命令无法找到 UTF-8 中文字符串（多字节序列），需用 `grep -a` 在二进制中搜索

### 前端设计系统（参考 person-summon）
- 主题：CSS 变量（`:root` 亮色 / `[data-theme="dark"]` 暗色）+ `transition: background-color 0.3s, color 0.2s, border-color 0.3s`
- 色板：`--gold: #f5c842` / `--gold-dim` / `--gold-glow` + `--red`/`--green` 语义色
- 卡片：`background: var(--card)` + `border: 1px solid var(--border)` + `box-shadow: var(--card-shadow)` + hover border transition
- 输入框：`background: var(--input-bg)` + focus 时 `border-color: var(--gold)` + `box-shadow: 0 0 0 3px var(--gold-dim)`
- 字体：Oswald（标题/比分）+ Noto Sans SC（正文）+ JetBrains Mono（数据）
- 特效：暗色模式 SVG 噪声纹理 + `fadeIn`/`slideUp`/`pulseGlow` 关键帧动画
- 布局：左侧 sticky 竖状导航 180px + 右侧滚动内容区 max-w-[1100px]
