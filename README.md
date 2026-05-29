# HLTV MCP Service

Go 单二进制全栈 HLTV MCP 服务 — MCP stdio + HTTP REST + React 管理面板。

> 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API），使用 Go 完全重建，去除 Python 上游依赖。

## 功能特性

- **9 个 MCP 工具**：队伍/选手解析、赛程/赛果查询、实时/归档新闻
- **Web 管理面板**：React SPA，6 个页面（Dashboard / Matches / Teams / Players / News / Cache）
- **反爬虫**：HTTP 直连优先 + chromedp 绕过 Cloudflare，5 分钟失败记忆窗口
- **中文输出**：26 支队伍民间昵称映射 + 98 名选手中文简称 + 中文摘要生成
- **翻译**：接入 OpenAI / DeepSeek / Groq / Ollama 等兼容 API，新闻标题自动翻译 + 正文一键翻译
- **Docker 一键部署**：三阶段构建，含 chrome-headless-shell，GitHub Actions 自动推送 GHCR

## 快速开始（Docker + GHCR 远端镜像）

### Windows（PowerShell）

```powershell
docker run -d --name hltv-mcp `
  -p 8082:8082 `
  -v hltv-chrome-data:/tmp `
  ghcr.io/arcdent/hltv-data:latest
```

浏览器访问 `http://localhost:8082`。

### Linux / macOS / WSL

```bash
docker run -d --name hltv-mcp \
  -p 8082:8082 \
  -v hltv-chrome-data:/tmp \
  ghcr.io/arcdent/hltv-data:latest
```

### 端口与进程管理

默认端口 `8082`（通过 `HTTP_PORT` 环境变量可修改）。

```bash
# 按端口杀进程
kill $(lsof -t -i:8082) 2>/dev/null || fuser -k 8082/tcp

# 按进程名杀
pkill -f hltv-mcp
```

### 自动同步

每次 push 到 main 分支，GitHub Actions 自动构建镜像推送到 GHCR。搭配系统计划任务实现自动更新：

**Windows（PowerShell 计划任务，以管理员运行）**

```powershell
$action = New-ScheduledTaskAction -Execute "docker" -Argument "run --rm -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp ghcr.io/arcdent/hltv-data:latest"
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Minutes 5) -RepetitionDuration (New-TimeSpan -Days 3650)
Register-ScheduledTask -TaskName "HLTV-Auto-Update" -Action $action -Trigger $trigger -RunLevel Highest
```

**Linux（crontab）**

```bash
*/5 * * * * docker pull ghcr.io/arcdent/hltv-data:latest && docker rm -f hltv-mcp && docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp ghcr.io/arcdent/hltv-data:latest
```

## 用法

### REST API

```bash
curl http://localhost:8082/api/health              # {"status":"ok"}
curl http://localhost:8082/api/status              # 服务状态
curl http://localhost:8082/api/matches/today       # 今日赛程
curl http://localhost:8082/api/search?q=Vitality&type=team
curl http://localhost:8082/api/news/realtime?limit=10
```

### MCP 工具列表

| 工具名 | 作用 | 主要参数 |
|--------|------|---------|
| `resolve_team` | 解析队伍名称为 HLTV 身份候选 | `name`(必填), `exact`, `limit` |
| `resolve_player` | 解析选手名称为 HLTV 身份候选 | `name`(必填), `exact`, `limit` |
| `hltv_team_recent` | 查询队伍近况、近期战绩和即将到来的比赛 | `team_id`, `team_name`, `limit` |
| `hltv_player_recent` | 查询选手近况和统计数据 | `player_id`, `player_name`, `limit` |
| `hltv_results_recent` | 查询近期赛果（支持队伍/赛事筛选） | `team`, `event`, `limit`(1-20), `days`(1-30) |
| `hltv_matches_upcoming` | 查询即将到来的比赛 | `team_id`, `team`, `event`, `limit`(1-20), `days`(1-30) |
| `hltv_matches_today` | 查询今日全部赛程（亚洲时区） | 无参数 |
| `hltv_realtime_news` | 获取 HLTV 实时/最新新闻 | `limit`(1-50), `page`, `offset` |
| `hltv_news_digest` | 获取 HLTV 月度归档新闻 | `limit`, `tag`, `year`, `month`, `page` |

### MCP 注册

Docker 部署后 MCP stdio 不可用（容器隔离）。如需 MCP 功能，使用手动编译启动（见下方）。

**标准 MCP 客户端**（Claude Desktop、VS Code Copilot、Gemini CLI 等）：

