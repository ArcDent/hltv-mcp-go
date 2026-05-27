# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈重建
- 目标：Go 单二进制，同时运行 MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS v4
- 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API）
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
- 2026-05-27：前端完全重写 — 体育转播大屏风格（Oswald 字体、48px 比分、两列赛程卡片）
- 2026-05-27：Dashboard 大卡片布局（56px 数值、2x2 网格）+ 暗/亮双主题切换（CSS transition 0.5s）
- 2026-05-27：修复 3 轮 Docker 构建问题（镜像名、frontend dist 路径、GOTOOLCHAIN=auto）
- 2026-05-27：全部 6/6 爬虫端点 E2E 验证通过（含 chromedp 反 Cloudflare）
- 2026-05-27：创建 GitHub 仓库并推送（ArcDent/hltv-mcp-go）

## 进行中
- 前端细节打磨（赛事名缩写、队伍名对齐、主题切换动效）

## 下一步
- 完整 70+ 队伍 localization 扩展
- 选手队伍推断实现（参考原 hltv-mcp 的优先队列 + roster 扫描）
- OpenCode slash command 模板
- 前端队伍/选手详情页展开

## 关键发现

### 爬虫
- 5/6 端点 HTTP 直连，仅 `/matches` 需 chromedp
- chromedp 反 CF 关键：`UserDataDir`（持久化 profile）+ `--disable-blink-features=AutomationControlled` + Chrome 132 UA

### HLTV HTML 结构
- 赛果 `.result-con` > `.line-align.team1 .team` / `.result-score`
- 赛程 `.match` > `.match-top`(赛事) + `.match-teams`(队伍) + `.match-info`(时间)
- 搜索 `table tbody tr > a[href*='/team/']` 正则提取 ID
- 新闻 `.newstext` 文本在 div 内，链接需父级查找

### 构建与部署
- `go build .` 因 `frontend/` 无 Go 文件失败 → 用 `go build github.com/arcdent/hltv-mcp`
- Docker: `GOTOOLCHAIN=auto` + `chromedp/headless-shell:latest` Chrome 路径 `/headless-shell/headless-shell`
- SPA fallback: `feFS.Open(path)` 必须 strip 前导 `/`

### 前端设计系统
- 字体：Oswald（标题/比分）+ Noto Sans SC（中文正文）+ JetBrains Mono（数据）
- 配色：深空背景 `#0a0a0f` + 金色 `#f5c842` + 红色 `#ff4444`（LIVE）
- 暗/亮双主题：CSS 变量 + `.light` class + transition 0.5s
- 布局：顶部导航栏 + 全宽内容区 + 两列网格
- 赛程卡片：队伍居中对齐 + 中文昵称预留高度 + 赛事名缩写
