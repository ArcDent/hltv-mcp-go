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
│   ├── config/                # 17 环境变量
│   ├── cache/                 # TTL + stale 窗口 + 并发合并
│   ├── client/                # HTTP + chromedp 反CF + fallback 记忆
│   ├── scraper/               # 6 爬虫（team/player/results/matches/news/realtime_news）
│   ├── localization/          # 26 队伍 + 3 赛事中英映射
│   ├── normalizer/            # HLTV HTML → 标准化类型
│   ├── facade/                # 核心编排层（withCache + resolve + matches + news）
│   ├── summary/               # 中文摘要生成
│   ├── renderer/              # 中文格式化 MCP 输出
│   ├── mcp/                   # 10 MCP 工具注册 + stdio 传输
│   └── http/                  # chi router + 12 REST API + SPA fallback
├── frontend/                  # React + Vite + Tailwind
│   └── src/pages/             # Dashboard, Matches, Teams, Players, News, Cache
├── cmd/                       # 调试/验证工具（scrapercheck, antifp, e2e）
└── docs/superpowers/          # spec + plan
```

## 最近操作
- 2026-05-27：修复队伍名显示 "Link" — `.playerTeam a` 选择器误抓合约来源链接，改用 `.playerTeam a[itemprop="text"]` 精确定位
- 2026-05-27：选手页面直接提取比赛比分 — `.playerpage-match-result` 元素包含比分（如 "2:0"），无需爬 `/stats/matches/`，同时清理空格格式
- 2026-05-27：修复 PlayerDetailProfile.Team/Country omitempty — 去除 omitempty 确保字段始终存在于 JSON 响应
- 2026-05-27：前端空值显示 "暂无队伍"/"未知国籍" — PlayerDetail.tsx 对 team/country/realname 增加空值 fallback
- 2026-05-27：确认 `/stats/matches/` 即使 Firecrawl stealth 代理也无法突破 CF 403

## 进行中
- 比赛个人 rating 数据获取 — 选手页 `.playerpage-match-rating` 始终为空，需要其他数据源（API 端点为 401 需认证）
- 选手队伍推断实现（参考原 hltv-mcp 的优先队列 + roster 扫描）

## 下一步
- 完整 70+ 队伍 localization 扩展
- OpenCode slash command 模板
- 清理不再需要的 MatchDetailScraper/NormalizeMatchDetail/EnrichWithScores 死代码（比分已在选手页获取）

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
- **可通过 chromedp**：`/player/`、`/results`、`/matches`、`/team/` — UserDataDir + anti-blink 有效
- **被 CF Challenge 阻断**：`/stats/matches/` — JS Challenge 无法在 headless 中完成，即使 20s 等待仍返回 "Just a moment..."

### 构建与部署
- `go build .` 因 `frontend/` 无 Go 文件失败 → 用 `go build github.com/arcdent/hltv-mcp`
- Docker: `GOTOOLCHAIN=auto` + `chromedp/headless-shell:latest` Chrome 路径 `/headless-shell/headless-shell`
- SPA fallback: `feFS.Open(path)` 必须 strip 前导 `/`

### 前端设计系统（参考 person-summon）
- 主题：CSS 变量（`:root` 亮色 / `[data-theme="dark"]` 暗色）+ `transition: background-color 0.3s, color 0.2s, border-color 0.3s`
- 色板：`--gold: #f5c842` / `--gold-dim` / `--gold-glow` + `--red`/`--green` 语义色
- 卡片：`background: var(--card)` + `border: 1px solid var(--border)` + `box-shadow: var(--card-shadow)` + hover border transition
- 输入框：`background: var(--input-bg)` + focus 时 `border-color: var(--gold)` + `box-shadow: 0 0 0 3px var(--gold-dim)`
- 字体：Oswald（标题/比分）+ Noto Sans SC（正文）+ JetBrains Mono（数据）
- 特效：暗色模式 SVG 噪声纹理 + `fadeIn`/`slideUp`/`pulseGlow` 关键帧动画
- 布局：左侧 sticky 竖状导航 180px + 右侧滚动内容区 max-w-[1100px]
