# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈重建
- 目标：Go 单二进制，同时运行 MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS
- 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API）

## 项目静态结构
```
├── main.go                    # 入口：MCP stdio + HTTP :8082 双 goroutine
├── Dockerfile                 # 三阶段构建：frontend → Go → chrome-headless-shell
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
│   └── http/                  # chi router + 12 REST API
├── frontend/                  # React + Vite + Tailwind
│   └── src/pages/             # Dashboard, Matches, Teams, Players, News, Cache
├── cmd/                       # 调试/验证工具
└── docs/superpowers/          # spec + plan
```

## 最近操作
- 2026-05-27：全部 6/6 爬虫端点验证通过（含 chromedp 反 Cloudflare）
- 2026-05-27：E2E 测试 6/6 PASS（赛程/赛果/队伍搜索/选手搜索/实时新闻/归档新闻）
- 2026-05-27：Go 后端 10 包 25 测试 PASS + React 前端 6 页面
- 2026-05-27：单二进制 17MB，MCP + HTTP :8082 双 goroutine 运行正常

## 进行中
- 创建 GitHub 远程仓库并推送

## 下一步
- 完整 70+ 队伍 localization 扩展
- 选手队伍推断实现（参考原 hltv-mcp 的优先队列 + roster 扫描）
- OpenCode slash command 模板

## 关键发现
### 爬虫
- 5/6 端点支持 HTTP 直连，仅 `/matches` 需 chromedp
- chromedp 反 CF 关键：`UserDataDir`（持久化 profile）+ `--disable-blink-features=AutomationControlled`

### HLTV HTML 结构
- 赛果（`/results`）：`.result-con` > `.line-align.team1 .team` / `.result-score`
- 赛程（`/matches`）：React 组件 `.match` > `.match-top`(赛事) + `.match-teams`(队伍) + `.match-info`(时间)
- 搜索：`table tbody tr > a[href*='/team/']` 或 `a[href*='/player/']`，正则提取 ID
- 新闻归档：`.newstext` 文本直接在 div 内，链接在父级

### 构建
- `go build .` 因 `frontend/` 无 Go 文件会失败，需用 `go build github.com/arcdent/hltv-mcp`
- chromedp v0.15 需要 Go >= 1.26
- mcp-go 请求参数用 `req.GetString/GetInt/GetBool(key, default)` 模式
