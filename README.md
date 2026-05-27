# HLTV MCP Service

Go 单二进制全栈 HLTV MCP 服务 — MCP stdio + HTTP REST + React 管理面板，一键启动。

> 灵感来源：[hltv-api](https://github.com/M3MONs/hltv-api)（Python Flask/Scrapy HLTV 爬虫 API），使用 Go 完全重建，去除 Python 上游依赖。

## 功能特性

- **10 个 MCP 工具**：队伍/选手解析、赛程/赛果查询、实时/归档新闻，与原项目功能等价
- **Web 管理面板**：React SPA，6 个页面（Dashboard / Matches / Teams / Players / News / Cache）
- **反爬虫**：HTTP 直连优先 + chromedp 绕过 Cloudflare，5 分钟失败记忆窗口
- **中文输出**：26 支队伍 + 3 个赛事的中英民间昵称映射，中文摘要生成
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
git clone https://github.com/ArcDent/hltv-mcp-go.git
cd hltv-mcp-go

# 2. 构建前端
cd frontend && npm install && npm run build && cd ..

# 3. 编译 Go（前端产物自动内嵌）
go build -o hltv-mcp github.com/arcdent/hltv-mcp

# 4. 启动
./hltv-mcp
```

启动后访问 `http://localhost:8082` 打开管理面板。

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
git clone https://github.com/ArcDent/hltv-mcp-go.git
cd hltv-mcp-go

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

#### Linux / macOS

```bash
git clone https://github.com/ArcDent/hltv-mcp-go.git
cd hltv-mcp-go
docker build -t hltv-mcp .
docker run -d --name hltv-mcp -p 8082:8082 hltv-mcp
```

浏览器访问 `http://localhost:8082`。

如需持久化 Chrome profile（提升 chromedp 稳定性）：

```bash
docker run -d --name hltv-mcp \
  -p 8082:8082 \
  -v hltv-chrome-data:/tmp \
  hltv-mcp
```

#### WSL（Docker Desktop）

```bash
# 在 WSL 终端中执行（确保 Docker Desktop 已安装并启用 WSL integration）
git clone https://github.com/ArcDent/hltv-mcp-go.git
cd hltv-mcp-go
docker build -t hltv-mcp .
docker run -d --name hltv-mcp -p 8082:8082 hltv-mcp
```

从 Windows 宿主机浏览器访问 `http://localhost:8082`。

#### Windows（Docker Desktop）

```powershell
# PowerShell 或 CMD
git clone https://github.com/ArcDent/hltv-mcp-go.git
cd hltv-mcp-go
docker build -t hltv-mcp .
docker run -d --name hltv-mcp -p 8082:8082 hltv-mcp
```

浏览器访问 `http://localhost:8082`。

#### MCP 注册（搭配 Docker）

Docker 部署后 MCP 通过 stdio 不可用（容器隔离），如需 MCP 功能请使用直接编译方式。
仅 Web 面板和 REST API 走 Docker，MCP 功能需单独在本地编译启动。

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

然后调用：

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
