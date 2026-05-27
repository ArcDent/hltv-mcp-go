# 新闻翻译组件设计

## 目标

为新闻页面添加中英双语展示——英文原标题 + 下方中文翻译，翻译由用户自定义的 OpenAI 兼容 API 完成。

## 数据流

**配置存储**：Go 后端 `/api/translate/config`

```
首次使用 → 前端弹出配置面板 → 填写 API 地址/Key/模型 → PUT /api/translate/config → 后端写入文件
后续使用 → 页面加载 → GET /api/translate/config → 后端返回配置 → 前端使用
```

**翻译流程**：

```
页面加载新闻列表 → GET /api/translate/config 获取配置 →
  每条英文标题 →
    查 localStorage 翻译缓存（key = title_md5_hash）→
      命中 → 直接显示中文
      未命中 → POST {provider}/v1/chat/completions
                { model, messages: [
                    {role:"system", content:"将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果"},
                    {role:"user", content: title}
                ]}
                → 解析 choices[0].message.content
                → 写入 localStorage 缓存
                → 显示中文
```

翻译缓存结构（localStorage key: `hltv_translations`）：
```json
{
  "a1b2c3...": { "zh": "FaZe 成为首批 XSE 职业联赛受邀队伍", "ts": 1716854400 },
  "d4e5f6...": { "zh": "Fluxo 确认签下 dav1deuS 和 Ltz", "ts": 1716854400 }
}
```

同一标题在缓存有效期（7 天）内不重复请求翻译 API。

## 配置存储（Go 后端）

新增 `internal/http/handlers/translate.go`，两个端点：

**GET `/api/translate/config`**
```json
{
  "provider_url": "https://api.openai.com/v1",
  "api_key": "sk-••••••••",
  "model": "gpt-4o-mini",
  "configured": true
}
```

**PUT `/api/translate/config`**
```json
// Request body
{
  "provider_url": "https://api.deepseek.com/v1",
  "api_key": "sk-xxxx",
  "model": "deepseek-chat"
}
// Response: {"status": "saved"}
```

配置存储为项目目录下的 `translate_config.json` 文件（gitignore），服务重启后配置不丢失。API Key 返回时前三段保持、后三段替换为 `****`。配置文件路径为二进制同目录下的 `translate_config.json`。Docker 部署时通过 volume 挂载持久化：

```yaml
# docker-compose.yml
services:
  hltv-mcp:
    volumes:
      - ./translate_config.json:/translate_config.json
```

首次启动时文件不存在则自动创建默认空配置，用户通过前端面板填写后持久化到宿主机文件。（gitignore 中已忽略 `*.json` 以外的模式，需单独添加）。provider_url 为用户填写的完整 base URL（如 `https://api.openai.com/v1`），前端直接拼接 `/chat/completions`，不额外添加 `/v1` 路径段。

## 配置组件 UI

位置：新闻页标签栏右侧，齿轮按钮 ⚙

点击弹出配置面板（遮罩 + 居中卡片，复用 Dashboard 数据源弹窗风格）：

```
┌─ 翻译设置 ────────────────────────────────── ✕ ─┐
│                                                    │
│  API 地址                                          │
│  ┌────────────────────────────────────────────┐    │
│  │ https://api.openai.com/v1                 │    │
│  └────────────────────────────────────────────┘    │
│                                                    │
│  API Key                                           │
│  ┌────────────────────────────────────────────┐    │
│  │ ••••••••••••••••••••                       │    │
│  └────────────────────────────────────────────┘    │
│                                                    │
│  模型                                              │
│  ┌────────────────────────────────────────────┐    │
│  │ gpt-4o-mini                                │    │
│  └────────────────────────────────────────────┘    │
│                                                    │
│  预设 ▾                                            │
│  ├─ OpenAI          gpt-4o-mini                    │
│  ├─ DeepSeek        deepseek-chat                  │
│  ├─ Groq            llama-3.3-70b-versatile        │
│  └─ Ollama 本地     qwen2.5:7b                     │
│                                                    │
│  [保存]                    状态: ● 已配置 / ○ 未配置│
└────────────────────────────────────────────────────┘
```

预设选项自动填充 API 地址和模型名，用户可手动修改。API Key 为 password 类型输入框。配置通过 `PUT /api/translate/config` 保存到 Go 后端文件。

## 翻译展示

新闻条目卡片结构：

```
┌─ 编号 ┬ 原标题（英文）──────────────────┬ 日期 ─┐
│  01   │ FaZe, Legacy among first XSE.. │ 05-26 │
│       │ FaZe、Legacy 成为首批受邀队伍    │       │  ← 13px 灰色
└───────┴─────────────────────────────────┴───────┘
```

- 第一行：英文原标题 16px，正常字重
- 第二行：中文翻译 13px，`var(--text-muted)` 灰色，翻译中显示"翻译中..."（闪烁点动画）
- 翻译失败的条目回退显示原文，不追加错误行

## 技术约束

- 配置存储：Go 后端 `/api/translate/config`（GET/PUT），文件持久化
- 翻译执行：前端 fetch 直接调 OpenAI 兼容 API（不经过 Go 后端）
- 翻译缓存：localStorage，7 天有效期
- 翻译请求并发控制：最多同时 3 个 in-flight 请求
- 不引入新的 npm 依赖（fetch + localStorage 即足够）
- Go 新增文件：`internal/http/handlers/translate.go`
- 前端新增文件：`frontend/src/components/TranslateProvider.tsx`（配置面板）
- 前端修改文件：`frontend/src/pages/News.tsx`（翻译展示）
