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
- 2026-05-27：前端采用 person-summon 设计语言 + 竖状侧栏 + 数据源弹窗 + 暗/亮主题
- 2026-05-27：选手详情卡片 — chromedp 抓取 + 八维雷达 + Top20 + 荣誉 + 近期 7 场
- 2026-05-27：新闻翻译 — 后端配置API(文件持久化) + 前端OpenAI兼容面板 + 双语展示 + localStorage缓存7天 + sessionStorage保留Key
- 2026-05-27：修复翻译组件3个bug — masked key覆盖、worker提前return、sessionStorage恢复
- 2026-05-27：默认为白天主题，修复首次点击无响应
- 2026-05-27：侧栏竖状导航 + 暗/亮主题切换按钮 + 全宽内容区 + 竖状侧栏 + 数据源弹窗 + 新闻中文翻译 + 标签切换动效 — CSS 变量双主题 + 卡片系统 + 聚焦光环 + 噪声纹理
- 2026-05-27：侧栏竖状导航 + 暗/亮主题切换按钮 + 全宽内容区
- 2026-05-27：修复 3 轮 Docker 构建问题（镜像名、frontend dist 路径、GOTOOLCHAIN=auto）
- 2026-05-27：全部 6/6 爬虫端点 E2E 验证通过（含 chromedp 反 Cloudflare）
- 2026-05-27：创建 GitHub 仓库并推送（ArcDent/hltv-mcp-go）

## 进行中（sessionStorage Key 持久化已修复）（提取 SearchableList 共享组件，Teams/Players 各减 82%）
- 前端细节打磨（赛事名缩写、队伍名对齐、主题切换动效）

## 下一步
- 完整 70+ 队伍 localization 扩展
- 选手队伍推断实现（参考原 hltv-mcp 的优先队列 + roster 扫描）
- OpenCode slash command 模板
- 选手详情卡片已完成（chromedp 抓取 + 八维雷达 + Top20 + 荣誉 + 近期 7 场）— ZywOo/s1mple/donk 验证通过

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

### 前端设计系统（参考 person-summon）
- 主题：CSS 变量（`:root` 亮色 / `[data-theme="dark"]` 暗色）+ `transition: background-color 0.3s, color 0.2s, border-color 0.3s`
- 色板：`--gold: #f5c842` / `--gold-dim` / `--gold-glow` + `--red`/`--green` 语义色
- 卡片：`background: var(--card)` + `border: 1px solid var(--border)` + `box-shadow: var(--card-shadow)` + hover border transition
- 输入框：`background: var(--input-bg)` + focus 时 `border-color: var(--gold)` + `box-shadow: 0 0 0 3px var(--gold-dim)`
- 字体：Oswald（标题/比分）+ Noto Sans SC（正文）+ JetBrains Mono（数据）
- 特效：暗色模式 SVG 噪声纹理 + `fadeIn`/`slideUp`/`pulseGlow` 关键帧动画
- 布局：左侧 sticky 竖状导航 180px + 右侧滚动内容区 max-w-[1100px]
