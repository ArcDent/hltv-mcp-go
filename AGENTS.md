# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈重建
- 目标：Go 单二进制，同时运行 MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS
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
- 2026-05-27：修复 Dockerfile（chromedp/headless-shell 镜像、GOTOOLCHAIN=auto、frontend dist 路径）
- 2026-05-27：修复 SPA 前端空白 — `feFS.Open` 路径需 strip 前导 `/`
- 2026-05-27：README 补充 4 平台（Linux/WSL/macOS/Windows）+ Docker 部署说明
- 2026-05-27：创建 GitHub 仓库并推送（`ArcDent/hltv-mcp-go`）
- 2026-05-27：全部 6/6 爬虫端点 E2E 验证通过

## 进行中
- Docker 镜像构建验证（等待 Windows Docker Desktop 构建结果）

## 下一步
- 完整 70+ 队伍 localization 扩展
- 选手队伍推断实现（参考原 hltv-mcp 的优先队列 + roster 扫描）
- OpenCode slash command 模板
- 前端页面功能细化（搜索防抖、详情展开、分页）

## 关键发现

### 爬虫
- 5/6 端点支持 HTTP 直连，仅 `/matches` 需 chromedp
- chromedp 反 CF 关键：`UserDataDir`（持久化 profile）+ `--disable-blink-features=AutomationControlled`

### HLTV HTML 结构
- 赛果（`/results`）：`.result-con` > `.line-align.team1 .team` / `.result-score`
- 赛程（`/matches`）：`.match` > `.match-top`(赛事) + `.match-teams`(队伍) + `.match-info`(时间)
- 搜索：`table tbody tr > a[href*='/team/']` 或 `a[href*='/player/']`，正则提取 ID
- 新闻归档：`.newstext` 文本直接在 div 内，链接需从父级查找

### 构建与部署
- `go build .` 因 `frontend/` 无 Go 文件会失败，需用 `go build github.com/arcdent/hltv-mcp`
- Docker：`GOTOOLCHAIN=auto` 解决 Go 1.24 镜像编译 1.26 代码；`vite outDir: '../dist'` → Docker COPY 路径是 `/app/dist`
- SPA fallback：`fs.FS` 路径不以 `/` 开头，必须 `strings.TrimPrefix(req.URL.Path, "/")`

### 已验证的 Docker 配置
- 前端构建：`node:22-alpine`
- Go 编译：`golang:1.24-alpine` + `ENV GOTOOLCHAIN=auto`
- 运行时：`chromedp/headless-shell:latest`，Chrome 路径 `/headless-shell/headless-shell`
