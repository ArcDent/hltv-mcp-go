# 选手详情卡片设计

## 目标

选手搜索结果点击后弹出详情卡片，展示从 HLTV 选手页面（chromedp）抓取的完整数据。无数据时布局自适应收缩。

## 数据来源

HLTV 选手详情页（如 `/player/11893/zywoo`），**必须用 chromedp 抓取**（JS 渲染页面，HTTP 直连只能拿到空壳）。

### 抓取字段与 HTML 选择器

| 字段 | HLTV 选择器 | 示例值 | 缺失时 |
|------|-----------|--------|--------|
| 游戏昵称 | `.playerNickname` | ZywOo | 必填，缺失则卡片不展示 |
| 真实姓名 | `.playerRealname` 或 `.playerNickname` 后的 text | Mathieu Herbaut | 隐藏该行 |
| 国籍 | `img.flag` 的 `title` 属性 | France | 隐藏 |
| 年龄 | `.playerAge .listRight` 或包含 "Age" 的 span | 25 | 隐藏 |
| 当前队伍 | `.playerTeam a` | Vitality | 隐藏该标签 |
| 奖金 | 含 "Prize money" 的 `.listRight` | $1,839,326 | 隐藏 |
| Rating 3.0 | `.player-stat` 中含 "Rating" 的 `.statsVal` | 1.37 | 对应条变灰 |
| 能力评分(8项) | `.player-stat` + `.statsVal` | 98/100 ... | 缺失项灰色 |
| Top 20 排名 | `.playerInfo` 或 profile 文字中正则提取 `#N('YY)` | #1('19)... | 整个 row 隐藏 |
| 生涯 Rating | `.all-time-stat` + `.stat`，含 "KDR" 上下文中取 rating | 1.29 | 显示 `—` |
| 总比赛 | `.all-time-stat` 含 "Matches" 的 `.stat` | 4734 | 显示 `—` |
| 胜率 | `.all-time-stat` 含 "Win rate" 的 `.stat` | 70.0% | 显示 `—` |
| K/D | `.all-time-stat` 含 "KDR" 的 `.stat` | 1.7 | 显示 `—` |
| 爆头率 | `.all-time-stat` 含 "Headshots" 的 `.stat` | 52.0% | 显示 `—` |
| Major 数 | `.highlighted-stat` 含 "Majors won" 的 `.stat` | 3 | 隐藏 |
| MVP 数 | `.highlighted-stat` 含 "Total MVPs" 的 `.stat` | 32 | 隐藏 |
| 近期比赛 | `.matches-table tbody tr` 或 `.recent-matches .match-row` | 7 行 | 隐藏整个 section |

### 抓取策略

1. 选手搜索 → 返回列表（已有，HTTP 直连即可）
2. 点击某条 → 前端请求 `GET /api/players/:id` → Go 后端用 chromedp 抓取详情页
3. 结果缓存 300s（同现有 player cache TTL）
4. chromedp 超时 30s，失败返回已有字段 + `partial: true` 标记

### Fallback 分级

| 级别 | 条件 | 展示 |
|------|------|------|
| **完整** | chromedp 成功 + 所有字段 | 全部 section |
| **部分** | chromedp 成功 + 部分字段缺失 | 隐藏缺失 section，能力评分缺失项灰色 |
| **最小** | chromedp 失败 / 超时 | 仅显示搜索页已有数据（名字+ID+slug），标记"详情暂时不可用" |
| **无数据** | 选手搜索接口返回空 | 不展示卡片 |

## 后端新增

### API: `GET /api/players/:id`

请求：`GET /api/players/11893`

