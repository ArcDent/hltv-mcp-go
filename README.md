# HLTV MCP Service

Go 单二进制全栈 HLTV MCP 服务 — MCP stdio + HTTP REST + React 管理面板，一键启动。

> 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API），使用 Go 完全重建，去除 Python 上游依赖。

## 功能特性

- **10 个 MCP 工具**：队伍/选手解析、赛程/赛果查询、实时/归档新闻，与原项目功能等价
- **Web 管理面板**：React SPA，6 个页面（Dashboard / Matches / Teams / Players / News / Cache）
- **反爬虫**：HTTP 直连优先 + chromedp 绕过 Cloudflare，5 分钟失败记忆窗口
- **中文输出**：26 支队伍的民间昵称映射，43+ 名选手中文简称，中文摘要生成
- **单二进制部署**：Go 编译 + React 内嵌，MCP stdio 和 HTTP 同时运行
- **Docker 一键**：三阶段构建，含 chrome-headless-shell

## 快速开始

### 通用前置条件

- **Go** >= 1.26
- **Node.js** >= 18（仅直接编译时需要，Docker 不需要）
- **Chrome/Chromium**（chromedp fallback 需要，可缺失——启动时自动检测并降级为 HTTP-only 模式）

### 直接编译部署

```bash
# 1. 克隆仓库
git clone https://github.com/ArcDent/HLTV-data.git
cd HLTV-data

# 2. 构建前端
cd frontend && npm install && npm run build && cd ..

# 3. 编译 Go（前端产物自动内嵌）
go build -o hltv-mcp github.com/arcdent/hltv-mcp

# 4. 启动
./hltv-mcp
```

启动后访问 `http://localhost:8082` 打开管理面板。

### 端口与进程管理

**默认端口**：`8082`（通过 `HTTP_PORT` 环境变量可修改）。

**强制关闭服务**：

```bash
# 方式一：按端口杀进程（推荐）
kill $(lsof -t -i:8082) 2>/dev/null || fuser -k 8082/tcp

# 方式二：按进程名杀
pkill -f hltv-mcp

# 方式三：查找 PID 后手动 kill
ps aux | grep hltv-mcp | grep -v grep
kill <PID>
```

**确认服务已停止**：
```bash
curl http://localhost:8082/api/health  # 应返回 Connection refused
```

#### Linux

```bash
# 安装依赖
sudo apt install -y golang-go nodejs npm chromium-browser

# 编译启动（同上）
cd frontend && npm install && npm run build && cd ..
go build -o hltv-mcp github.com/arcdent/hltv-mcp
HLTV_CHROME_PATH=$(which chromium-browser) ./hltv-mcp
```

#### WSL (Windows Subsystem for Linux)

WSL 内直接编译，前端在 WSL 内构建，Go 交叉编译在 WSL 内完成：

```bash
# WSL 安装依赖
sudo apt install -y golang-go nodejs npm

# 编译
cd frontend && npm install && npm run build && cd ..
go build -o hltv-mcp github.com/arcdent/hltv-mcp

# Chrome 路径（WSL 使用宿主 Windows 的 Chrome）
HLTV_CHROME_PATH="/mnt/c/Program Files/Google/Chrome/Application/chrome.exe" ./hltv-mcp

# 若宿主机没有 Chrome，降级为纯 HTTP 模式
HLTV_DATA_SOURCE=direct ./hltv-mcp
```

WSL2 网络说明：MCP stdio 客户端在 WSL 内直接通信不受影响。Web 面板从宿主机浏览器访问 `http://localhost:8082` 即可。

#### macOS

```bash
# 安装依赖（Homebrew）
brew install go node chromium

# 编译启动
cd frontend && npm install && npm run build && cd ..
go build -o hltv-mcp github.com/arcdent/hltv-mcp
HLTV_CHROME_PATH=$(which chromium) ./hltv-mcp
```

#### Windows（原生，非 WSL）

推荐使用 PowerShell 或 Git Bash：

