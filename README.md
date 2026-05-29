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
  -e FIRECRAWL_API_KEY=fc-xxxxxxxxxxxxxxxx `
  -v hltv-chrome-data:/tmp `
  ghcr.io/arcdent/hltv-data:latest
```

浏览器访问 `http://localhost:8082`。

> `FIRECRAWL_API_KEY` 用于绕过 Cloudflare 抓取赛程（/matches），不配置不影响搜索/队伍/选手功能。

### Linux / macOS / WSL

```bash
docker run -d --name hltv-mcp \
  -p 8082:8082 \
  -e FIRECRAWL_API_KEY=fc-xxxxxxxxxxxxxxxx \
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

### 更新镜像

```bash
# 拉取最新镜像
docker pull ghcr.io/arcdent/hltv-data:latest

# 停止旧容器 → 删除 → 启动新容器
docker rm -f hltv-mcp \
  && docker run -d --name hltv-mcp \
    -p 8082:8082 \
    -v hltv-chrome-data:/tmp \
    ghcr.io/arcdent/hltv-data:latest

# 一行更新
docker pull ghcr.io/arcdent/hltv-data:latest && docker rm -f hltv-mcp && docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp ghcr.io/arcdent/hltv-data:latest
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
| `FIRECRAWL_API_KEY` | — | Firecrawl API Key（绕过 CF 封锁赛程抓取） |
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

- **搜索页路由切换修复**：修复队伍/选手搜索页面切换时 state 不刷新的 bug（React Router 复用相同组件类型导致），`SearchableList` 添加 `key={type}` 强制重新挂载；修正 Go embed 指令 `dist/*` → `dist` 确保 `dist/assets/` 子目录静态资源嵌入
- **Firecrawl 集成 + 赛程全面恢复**：通过 Firecrawl API 绕过 HLTV Cloudflare 封锁，`/matches` 端点恢复正常，覆盖范围从今日到 IEM Cologne Major 2026 主赛事（6 月 11 日），共 67 场比赛；重写 `NormalizeUpcomingMatches` 支持多个 `.matches-list-section` 容器；新增 `FIRECRAWL_API_KEY` 环境变量；匹配详情增加 `GoFrame` 兼容性修复
- **HLTV Cloudflare 封锁修复**：HLTV 全面启用 Cloudflare 防护，`/matches`、`/results` 等页面返回 403 Challenge；修复 `NormalizeUpcomingMatches` nil pointer panic；所有 HTTP handler 添加超时保护
- **选手数据分层修复**：HLTV 新版选手页无 `.all-time-stat`，代码改为先行尝试旧版生涯战斗统计，再回退到 `.highlighted-stat` 提取生涯概览；3 月 Rating 与生涯统计明确分离不再混淆；前端新增 `生涯概览` 网格和 `StatBadge` 组件
- **HLTV 爬虫适应新布局**：HLTV 选手页正在进行 A/B 式改版，前端修复 React 条件渲染零值陷阱

### 2026-05-28

- **前端别名编辑 UX 改进**：TeamDetail / PlayerDetail 去掉 ✏️ 铅笔图标，改为点击别名文字本身触发编辑；PlayerDetail 的国籍、年龄、队伍信息移至别名同一行显示
- **代码深度收敛**：删除 6 个文件（`events.go`、`news_article.go` ×2、`nicknames.json`、`Teams.tsx`、`Players.tsx`），合并前端搜索页为 `SearchPage`；废弃 `match_command_parse` MCP 工具（MCP 工具总数 10 → 9）
- **chromedp 修复**：`chromedp/headless-shell`（Chromium 148）传入 `--headless` 会导致 Chrome 无法启动，修复为 `Flag("headless", false)` 覆盖默认值，并添加 10s `NewContext` 超时保护
- **Docker SSL 修复**：`chromedp/headless-shell` 基础镜像不含 `ca-certificates`，导致 Go HTTP 客户端 TLS 验证失败，添加 `apt-get install ca-certificates`
- **昵称字典后端迁移**：新增 `overrides.go` 持久化覆盖层 + REST API（`GET/PUT /api/nicknames*`），前端硬编码昵称字典全部迁移至后端

## 灵感来源

本项目是对 [hltv-api](https://github.com/M3MONs/hltv-api) 的 Go 语言完全重建。将原 TypeScript MCP 服务（基于 [hltv-api](https://github.com/M3MONs/hltv-api) Python 爬虫 API 构建）统一为 Go 单一二进制，去掉外部 Python 依赖，保留全部 10 个 MCP 工具和中文本地化体系，并增加 React Web 管理面板。