```jsonc
{
  "mcpServers": {
    "hltv": {
      "command": "/path/to/hltv-mcp",
      "args": []
    }
  }
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HTTP_PORT` | `8082` | HTTP 监听端口 |
| `HTTP_HOST` | `0.0.0.0` | HTTP 监听地址 |
| `HLTV_DATA_SOURCE` | `auto` | `auto` / `direct` / `chromedp` |
| `HLTV_CHROME_PATH` | 自动查找 | Chrome/Chromium 路径 |
| `HLTV_HTTP_TIMEOUT_MS` | `8000` | HTTP 超时（毫秒） |
| `HLTV_RETRY_COUNT` | `2` | HTTP 重试次数 |
| `DEFAULT_RESULT_LIMIT` | `5` | 默认查询结果数 |

完整配置见 `internal/config/config.go`。

## 项目结构

```
├── main.go                    # MCP stdio + HTTP 双 goroutine 入口
├── Dockerfile                 # 三阶段构建
├── internal/
│   ├── types/         # 共享类型定义
│   ├── errors/        # AppError 错误体系（8 错误码）
│   ├── config/        # 环境变量配置
│   ├── crypto/        # AES-256-GCM 加解密（API Key 持久化）
│   ├── cache/         # 内存缓存（TTL + stale + 并发合并）
│   ├── client/        # HTTP 客户端 + chromedp 反 CF
│   ├── scraper/       # 6 个 HLTV 爬虫模块
│   ├── localization/  # 中英文名称映射（26 队伍 + 98 选手）
│   ├── normalizer/    # HTML → 标准化数据结构
│   ├── facade/        # 核心编排层
│   ├── summary/       # 中文摘要
│   ├── renderer/      # 中文格式化输出
│   ├── mcp/           # 9 MCP 工具注册 + stdio 传输
│   └── http/          # chi router + REST API + SPA fallback
├── frontend/          # React + Vite + Tailwind
│   └── src/pages/     # 6 个管理面板页面
```

## 测试

```bash
go test github.com/arcdent/hltv-mcp/internal/... -v -timeout 30s
```

## 手动构建部署

### WSL / Linux 直接编译

```bash
git clone https://github.com/ArcDent/HLTV-data.git
cd HLTV-data

# 安装依赖
sudo apt install -y golang-go nodejs npm

# 构建前端
cd frontend && npm install && npm run build && cd ..

# 编译 Go
go build -o hltv-mcp github.com/arcdent/hltv-mcp

# 启动（指定 Chrome 路径或降级为纯 HTTP）
HLTV_CHROME_PATH=$(which chromium-browser) ./hltv-mcp
# 若无 Chrome：HLTV_DATA_SOURCE=direct ./hltv-mcp
```

### Docker 从源码构建

```bash
git clone https://github.com/ArcDent/HLTV-data.git
cd HLTV-data
docker build -t hltv-mcp .
docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp hltv-mcp
```

## 最近更新

### 2026-05-29

- **HLTV 爬虫适应新布局**：HLTV 选手页正在进行 A/B 式改版（部分选手已移除 `.all-time-stat` 生涯统计区域），新增 `NormalizeCareerFromOverview` 解析 `/stats/players/` 统计页数据；前端修复 React 条件渲染零值陷阱（`{0 && <Component/>}` 渲染为文字 "0"）

### 2026-05-28

- **前端别名编辑 UX 改进**：TeamDetail / PlayerDetail 去掉 ✏️ 铅笔图标，改为点击别名文字本身触发编辑；PlayerDetail 的国籍、年龄、队伍信息移至别名同一行显示
- **代码深度收敛**：删除 6 个文件（`events.go`、`news_article.go` ×2、`nicknames.json`、`Teams.tsx`、`Players.tsx`），合并前端搜索页为 `SearchPage`；废弃 `match_command_parse` MCP 工具（MCP 工具总数 10 → 9）
- **chromedp 修复**：`chromedp/headless-shell`（Chromium 148）传入 `--headless` 会导致 Chrome 无法启动，修复为 `Flag("headless", false)` 覆盖默认值，并添加 10s `NewContext` 超时保护
- **Docker SSL 修复**：`chromedp/headless-shell` 基础镜像不含 `ca-certificates`，导致 Go HTTP 客户端 TLS 验证失败，添加 `apt-get install ca-certificates`
- **昵称字典后端迁移**：新增 `overrides.go` 持久化覆盖层 + REST API（`GET/PUT /api/nicknames*`），前端硬编码昵称字典全部迁移至后端

## 灵感来源

本项目是对 [hltv-api](https://github.com/M3MONs/hltv-api) 的 Go 语言完全重建。将原 TypeScript MCP 服务（基于 [hltv-api](https://github.com/M3MONs/hltv-api) Python 爬虫 API 构建）统一为 Go 单一二进制，去掉外部 Python 依赖，保留全部 10 个 MCP 工具和中文本地化体系，并增加 React Web 管理面板。
