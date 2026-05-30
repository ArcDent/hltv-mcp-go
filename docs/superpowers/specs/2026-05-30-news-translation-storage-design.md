# 新闻翻译结果长效化存储设计

## 目标

在现有三级缓存架构（Cache → SQLite → HLTV）基础上，增设新闻标题和正文翻译结果的长效化存储，避免重复翻译，降低 LLM token 消耗。

## 触发策略

- **标题**：入库后立即异步翻译（`BatchUpsertNews` / `BatchUpsertRealtimeNews` 后启动 goroutine）
- **正文**：用户在前端点击翻译时按需翻译，结果写库

## Schema 变更

### migration v2

```sql
ALTER TABLE news ADD COLUMN title_zh TEXT;
ALTER TABLE news ADD COLUMN body_text_zh TEXT;
ALTER TABLE realtime_news ADD COLUMN title_zh TEXT;
```

- 新列为 `TEXT`，默认 `NULL`
- v2 只加列，不建索引

### types 变更

| 类型 | 新增字段 | JSON tag |
|---|---|---|
| `NewsItem` | `TitleZh string` | `title_zh,omitempty` |
| `NewsArticle` | `TitleZh string`, `BodyTextZh string` | `title_zh,omitempty`, `body_text_zh,omitempty` |
| `RealtimeNewsItem` | `TitleZh string` | `title_zh,omitempty` |

- 所有新增字段使用 `omitempty`，未翻译时不出现在 JSON 中

## 新增文件

### `internal/translator/translator.go`

```go
type Translator struct {
    providerURL string
    apiKey      string
    model       string
    client      *http.Client
}

func New(cfg TranslateConfig) *Translator

func (t *Translator) TranslateTitle(ctx context.Context, text string) (string, error)
func (t *Translator) TranslateBody(ctx context.Context, text string) (string, error)
```

- 翻译逻辑从 `handlers/translate.go` 的 `PostTranslate` 中提取
- 标题 prompt：`"将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释"`
- 正文 prompt：`"将以下CS电竞新闻正文翻译为简体中文"`
- 无状态、不缓存、不排队，每次调用独立
- `TranslateConfig` 复用 `handlers` 包已有结构体

## 变更文件

### `internal/storage/news.go`

新增三个单列更新方法：

```go
func (s *Store) UpdateNewsTitleZh(url string, titleZh string) error
func (s *Store) UpdateNewsBodyZh(url string, bodyZh string) error
func (s *Store) UpdateRealtimeTitleZh(url string, titleZh string) error
```

保留方法变更：`GetNewsArticle` / `QueryNews` / `QueryRealtimeNews` 的 SQL 和 Scan 增加新列。

### `internal/facade/facade.go`

- `HltvFacade` 新增字段：`translateCfgFn func() (TranslateConfig, error)`
- 新增方法 `translateNewTitles(items []types.NewsItem)` 和 `translateNewRealtimeTitles(items []types.RealtimeNewsItem)`
- `translateCfgFn` 为 nil 时静默跳过（翻译功能未配置）

### `internal/facade/news.go`

- `GetRealtimeNews` 的 `compute` 函数：`BatchUpsertRealtimeNews` 后 `go f.translateNewRealtimeTitles(allItems)`
- `GetNewsDigest` 的 `compute` 函数：`BatchUpsertNews` 后 `go f.translateNewTitles(allItems)`

### `internal/http/handlers/translate.go`

- `PostTranslate`：改用 `translator.Translator` 代理翻译
- 新增可选请求字段 `url`：非空时翻译完成后将结果写库（`UpdateNewsBodyZh`）
- `PutTranslateConfig`：不变

### `internal/storage/migration.go`

- 新增 `applyV2`，追加三个 ALTER TABLE 语句

### `main.go`

- 初始化 `translator.Translator`，注入 `HltvFacade` 和 `Handlers`

## 数据流

### 标题自动翻译

```
GetRealtimeNews / GetNewsDigest
  → scrape HLTV
  → BatchUpsert* (写入 title/link 等)
  → go translateNewTitles (逐条: TranslateTitle → UpdateXxxTitleZh)
  → 返回 items (本次不包含翻译结果，下次请求带出)
```

### 正文按需翻译

```
POST /api/translate {"text": "...", "type": "body", "url": "https://..."}
  → Translator.TranslateBody
  → UpdateNewsBodyZh (url 非空时)
  → 返回 translated text
```

### 读取

```
GET /api/news/article?url=...
  → Cache → SQLite (已含 title_zh/body_text_zh) → HLTV
  → JSON 自动带出 title_zh/body_text_zh (omitempty)
```

## 配置热加载

facade 不持有 Translator 实例，改为持有 `func() (TranslateConfig, error)` 工厂函数。后台翻译时调用工厂获取最新配置，创建临时 Translator。`PUT /api/translate/config` 修改配置后下次翻译自动生效。

## 降级与容错

- `translate_config.json` 不存在 → `translateCfgFn` 为 nil，自动翻译和存储静默跳过
- 单条翻译失败 → log + continue，不影响批处理中其他条目
- 写库失败 → log，翻译结果已返回用户，下次会重新翻译

## 前端

无需改动。API 自动返回 `title_zh`/`body_text_zh`，React 组件按 JSON 渲染。