```powershell
# 1. 安装依赖
# Go:    https://go.dev/dl/  （下载 .msi 安装，勾选 Add to PATH）
# Node:  https://nodejs.org/ （下载 LTS .msi 安装）
# Chrome: 通常已预装，路径为 C:\Program Files\Google\Chrome\Application\chrome.exe

# 2. 克隆仓库
git clone https://github.com/ArcDent/HLTV-data.git
cd HLTV-data

# 3. 构建前端
cd frontend; npm install; npm run build; cd ..

# 4. 编译 Go（前端产物自动内嵌）
go build -o hltv-mcp.exe github.com/arcdent/hltv-mcp

# 5. 启动
set HLTV_CHROME_PATH=C:\Program Files\Google\Chrome\Application\chrome.exe
hltv-mcp.exe
```

若系统无 Chrome，降级为纯 HTTP 模式（无需浏览器）：

```powershell
set HLTV_DATA_SOURCE=direct
hltv-mcp.exe
```

> **注意**：Windows 原生下 MCP stdio 模式需要客户端支持。OpenCode / Claude Code 等工具通常通过 WSL 或直接调用 exe，路径写 `C:\path\to\hltv-mcp.exe` 即可。

### Docker 部署

> **前置条件**：Windows/macOS 需先启动 Docker Desktop 并等待引擎就绪（右下角图标变绿）。

#### Linux / macOS / WSL

```bash
# 方式一：从源码构建
git clone https://github.com/ArcDent/HLTV-data.git
cd HLTV-data
docker build -t hltv-mcp .
docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp hltv-mcp

# 方式二：使用 GHCR 预构建镜像
docker run -d --name hltv-mcp \
  -p 8082:8082 \
  -v hltv-chrome-data:/tmp \
  ghcr.io/arcdent/hltv-data:latest

# 方式三：docker compose（GHCR 预构建镜像 + 自动拉取）
docker compose -f docker-compose.ghcr.yml up -d
```

#### Windows（PowerShell）

```powershell
# 确保 Docker Desktop 已启动

# 方式一：从源码构建
git clone https://github.com/ArcDent/HLTV-data.git
cd HLTV-data
docker build -t hltv-mcp .
docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp hltv-mcp

# 方式二：使用 GHCR 预构建镜像（PowerShell 用反引号续行，非 \）
docker run -d --name hltv-mcp `
  -p 8082:8082 `
  -v hltv-chrome-data:/tmp `
  ghcr.io/arcdent/hltv-data:latest

# 方式三：docker compose（GHCR 预构建镜像 + 自动拉取）
docker compose -f docker-compose.ghcr.yml up -d
```

> **注意**：Windows CMD 不支持 `\` 或 `` ` `` 续行，直接写一行即可。

浏览器访问 `http://localhost:8082`。

#### MCP 注册（搭配 Docker）

Docker 部署后 MCP 通过 stdio 不可用（容器隔离），如需 MCP 功能请使用直接编译方式。
仅 Web 面板和 REST API 走 Docker，MCP 功能需单独在本地编译启动。

### 预构建镜像与自动同步

每次 push 到 main 分支，GitHub Actions 自动构建镜像并推送到 GHCR。
推荐使用 `docker compose` + 系统计划任务实现自动同步（不依赖第三方容器）。

#### 自动同步配置

使用 `docker-compose.ghcr.yml`（项目内置），已配置 `pull_policy: always`：

**Linux（crontab）**

```bash
# 每 5 分钟检查更新，有则拉取重建
*/5 * * * * cd /path/to/HLTV-data && docker compose -f docker-compose.ghcr.yml up -d --pull always
```

**Windows（PowerShell 计划任务，以管理员运行）**

```powershell
$action = New-ScheduledTaskAction -Execute "docker" -Argument "compose -f D:\ArcSysFiles\HLTV-data\docker-compose.ghcr.yml up -d --pull always"
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Minutes 5) -RepetitionDuration (New-TimeSpan -Days 3650)
Register-ScheduledTask -TaskName "HLTV-Auto-Update" -Action $action -Trigger $trigger -RunLevel Highest
```

推送代码后流程：`git push` → GitHub Actions 构建镜像推送到 GHCR → 计划任务执行 `compose up --pull always` → 有新镜像则自动拉取重建。

> **注意**：GHCR 镜像默认私有，需先在 GitHub Package Settings 中改为 Public，或用 Personal Access Token 登录：`echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin`

## 用法示例

