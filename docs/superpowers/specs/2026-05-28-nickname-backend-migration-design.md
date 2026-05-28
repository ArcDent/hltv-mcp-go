# 昵称字典后端迁移 + 编辑功能 设计文档

## 目标

将选手/队伍昵称字典从前端硬编码迁移到后端，提供 API 查询和覆盖编辑能力，前端选手和队伍详情页支持内联编辑。

## 架构

覆盖层模式：后端保留硬编码默认值（`TeamCatalog.Colloquial` + 新增 `PlayerCatalog`），用户编辑写入 `data/nicknames.json` 持久化覆盖。读取时 override 优先，未覆盖回退默认值。前端删除 `nicknames.ts`，启动时从 `GET /api/nicknames` 一次性拉取完整字典。

## 后端

### 数据模型

**新文件 `internal/localization/overrides.go`** — 覆盖层管理

```
data/nicknames.json
{
  "teams":  { "Vitality": "蜜蜂" },
  "players": { "donk": "小驴", "ZywOo": "载物" }
}
```

- 只存用户编辑过的条目，不重复默认值
- 启动时加载到内存 `map[string]string`，读写用 `sync.RWMutex`
- 写操作同时更新内存 + 序列化刷盘

**修改 `internal/localization/catalog.go`**

- 新增 `PlayerCatalog`：从 `nicknames.ts` 中 58 条选手简称迁移为 Go 硬编码 `[]PlayerEntry` 切片
- 新增 `PlayerEntry` 类型：`{ Canonical, Colloquial string }`
- 新增 `TeamNickname(name string) string`：先查 overrides，miss 回退 `TeamCatalog.Colloquial`
- 新增 `PlayerNickname(name string) string`：先查 overrides，miss 回退 `PlayerCatalog`
- 新增 `BuildFullDict() (teams, players map[string]string)`：合并默认+覆盖，返回完整字典供 API 使用

### API

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/nicknames` | 返回 `{ teams: {...}, players: {...} }` 完整字典 |
| `PUT` | `/api/nicknames/team` | `{ name, nickname }` 保存覆盖，空 nickname 删除覆盖 |
| `PUT` | `/api/nicknames/player` | `{ name, nickname }` 保存覆盖，空 nickname 删除覆盖 |

错误响应：
- 400：`name` 不在默认 catalog 中（"team/player not found in catalog"）
- 400：JSON 格式无效

### 启动流程

```
main.go → crypto.InitKey() → handlers.MigrateConfig() → localization.InitOverrides() → 就绪
```

## 前端

### 删除文件

- `frontend/src/data/nicknames.ts`

### 新增文件

- `frontend/src/hooks/useNicknames.ts` — 全局共享 hook
  - `GET /api/nicknames` 拉取字典，结果缓存在模块级变量
  - 暴露 `{ teamNicknames, playerNicknames, saveTeamNickname, savePlayerNickname, loading }`
  - `saveTeamNickname(name, nickname)` → `PUT /api/nicknames/team`，更新本地缓存
  - `savePlayerNickname(name, nickname)` → `PUT /api/nicknames/player`，更新本地缓存

### 修改文件

| 文件 | 改动 |
|------|------|
| `TeamDetail.tsx` | 团队 header 金标旁加 ✏️ 编辑按钮，点击内联编辑；roster 列表每个选手昵称旁加 ✏️ 编辑按钮 |
| `PlayerDetail.tsx` | header 名字旁显示昵称（当前无则仅 hover 显示编辑按钮），点击内联编辑 |
| `SearchableList.tsx` | 改用 `useNicknames()`，去掉 `nicknames.ts` 导入 |
| `Matches.tsx` | 同上 |

### 内联编辑交互

- 点击 ✏️ → 昵称文字替换为 `<input>`，自动聚焦
- Enter / 失焦 → 调 save API → 恢复文字显示
- Esc → 取消，恢复原文字
- 保存中 input 显示浅灰背景 + 小 spinner

## 边界情况

- **删除昵称**：传空 nickname 字符串，删除 override 条目，回退默认值
- **首次启动无文件**：静默跳过，内存 map 为空，全部回退默认值
- **并发写**：`sync.RWMutex` 保护，刷盘在持锁内完成
- **PUT 未识别的 name**：返回 400 "not found in catalog"
- **API 失败**：前端保留旧缓存值，console.error 记录
