# 长期化存储设计规范

> 2026-05-29 | 参考项目: [person-summon](../person-summon)

## 目标

为 HLTV MCP 增加 SQLite 持久化存储层，实现三层数据回退（内存缓存 → SQLite 历史 → HLTV 实时抓取），配合 SSE 推送前端自动刷新。

## 数据流

```
请求 (GET /api/players/123)
  │
  ├─ Cache hit → 返回 (不变)
  │
  └─ Cache miss
       │
       ├─ SQLite hit → 立即返回 (stale=true)
       │     │
       │     └─ 后台 goroutine → 抓 HLTV → SQLite upsert → Cache set
       │                                                  │
       │                                          SSE broadcast → 前端重取
       │
       └─ SQLite miss → 同步抓 HLTV → SQLite upsert → Cache set → 返回 (fresh)
```

## 数据表设计（5 张表）

### teams
| 列 | 类型 | 说明 |
|---|---|---|
| `id` | INTEGER PRIMARY KEY | HLTV 队伍 ID |
| `name` | TEXT NOT NULL | 队伍名称 |
| `slug` | TEXT | URL slug |
| `country` | TEXT | 国家代码 |
| `rank` | INTEGER | 世界排名 |
| `stats_json` | TEXT | TeamStats JSON |
| `achievements_json` | TEXT | []TeamAchievement JSON |
| `roster_json` | TEXT | []TeamRosterPlayer JSON |
| `highlights_json` | TEXT | TeamHighlights JSON |
| `recent_matches_json` | TEXT | []NormalizedMatch JSON |
| `fetched_at` | TEXT | 抓取时间 (RFC3339) |
| `updated_at` | TEXT | 更新时间 (RFC3339) |

### players
| 列 | 类型 | 说明 |
|---|---|---|
| `id` | INTEGER PRIMARY KEY | HLTV 选手 ID |
| `name` | TEXT NOT NULL | 游戏 ID |
| `slug` | TEXT | URL slug |
| `real_name` | TEXT | 真实姓名 |
| `country` | TEXT | 国家代码 |
| `age` | INTEGER | 年龄 |
| `team` | TEXT | 所属队伍名 |
| `rating_json` | TEXT | PlayerRating JSON |
| `career_json` | TEXT | PlayerCareer JSON |
| `abilities_json` | TEXT | []PlayerAbility JSON |
| `overview_json` | TEXT | PlayerSummary JSON |
| `honors_json` | TEXT | []PlayerHonor JSON |
| `recent_matches_json` | TEXT | []PlayerRecentMatch JSON |
| `top20_json` | TEXT | map[string]int JSON |
| `fetched_at` | TEXT | 抓取时间 |
| `updated_at` | TEXT | 更新时间 |

### matches
| 列 | 类型 | 说明 |
|---|---|---|
| `match_id` | INTEGER PRIMARY KEY | HLTV 比赛 ID |
| `team1` | TEXT | 队伍 1 名称 |
| `team2` | TEXT | 队伍 2 名称 |
| `team1_id` | INTEGER | 队伍 1 HLTV ID |
| `team2_id` | INTEGER | 队伍 2 HLTV ID |
| `opponent` | TEXT | perspective 对手名 |
| `opponent_id` | INTEGER | 对手 HLTV ID |
| `event` | TEXT | 赛事名称 |
| `score` | TEXT | 比分 |
| `result` | TEXT | 结果 (win/loss/draw/scheduled/unknown) |
| `winner` | TEXT | 胜者 |
| `best_of` | TEXT | BO1/BO3/BO5 |
| `scheduled_at` | TEXT | 赛程时间 |
| `played_at` | TEXT | 比赛日期 |
| `map_text` | TEXT | 地图信息 |
| `source` | TEXT | upcoming/results — 区分赛程和赛果 |
| `fetched_at` | TEXT | 抓取时间 |
| `updated_at` | TEXT | 更新时间 |

