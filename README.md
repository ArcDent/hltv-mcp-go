# HLTV MCP Service

Go 单二进制全栈 HLTV MCP 服务 — MCP stdio + HTTP REST + React 管理面板。

## 功能特性

- **10 个 MCP 工具**：队伍/选手解析、赛程/赛果查询、实时/归档新闻
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

### MCP 注册（搭配 Docker）

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
│   ├── mcp/           # 10 MCP 工具注册 + stdio 传输
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