响应：
```json
{
  "query": {"player_id": 11893},
  "data": {
    "profile": {
      "id": 11893,
      "name": "ZywOo",
      "real_name": "Mathieu Herbaut",
      "slug": "zywoo",
      "country": "France",
      "age": 25,
      "team": "Vitality",
      "prize_money": "$1,839,326"
    },
    "rating": {"value": 1.37, "maps": 49},
    "abilities": [
      {"key": "rating",      "label_en": "Rating",     "label_zh": "综合",   "value": 1.37, "format": "decimal"},
      {"key": "firepower",   "label_en": "Firepower",  "label_zh": "火力",   "value": 98, "max": 100},
      {"key": "opening",     "label_en": "Opening",    "label_zh": "突破",   "value": 84, "max": 100},
      {"key": "clutching",   "label_en": "Clutching",  "label_zh": "残局",   "value": 88, "max": 100},
      {"key": "sniping",     "label_en": "Sniping",    "label_zh": "狙击",   "value": 91, "max": 100},
      {"key": "entrying",    "label_en": "Entrying",   "label_zh": "进点",   "value": 30, "max": 100},
      {"key": "trading",     "label_en": "Trading",    "label_zh": "补枪",   "value": 55, "max": 100},
      {"key": "utility",     "label_en": "Utility",    "label_zh": "道具",   "value": 64, "max": 100}
    ],
    "career": {
      "rating": 1.29,
      "matches": 4734,
      "win_rate": "70.0%",
      "kd": 1.7,
      "headshot_pct": "52.0%",
      "win_streak": 26
    },
    "top20_ranks": {"2019":1,"2020":1,"2021":2,"2022":2,"2023":1,"2024":3,"2025":1},
    "honors": [
      {"label": "Major 冠军", "value": 3},
      {"label": "Major MVP",  "value": 3},
      {"label": "总 MVP",     "value": 32},
      {"label": "年度 #1",    "value": 4}
    ],
    "recent_matches": [
      {
        "date": "05/16", "team": "Vitality", "opponent": "Natus Vincere",
        "score": "1:2", "result": "loss", "rating": 1.52, "kills": 78, "deaths": 52,
        "event": "IEM Atlanta 2026"
      }
    ]
  },
  "meta": {"fetched_at": "...", "partial": false}
}
```

### 实现文件

- `internal/http/handlers/` 新增 `player_detail.go`（扩展现有桩 `GET /api/players/{id}`，替换 `"not yet implemented"`）
- `internal/scraper/player.go` 新增 `GetPlayerDetail(ctx, id, slug)` 方法（chromedp 抓取）
- `internal/normalizer/player.go` 新增 `NormalizePlayerDetail(doc)` 方法（HTML → struct）

## 前端

### 交互

- 搜索结果列表 **保持不变**（名字 + ID + slug）
- 点击某条 → 居中弹出详情卡片（遮罩 + slideUp 动画，复用现有 modal 样式）
- 卡片内部根据 `data.profile`、`data.abilities`、`data.career`、`data.recent_matches` 是否存在决定显示哪些 section
- 点击遮罩或 ✕ 关闭

### 卡片布局（与预览一致）

```
┌─ ◎ ZywOo ──────────────────────────────────────┐
│ [Z] ZywOo                    🇫🇷 France Age 25  │
│     Mathieu Herbaut           Vitality 💰 $1.8M │
│                                                  │
│ [2019#1] [2020#1] [2021#2] ... [2025#1]        │
│                                                  │
│ 能力评分(近3月·49maps)                            │
│     [六维雷达图]    Rating  1.37                 │
│                    Firepower  98/100              │
│                    ...                           │
│                                                  │
│ 生涯 Rating 1.29 | 4734 比赛 | 70% 胜率 | 1.7KD  │
│ 🏆3×Major ⭐3×MajorMVP 🥇32×MVP 🔝4×#1 ...      │
│                                                  │
│ 近期7场比赛                                       │
│ 05/16 1.52 Vitality vs NaVi    1:2  78-52       │
│ ...                                              │
└──────────────────────────────────────────────────┘
```

### Fallback 展示逻辑

- `top20_ranks` 为空或 `{}` → 整个 Top 20 pill 行隐藏
- `abilities` 某条缺失 → 该项显示灰色 dot + 灰色文字 + `—`
- `career` 某字段为 0 或空 → 显示 `—`
- `honors` 为空 → 整个荣誉行隐藏
- `recent_matches` 为空 → 整个 section 隐藏
- `meta.partial: true` → 卡片底部显示 "部分数据不可用" 提示

## 技术约束

- 选手详情页必须 chromedp 抓取，复用现有反 CF 配置（UserDataDir + stealth flags）
- 单个选手详情缓存 300s
- chromedp 超时 30s
- 不新增 npm 依赖
- Go 新增文件 ≤ 2 个（handler + normalizer 扩展）
- 前端修改文件：`SearchableList.tsx`（点击展开）、`News.tsx` 不受影响
