# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务 Go 全栈重建
- 目标：Go 单二进制全栈应用，MCP stdio + HTTP REST + React 管理面板
- 技术栈：Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS

## 项目静态结构
```
hltv-mcp-fully-rebuild/
├── main.go              # 入口：双 goroutine 启动 MCP + HTTP
├── Dockerfile            # 三阶段构建：frontend → Go → chrome-headless-shell
├── docker-compose.yml    # 一键部署
├── internal/
│   ├── types/types.go    # 全部共享类型
│   ├── errors/errors.go  # AppError + 11 错误码
│   ├── config/config.go  # 17 环境变量配置
│   ├── cache/cache.go    # TTL + stale + 并发合并
│   ├── client/           # HTTP + chromedp fallback + 记忆追踪
│   ├── scraper/          # 6 爬虫模块
│   ├── localization/     # 26 队伍 + 3 赛事中英文映射
│   ├── normalizer/       # HLTV HTML → 标准化类型
│   ├── facade/           # 核心编排（withCache + resolve + matches + news）
│   ├── summary/          # 中文摘要
│   ├── renderer/         # 中文格式化输出
│   ├── mcp/              # 10 MCP 工具注册 + stdio
│   └── http/             # chi router + REST API handlers
├── frontend/             # React + Vite + Tailwind
│   └── src/pages/        # Dashboard, Matches, Teams, Players, News, Cache
└── docs/superpowers/     # spec + plan
```

## 最近操作
- 2026-05-27：完成 Go 后端全部 10 个包（25 测试 PASS）+ React 前端 6 页面 + Dockerfile
- 2026-05-27：单二进制编译成功（17MB），MCP stdio + HTTP :8082 双 goroutine 运行正常
- 2026-05-26：完成原 hltv-mcp 项目源码分析，产出 spec 和 plan

## 进行中
- 前端集成验证（二进制内嵌 React 产物已确认工作）

## 下一步
- 实际爬虫集成测试（需 HLTV.org 网络访问）
- Docker 构建验证
- 可选：扩展 localization 到 70+ 队伍（当前 26）

## 关键发现
- `go build .` 在项目根目录因 frontend/ 无 Go 文件会失败，需用 `go build github.com/arcdent/hltv-mcp`
- chromedp v0.15 需要 Go >= 1.26
- mcp-go API 使用 `req.GetString/GetInt/GetBool(key, default)` 模式
