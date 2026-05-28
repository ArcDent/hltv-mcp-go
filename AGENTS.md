# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈重建
- 目标：Go 单二进制，同时运行 MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS v4
- 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API）
- 前端参考：[person-summon](https://github.com/arcdent/person-summon)（暗/亮双主题 CSS 变量体系 + Space Grotesk + 噪声纹理）
- 远端仓库：[ArcDent/HLTV-data](https://github.com/ArcDent/HLTV-data)
- 许可证：MIT

## 项目静态结构
```
├── main.go                    # 入口：MCP stdio + HTTP :8082 双 goroutine
├── Dockerfile                 # 三阶段：frontend → Go → chromedp/headless-shell
├── docker-compose.yml
├── internal/
│   ├── types/                 # 全部共享类型
│   ├── errors/                # AppError + 8 错误码
│   ├── config/                # 17 环境变量
│   ├── crypto/                # AES-256-GCM 加密/解密 + 密钥生成/持久化
│   ├── cache/                 # TTL + stale 窗口 + 并发合并
│   ├── client/                # HTTP + chromedp 反CF + fallback 记忆
│   ├── scraper/               # 6 爬虫（team/player/results/matches/news/realtime_news/news_article）
│   ├── localization/          # 26 队伍 + 98 选手中英映射
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
```

## 最近操作
- 2026-05-28：修复 Docker 翻译 502 — `chromedp/headless-shell` 基础镜像缺少 `ca-certificates`，Go HTTP 客户端无法完成 TLS 验证；Dockerfile 新增 `apt-get install ca-certificates`；`PostTranslate` 连接错误路径新增 `log.Printf`；`PutTranslateConfig` 遮罩恢复失败时不再静默写入遮罩 key
- 2026-05-28：代码深度收敛 — 删 10 个文件（.clinerules-* ×5 + frontend/hltv-mcp + docs/superpowers/ ×4）、合并 shared.go/transport.go 薄包装、删除 SummaryMode/Raw 死代码路径、删除 3 个未用错误码、删除 7 个未使用查询字段，源文件 43→40 个
- 2026-05-28：昵称字典后端迁移+编辑功能 — 新增 `overrides.go` 持久化覆盖层 + `PlayerCatalog`（95 选手）+ 3 个 REST API（`GET/PUT /api/nicknames*`）+ 前端 `useNicknames` hook + TeamDetail/PlayerDetail 内联编辑，删除 `frontend/src/data/nicknames.ts`
- 2026-05-28：CI/CD 与文档完善 — GitHub Actions 自动构建推送 GHCR、添加 MIT 许可证、修正 GHCR 镜像路径、Docker 部署示例按平台汇总
- 2026-05-28：本地化字典全面修正 + 补全 — Official 字段清空、G2/HEROIC/Complexity/MongolZ/fnatic/EF/RED Canids Colloquial 修正、赛事翻译全部删除、选手简称补全至 98 名

## 进行中
- 无

## 下一步
- 无

## 关键发现

### Docker SSL 证书
- **`chromedp/headless-shell` 基础镜像不含 `ca-certificates`**，Go 的 `crypto/x509` 在 Linux 容器内找不到系统根 CA 时，所有 HTTPS 出站连接（`http.DefaultClient`）会因 `x509: certificate signed by unknown authority` 失败
- **HLTV 爬虫不受影响**：爬虫优先使用 chromedp（Chrome 自带 CA 存储），不经过 Go HTTP 客户端
- **修复**：`Dockerfile` runtime stage 添加 `apt-get install -y ca-certificates`

### nickname 覆盖层
- **`internal/localization/overrides.go`**：线程安全的内存缓存 + JSON 文件持久化
- `data/nicknames.json` 结构 `{"teams": {...}, "players": {...}}`，仅存用户编辑条目
- `SetTeamOverride("name", "")` / `SetPlayerOverride("name", "")` 空值语义 = 删除条目
- 写操作：先更新内存 map（持写锁），释放锁后写磁盘（持读锁读数据），避免磁盘 I/O 被锁阻塞
- 测试从 `internal/localization/` 目录运行，`data/nicknames.json` 相对路径 `../../data/nicknames.json`

### nickname API
- `PUT /api/nicknames/team` 先通过 `LookupTeam` 解析为 canonical 名再存储（支持别名输入）
- `PUT /api/nicknames/player` 直接按输入名存储（开放模式，不限制 catalog 内选手）
- `GET /api/nicknames` 返回 `{"teams": {...}, "players": {...}}`，所有变体已展开，覆盖已应用

### 本地化
- **nickname 已完全迁移到后端**：前端硬编码 `frontend/src/data/nicknames.ts` 已删除，改用 `useNicknames` hook 从 `GET /api/nicknames` 获取
- **前端内联编辑**：TeamDetail 队伍简称 badge 和队员昵称、PlayerDetail 选手昵称均可点击铅笔图标编辑，Enter/Escape/失焦保存
- **nickname 核心在后端**：`internal/localization/catalog.go` 的 `TeamCatalog.Colloquial` + `PlayerCatalog.Nicknames`，overrides 通过 `data/nicknames.json` 持久化
- **赛事翻译已全部移除**：`events.go` 中 `EventCatalog` 清空，`FormatEventDisplay` 回退为原文输出
- 当前 26 支队伍 + 98 名选手的简称字典已逐条人工确认

### 爬虫
- 5/6 端点 HTTP 直连，仅 `/matches` 需 chromedp
- chromedp 反 CF 关键：`UserDataDir`（持久化 profile）+ `--disable-blink-features=AutomationControlled` + Chrome 132 UA

### HLTV HTML 结构
- 赛果 `.result-con` > `.line-align.team1 .team` / `.result-score`
- 赛程 `.matches-list-headline`(日期标题) 与 `.match-wrapper`(含 `.match`) 同级子节点，遍历 parent children 文档顺序提取日期 → `.match` 子树 `.match-top`(赛事) + `.match-teams`(队伍) + `.match-info`(时间)
- 搜索 `table tbody tr > a[href*='/team/']` 正则提取 ID
- 新闻 `.newstext` 文本在 div 内，链接需父级查找
- **选手页** `.playerNickname` / `.playerRealname` / `.playerTeam a[itemprop="text"]`(队伍，不可用裸 `a` 会抓到合约 Link) / `.player-stat` > `.statsVal p b`(能力值) / `.stats-window`(maps 数) / `.playerpage-matchbox`(近期比赛) / `.playerpage-match-result`(比分，格式 "2 : 0") / `.playerpage-match-rating`(个人 rating，始终为空) / `.majorSection` > `.majorWinner/.majorMVP`(荣誉) / `.mvp-count`(MVP 数) / `.all-time-stat` > `.stat` + `.description`(生涯) / `.playerInfoRow.playerAge` / `.playerTop20`
- **比赛链接** 从 `.playerpage-matchbox[href]` 正则 `/stats/matches/(\d+)/([^"\s]+)` 提取 match ID + slug
- **球员无队伍** `.playerTeam` 内 `<span itemprop="text">No team</span>` 表示无队伍，这是正常状态
- **球员有合约链接** 部分球员 .playerTeam 内包含 `<a class="contract-link">Link</a>`（合约来源链接），会被误抓为队伍名
- **队伍页** — `h1.profile-team-name` / `.profile-team-container` / `.value.h-rank` / `.bodyshot-team.g-grid a[href*='/player/']`(现役5人) / `.trophySection .trophyDescription[title]` / **队员链接重复5-10次**，须限定 `.bodyshot-team`
- **队伍高亮** — `.highlighted-stat`(含胜率"76.2%"+"Win rate" / 连胜"6"+"Current win streak") / `.last-5-matches`(最近5场) / `.highlighted-team-name.text-ellipsis`(对手名) / `.highlighted-match-status.match-won/.match-lost`(胜负) — **无具体比分**，仅对手名+W/L
- **结果页日期** `.results-sublist > .standard-headline`("Results for May 28th 2026")
- **赛程页日期** `.matches-list-headline`("Thursday - 2026-05-28")，Live 区 `.liveMatches` 无日期，默认当天
- **队伍赛程页** `/team/{id}/matches` — 非标准 `table.match-table` 布局，无法用现有 normalizer
- **新闻文章页** — `.news-block` 内正文在 `<p>` 标签中，页面底部含 `.related-news` / `.player-card` / `.teammate` / `.comment-section` 等推荐卡片。提取时必须只取 `<p>` 标签（`doc.Find(".news-block p")`），不可用 `.Text()` 取整个容器否则会混入推荐新闻/选手卡片等垃圾内容

### CF 阻断分层
- **HTTP 直连可用（无需 chromedp）**：`/player/`、`/results`、`/matches`、`/team/`、**`/news/`** — 无需反 CF 措施即可直连
- **可通过 chromedp**：`/player/`、`/results`、`/matches`、`/team/` — UserDataDir + anti-blink 有效
- **被 CF Challenge 阻断**：`/stats/matches/` — JS Challenge 无法在 headless 中完成，即使 20s 等待仍返回 "Just a moment..."

### 构建与部署
- `go build .` 因 `frontend/` 无 Go 文件失败 → 用 `go build github.com/arcdent/hltv-mcp`
- **前端嵌入变更需重建二进制**：修改 `frontend/src/` 后，必须 `vite build` + `go build` + 重启服务，否则浏览器看到的是旧内嵌前端（即使 dist/ 已更新）
- Docker: `GOTOOLCHAIN=auto` + `chromedp/headless-shell:latest` Chrome 路径 `/headless-shell/headless-shell`
- SPA fallback: `feFS.Open(path)` 必须 strip 前导 `/`
- **CI/CD**：`.github/workflows/docker-build.yml` — push main 自动构建镜像推送到 `ghcr.io/arcdent/hltv-data:latest`，用 `docker/metadata-action` 打 `latest` + commit SHA 双 tag。服务器端搭配 Watchtower 自动拉取实现零手动部署。

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