### news
| 列 | 类型 | 说明 |
|---|---|---|
| `url_hash` | TEXT PRIMARY KEY | URL MD5 |
| `title` | TEXT NOT NULL | 标题 |
| `link` | TEXT | 原文链接 |
| `published_at` | TEXT | 发布时间 |
| `tag` | TEXT | 标签 |
| `body_text` | TEXT | 正文 |
| `author` | TEXT | 作者 |
| `fetched_at` | TEXT | 抓取时间 |

### realtime_news
| 列 | 类型 | 说明 |
|---|---|---|
| `url_hash` | TEXT PRIMARY KEY | URL MD5 |
| `section` | TEXT | 分区 |
| `category` | TEXT | 分类 |
| `title` | TEXT NOT NULL | 标题 |
| `link` | TEXT | 原文链接 |
| `relative_time` | TEXT | 相对时间 |
| `comments` | TEXT | 评论数 |
| `fetched_at` | TEXT | 抓取时间 |

### schema_version（元数据表）
| 列 | 类型 | 说明 |
|---|---|---|
| `version` | INTEGER PRIMARY KEY | 版本号 |
| `applied_at` | TEXT | 应用时间 |

## 过期清理策略

| 数据类型 | 保留期限 | 环境变量 |
|---|---|---|
| 队伍详情 / 选手详情 | 永久 | — |
| 赛程 / 赛果 | 90 天 | `HLTV_DB_RETENTION_MATCHES` |
| 新闻 | 30 天 | `HLTV_DB_RETENTION_NEWS` |
| 实时新闻 | 7 天 | `HLTV_DB_RETENTION_REALTIME_NEWS` |

启动时执行一次清理，之后每 24 小时循环。

## SSE 推送

- 端点: `GET /api/events`
- 后台刷新完成后 broadcast，前端 `EventSource` 监听
- 事件格式: `{entity: "player"|"team"|"matches"|"news", id: 123, name: "..."}`
- 30s keep-alive 心跳
- SSE hub 单例，全局 register/unregister/broadcast

## 错误处理

| 场景 | 行为 |
|---|---|
| SQLite 打开失败 | 日志 warn，降级为纯内存缓存模式，服务不中断 |
| db.Get 失败 | 跳过历史数据，走 HLTV 实时抓取 |
| db.Upsert 失败 | 日志 warn，不影响请求响应 |
| 数据库文件损坏 | 自动删旧建新 |

核心原则：**数据库不是关键路径**。数据库挂了，服务降级但不中断。

## 新增/改动文件

| 文件 | 改动 |
|---|---|
| `internal/storage/storage.go` | **新增** — Store 结构体、DB 生命周期 |
| `internal/storage/teams.go` | **新增** — TeamDetail Upsert/Get |
| `internal/storage/players.go` | **新增** — PlayerDetail Upsert/Get |
| `internal/storage/matches.go` | **新增** — NormalizedMatch Upsert/Get/Query |
| `internal/storage/news.go` | **新增** — NewsItem/NewsArticle/RealtimeNews Upsert/Get |
| `internal/storage/migration.go` | **新增** — 建表 + 迁移 + 清理 |
| `internal/http/sse.go` | **新增** — SSE hub + handler |
| `internal/http/router.go` | **改** — 注册 `/api/events` |
| `internal/facade/facade.go` | **改** — 每个 GetXxxCached 插入 db.Get → db.Upsert + SSE broadcast |
| `internal/config/config.go` | **改** — 新增 DBPath / retention 配置项 |
| `main.go` | **改** — storage init/shutdown，SSE hub 初始化 |
| `Dockerfile` | **微改** — 声明 data volume |
| `docker-compose.yml` | **改** — 挂载 data volume |
| `README.md` | **改** — 更新 Docker 启动命令 |
| `go.mod` | **改** — 新增 `modernc.org/sqlite` 依赖 |

## 技术选型

- **SQLite 驱动**: `modernc.org/sqlite` — 纯 Go，无 CGO，alpine Docker 零额外依赖
- **数据库接口**: `database/sql` 标准库
- **SSE**: 标准 `net/http` + `Flusher`，无第三方依赖
- **迁移**: 内建 `schema_version` 表 + Go 代码驱动，不引入 golang-migrate 等外部工具