### MCP 工具

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

**OpenCode** 使用下方格式：

```jsonc
{
  "mcp": {
    "hltv_local": {
      "type": "local",
      "command": ["/path/to/hltv-mcp"],
      "enabled": true
    }
  }
}
```

### 工具列表

| 工具名 | 作用 | 主要参数 |
|--------|------|---------|
| `resolve_team` | 解析队伍名称为 HLTV 身份候选 | `name`(必填), `exact`, `limit` |
| `resolve_player` | 解析选手名称为 HLTV 身份候选 | `name`(必填), `exact`, `limit` |
| `hltv_team_recent` | 查询队伍近况、近期战绩和即将到来的比赛 | `team_id`, `team_name`, `limit`, `include_upcoming`, `include_recent_results` |
| `hltv_player_recent` | 查询选手近况和统计数据 | `player_id`, `player_name`, `limit` |
| `hltv_results_recent` | 查询近期赛果（支持队伍/赛事筛选） | `team_id`, `team`, `event`, `limit`(1-20), `days`(1-30) |
| `hltv_matches_upcoming` | 查询即将到来的比赛 | `team_id`, `team`, `event`, `limit`(1-20), `days`(1-30) |
| `hltv_matches_today` | 查询今日全部赛程（亚洲时区） | 无参数 |
| `match_command_parse` | 解析 `/match` 命令参数 | `raw_args` |
| `hltv_realtime_news` | 获取 HLTV 实时/最新新闻 | `limit`(1-50), `page`, `offset` |
| `hltv_news_digest` | 获取 HLTV 月度归档新闻 | `limit`, `tag`, `year`, `month`, `page` |

示例调用：

```text
调用 hltv_local_resolve_team，搜索 Vitality
调用 hltv_local_hltv_team_recent，传 team_id=9565
调用 hltv_local_hltv_matches_today
调用 hltv_local_hltv_realtime_news
```

### REST API

```bash
curl http://localhost:8082/api/health          # {"status":"ok"}
curl http://localhost:8082/api/status          # 服务状态
curl http://localhost:8082/api/matches/today   # 今日赛程
curl http://localhost:8082/api/search?q=Vitality&type=team
curl http://localhost:8082/api/news/realtime?limit=10
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
| `SUMMARY_MODE` | `template` | `template` / `raw` |

完整配置见 `internal/config/config.go`。

## 项目结构

```
├── main.go                    # MCP stdio + HTTP 双 goroutine 入口
├── Dockerfile                 # 三阶段构建
├── internal/
│   ├── types/         # 共享类型定义
│   ├── errors/        # AppError 错误体系
│   ├── config/        # 环境变量配置
│   ├── cache/         # 内存缓存（TTL + stale + 并发合并）
│   ├── client/        # HTTP 客户端 + chromedp 反 CF
│   ├── scraper/       # 6 个 HLTV 爬虫模块
│   ├── localization/  # 中英文名称映射
│   ├── normalizer/    # HTML → 标准化数据结构
│   ├── facade/        # 核心编排层
│   ├── summary/       # 中文摘要
│   ├── renderer/      # 中文格式化输出
│   ├── mcp/           # MCP 工具注册 + stdio 传输
│   └── http/          # chi router + REST API
├── frontend/          # React + Vite + Tailwind
│   └── src/pages/     # 6 个管理面板页面
├── cmd/               # 调试/验证工具
└── docs/superpowers/  # 设计文档
```

## 依赖/环境要求

- **Go** >= 1.26
- **Node.js** >= 18（仅构建前端时需要）
- **Chrome/Chromium**（chromedp fallback 需要，可选）
- **Docker**（可选，用于容器化部署）

## 测试

```bash
go test github.com/arcdent/hltv-mcp/internal/... -v -timeout 30s
```

## 灵感来源

本项目是对 [hltv-api](https://github.com/M3MONs/hltv-api) 的 Go 语言完全重建。本重建将原 TypeScript MCP 服务（基于 [hltv-api](https://github.com/M3MONs/hltv-api) Python 爬虫 API 构建）统一为 Go 单一二进制，去掉外部 Python 依赖，保留了全部 10 个 MCP 工具和中文本地化体系，并增加了 React Web 管理面板。
