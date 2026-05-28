# 昵称字典后端迁移 + 编辑功能 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将选手/队伍昵称字典从前端硬编码迁移到后端（覆盖层模式），前端详情页支持内联编辑。

**Architecture:** 后端新增 `overrides.go`（`data/nicknames.json` 持久化 + 内存缓存 + `sync.RWMutex`）+ 扩展 `catalog.go`（PlayerCatalog/TeamNickname/PlayerNickname/BuildFullDict）+ 3 个 API。前端删除 `nicknames.ts`，新增 `useNicknames` hook 从 API 拉取字典，TeamDetail/PlayerDetail 加内联编辑按钮。

**Tech Stack:** Go 1.26, chi, React 18, TypeScript

---

### Task 1: Backend — 覆盖层存储 (`internal/localization/overrides.go`)

**Files:**
- Create: `internal/localization/overrides.go`
- Create: `internal/localization/overrides_test.go`

- [ ] **Step 1: 编写测试**

```go
package localization

import (
    "os"
    "path/filepath"
    "testing"
)

func TestInitOverrides_NoFile(t *testing.T) {
    // Ensure no file exists
    os.Remove("../../data/nicknames.json")
    if err := InitOverrides(); err != nil {
        t.Fatalf("InitOverrides: %v", err)
    }
    if n := GetTeamOverride("Vitality"); n != "" {
        t.Errorf("expected empty, got %q", n)
    }
    if n := GetPlayerOverride("donk"); n != "" {
        t.Errorf("expected empty, got %q", n)
    }
}

func TestSetAndGetTeamOverride(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    
    if err := SetTeamOverride("Vitality", "蜜蜂"); err != nil {
        t.Fatalf("SetTeamOverride: %v", err)
    }
    if n := GetTeamOverride("Vitality"); n != "蜜蜂" {
        t.Errorf("expected 蜜蜂, got %q", n)
    }
}

func TestSetAndGetPlayerOverride(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    
    if err := SetPlayerOverride("donk", "小驴"); err != nil {
        t.Fatalf("SetPlayerOverride: %v", err)
    }
    if n := GetPlayerOverride("donk"); n != "小驴" {
        t.Errorf("expected 小驴, got %q", n)
    }
}

func TestDeleteOverride_EmptyNickname(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    SetTeamOverride("Vitality", "蜜蜂")
    
    // Delete by setting empty
    if err := SetTeamOverride("Vitality", ""); err != nil {
        t.Fatalf("delete: %v", err)
    }
    if n := GetTeamOverride("Vitality"); n != "" {
        t.Errorf("expected empty after delete, got %q", n)
    }
}

func TestOverridePersistence(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    SetTeamOverride("Vitality", "蜜蜂")
    
    // Re-init should load from file
    if err := InitOverrides(); err != nil {
        t.Fatalf("re-init: %v", err)
    }
    if n := GetTeamOverride("Vitality"); n != "蜜蜂" {
        t.Errorf("persistence failed, got %q", n)
    }
    os.Remove("../../data/nicknames.json")
}

func TestConcurrentReadWrite(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    SetTeamOverride("Vitality", "test")
    
    done := make(chan bool)
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 100; j++ {
                GetTeamOverride("Vitality")
                GetPlayerOverride("donk")
            }
            done <- true
        }()
    }
    go func() {
        for j := 0; j < 50; j++ {
            SetTeamOverride("Vitality", "test")
        }
        done <- true
    }()
    for i := 0; i < 11; i++ {
        <-done
    }
    os.Remove("../../data/nicknames.json")
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/localization/ -run "TestInitOverrides_NoFile|TestSetAndGet|TestDelete|TestOverridePersistence|TestConcurrentReadWrite" -v -count=1
```
Expected: 全部 FAIL（overrides.go 不存在）

- [ ] **Step 3: 实现 `overrides.go`**

```go
package localization

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
)

type overridesStore struct {
    mu      sync.RWMutex
    teams   map[string]string
    players map[string]string
}

var ov = &overridesStore{}

const overridesFile = "data/nicknames.json"

type overridesData struct {
    Teams   map[string]string `json:"teams"`
    Players map[string]string `json:"players"`
}

func InitOverrides() error {
    data, err := os.ReadFile(overridesFile)
    if err != nil {
        ov.mu.Lock()
        ov.teams = make(map[string]string)
        ov.players = make(map[string]string)
        ov.mu.Unlock()
        return nil
    }
    var d overridesData
    if err := json.Unmarshal(data, &d); err != nil {
        return err
    }
    ov.mu.Lock()
    defer ov.mu.Unlock()
    ov.teams = d.Teams
    ov.players = d.Players
    if ov.teams == nil {
        ov.teams = make(map[string]string)
    }
    if ov.players == nil {
        ov.players = make(map[string]string)
    }
    return nil
}

func saveOverrides() error {
    ov.mu.RLock()
    d := overridesData{Teams: ov.teams, Players: ov.players}
    ov.mu.RUnlock()
    data, err := json.MarshalIndent(d, "", "  ")
    if err != nil {
        return err
    }
    if err := os.MkdirAll(filepath.Dir(overridesFile), 0700); err != nil {
        return err
    }
    return os.WriteFile(overridesFile, data, 0600)
}

func GetTeamOverride(name string) string {
    ov.mu.RLock()
    defer ov.mu.RUnlock()
    return ov.teams[name]
}

func GetPlayerOverride(name string) string {
    ov.mu.RLock()
    defer ov.mu.RUnlock()
    return ov.players[name]
}

func SetTeamOverride(name, nickname string) error {
    ov.mu.Lock()
    if nickname == "" {
        delete(ov.teams, name)
    } else {
        ov.teams[name] = nickname
    }
    ov.mu.Unlock()
    return saveOverrides()
}

func SetPlayerOverride(name, nickname string) error {
    ov.mu.Lock()
    if nickname == "" {
        delete(ov.players, name)
    } else {
        ov.players[name] = nickname
    }
    ov.mu.Unlock()
    return saveOverrides()
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/localization/ -run "TestInitOverrides_NoFile|TestSetAndGet|TestDelete|TestOverridePersistence|TestConcurrentReadWrite" -v -count=1
```
Expected: 全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/localization/overrides.go internal/localization/overrides_test.go
git commit -m "feat: add nickname overrides storage layer with JSON persistence"
```

---

### Task 2: Backend — PlayerCatalog + 查找/合并函数 (`internal/localization/catalog.go`)

**Files:**
- Modify: `internal/localization/catalog.go`
- Modify: `internal/localization/catalog_test.go`

- [ ] **Step 1: 在 `catalog.go` 中添加 `PlayerEntry`、`PlayerCatalog`、`PlayerNickname`、`TeamNickname`、`BuildFullDict`**

在 `catalog.go` 的 `package localization` 声明之后，`import` 块之前，`TeamEntry` 定义之后添加：

```go
// PlayerEntry represents a player nickname mapping.
type PlayerEntry struct {
    Canonical  string
    Colloquial string
}

// PlayerCatalog holds all built-in player nicknames.
var PlayerCatalog = []PlayerEntry{
    {Canonical: "ZywOo", Colloquial: "载物"},
    {Canonical: "s1mple", Colloquial: "森破"},
    {Canonical: "m0NESY", Colloquial: "小孩"},
    {Canonical: "donk", Colloquial: "小驴"},
    {Canonical: "NiKo", Colloquial: "🦐"},
    {Canonical: "dev1ce", Colloquial: "老爸"},
    {Canonical: "ropz", Colloquial: "车主"},
    {Canonical: "karrigan", Colloquial: "大表哥"},
    {Canonical: "apEX", Colloquial: "豆豆"},
    {Canonical: "flameZ", Colloquial: "火仔/🗿"},
    {Canonical: "Spinx", Colloquial: "米人"},
    {Canonical: "mezii", Colloquial: "妹子"},
    {Canonical: "jL", Colloquial: "jL"},
    {Canonical: "Aleksib", Colloquial: "小李子"},
    {Canonical: "b1t", Colloquial: "b1t"},
    {Canonical: "iM", Colloquial: "iM"},
    {Canonical: "w0nderful", Colloquial: "wdf"},
    {Canonical: "broky", Colloquial: "箱子"},
    {Canonical: "frozen", Colloquial: "寒王"},
    {Canonical: "Twistzz", Colloquial: "总监"},
    {Canonical: "huNter-", Colloquial: "表哥"},
    {Canonical: "jks", Colloquial: "jks"},
    {Canonical: "NAF", Colloquial: "NAF"},
    {Canonical: "YEKINDAR", Colloquial: "狂哥"},
    {Canonical: "cadiaN", Colloquial: "点子哥"},
    {Canonical: "stavn", Colloquial: "蛇"},
    {Canonical: "jabbi", Colloquial: "jabbi"},
    {Canonical: "TeSeS", Colloquial: "龙哥"},
    {Canonical: "EliGE", Colloquial: "鸡哥"},
    {Canonical: "Magisk", Colloquial: "魔男"},
    {Canonical: "dupreeh", Colloquial: "阿杜"},
    {Canonical: "Xyp9x", Colloquial: "九爷"},
    {Canonical: "gla1ve", Colloquial: "队长"},
    {Canonical: "electroNic", Colloquial: "电子哥"},
    {Canonical: "Perfecto", Colloquial: "P皇"},
    {Canonical: "Boombl4", Colloquial: "胖球"},
    {Canonical: "sh1ro", Colloquial: "息若"},
    {Canonical: "Ax1Le", Colloquial: "Ax1Le"},
    {Canonical: "Hobbit", Colloquial: "霍比特"},
    {Canonical: "KSCERATO", Colloquial: "KSCERATO/卡神"},
    {Canonical: "yuurih", Colloquial: "尤里"},
    {Canonical: "arT", Colloquial: "艺术哥"},
    {Canonical: "FalleN", Colloquial: "教父"},
    {Canonical: "rain", Colloquial: "雨神"},
    {Canonical: "siuhy", Colloquial: "siuhy"},
    {Canonical: "xertioN", Colloquial: "xertioN"},
    {Canonical: "torzsi", Colloquial: "托子"},
    {Canonical: "Brollan", Colloquial: "宝蓝"},
    {Canonical: "magixx", Colloquial: "马西西"},
    {Canonical: "chopper", Colloquial: "大超"},
    {Canonical: "zont1x", Colloquial: "宗主"},
    {Canonical: "degster", Colloquial: "degster"},
    {Canonical: "Kyojin", Colloquial: "巨人"},
    {Canonical: "malbsMd", Colloquial: "malbsMd"},
    {Canonical: "HeavyGod", Colloquial: "重God/重神"},
    {Canonical: "ultimate", Colloquial: "ultimate"},
    {Canonical: "jottAA", Colloquial: "jottAA"},
    {Canonical: "NertZ", Colloquial: "NertZ"},
    {Canonical: "misutaaa", Colloquial: "米苏塔"},
    {Canonical: "blameF", Colloquial: "胖虎"},
    {Canonical: "FL1T", Colloquial: "FL1T"},
    {Canonical: "fame", Colloquial: "fame"},
    {Canonical: "n0rb3r7", Colloquial: "n0rb3r7"},
    {Canonical: "XANTARES", Colloquial: "狠人"},
    {Canonical: "woxic", Colloquial: "woxic"},
    {Canonical: "Staehr", Colloquial: "野榜"},
    {Canonical: "br0", Colloquial: "br0"},
    {Canonical: "chelo", Colloquial: "chelo"},
    {Canonical: "skullz", Colloquial: "skullz"},
    {Canonical: "biguzera", Colloquial: "biguzera"},
    {Canonical: "bLitz", Colloquial: "布里茨"},
    {Canonical: "910", Colloquial: "910"},
    {Canonical: "Techno4K", Colloquial: "Techno4K"},
    {Canonical: "Senzu", Colloquial: "森组"},
    {Canonical: "nqz", Colloquial: "nqz"},
    {Canonical: "sl3nd", Colloquial: "sl3nd"},
    {Canonical: "volt", Colloquial: "volt"},
    {Canonical: "SunPayus", Colloquial: "阳叔"},
    {Canonical: "Nilo", Colloquial: "Nilo"},
    {Canonical: "maden", Colloquial: "maden"},
    {Canonical: "Patsi", Colloquial: "Patsi"},
    {Canonical: "BELCHONOKK", Colloquial: "BELCHONOKK"},
    {Canonical: "hallzerk", Colloquial: "hallzerk"},
    {Canonical: "Grim", Colloquial: "Grim"},
    {Canonical: "r1nkle", Colloquial: "r1nkle"},
    {Canonical: "bodyy", Colloquial: "bodyy"},
    {Canonical: "MATYS", Colloquial: "MATYS"},
    {Canonical: "Maka", Colloquial: "Maka"},
    {Canonical: "Lucky", Colloquial: "好运or坏运"},
    {Canonical: "JamYoung", Colloquial: "小鞠"},
    {Canonical: "Mercury", Colloquial: "汉堡"},
    {Canonical: "Starry", Colloquial: "小舅子"},
    {Canonical: "somebody", Colloquial: "sbd"},
    {Canonical: "Lack1", Colloquial: "Lack1"},
    {Canonical: "tN1R", Colloquial: "少爷/特尼尔"},
}

var playerCatalogMap map[string]string

func init() {
    playerCatalogMap = make(map[string]string, len(PlayerCatalog))
    for _, e := range PlayerCatalog {
        playerCatalogMap[e.Canonical] = e.Colloquial
    }
}
```

在文件末尾（`FormatTeamDisplay` 之后）添加：

```go
// PlayerNickname returns the colloquial name for a player.
// Checks overrides first, then falls back to the built-in catalog.
func PlayerNickname(name string) string {
    if n := GetPlayerOverride(name); n != "" {
        return n
    }
    return playerCatalogMap[name]
}

// TeamNickname returns the colloquial name for a team.
// Checks overrides first, then falls back to TeamCatalog.Colloquial.
func TeamNickname(name string) string {
    if n := GetTeamOverride(name); n != "" {
        return n
    }
    e := LookupTeam(name)
    if e == nil || e.Colloquial == "" {
        return ""
    }
    return e.Colloquial
}

// BuildFullDict returns the complete team+player nickname dictionaries
// with all variants expanded and user overrides applied on top.
func BuildFullDict() (teams, players map[string]string) {
    teams = make(map[string]string)
    for _, e := range TeamCatalog {
        if e.Colloquial == "" {
            continue
        }
        for _, v := range allVariants(&e) {
            if v != "" {
                teams[v] = e.Colloquial
            }
        }
    }
    // Apply team overrides — resolve canonical to all variants
    ov.mu.RLock()
    for canonicalName, nickname := range ov.teams {
        if nickname == "" {
            continue
        }
        if e := LookupTeam(canonicalName); e != nil {
            for _, v := range allVariants(e) {
                if v != "" {
                    teams[v] = nickname
                }
            }
        }
    }
    ov.mu.RUnlock()

    players = make(map[string]string, len(playerCatalogMap))
    for k, v := range playerCatalogMap {
        players[k] = v
    }
    ov.mu.RLock()
    for k, v := range ov.players {
        if v != "" {
            players[k] = v
        }
    }
    ov.mu.RUnlock()

    return
}
```

- [ ] **Step 2: 在 `catalog_test.go` 中添加测试**

修改文件顶部的 import 为：

```go
import (
    "os"
    "testing"
)
```

在文件末尾追加：

```go
func TestPlayerNickname_Builtin(t *testing.T) {
    if n := PlayerNickname("ZywOo"); n != "载物" {
        t.Errorf("expected 载物, got %q", n)
    }
    if n := PlayerNickname("donk"); n != "小驴" {
        t.Errorf("expected 小驴, got %q", n)
    }
}

func TestPlayerNickname_Unknown(t *testing.T) {
    if n := PlayerNickname("UnknownPlayerXYZ"); n != "" {
        t.Errorf("expected empty for unknown player, got %q", n)
    }
}

func TestTeamNickname(t *testing.T) {
    if n := TeamNickname("Spirit"); n != "绿龙" {
        t.Errorf("expected 绿龙 via alias, got %q", n)
    }
    if n := TeamNickname("Vitality"); n != "小蜜蜂" {
        t.Errorf("expected 小蜜蜂, got %q", n)
    }
}

func TestBuildFullDict_TeamsHasVariants(t *testing.T) {
    teams, _ := BuildFullDict()
    if teams["Spirit"] != "绿龙" {
        t.Errorf("expected Spirit→绿龙, got %q", teams["Spirit"])
    }
    if teams["Team Spirit"] != "绿龙" {
        t.Errorf("expected Team Spirit→绿龙, got %q", teams["Team Spirit"])
    }
    if teams["NAVI"] != "NaVi" {
        t.Errorf("expected NAVI→NaVi, got %q", teams["NAVI"])
    }
}

func TestBuildFullDict_Players(t *testing.T) {
    _, players := BuildFullDict()
    if players["ZywOo"] != "载物" {
        t.Errorf("expected ZywOo→载物, got %q", players["ZywOo"])
    }
}

func TestTeamNickname_WithOverride(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    SetTeamOverride("Vitality", "蜜蜂战队")
    defer os.Remove("../../data/nicknames.json")

    if n := TeamNickname("Vitality"); n != "蜜蜂战队" {
        t.Errorf("expected 蜜蜂战队 override, got %q", n)
    }
}

func TestPlayerNickname_WithOverride(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    SetPlayerOverride("donk", "小驴子")
    defer os.Remove("../../data/nicknames.json")

    if n := PlayerNickname("donk"); n != "小驴子" {
        t.Errorf("expected 小驴子 override, got %q", n)
    }
}

func TestBuildFullDict_WithOverride(t *testing.T) {
    os.Remove("../../data/nicknames.json")
    InitOverrides()
    SetTeamOverride("Natus Vincere", "天生赢家")
    defer os.Remove("../../data/nicknames.json")

    teams, _ := BuildFullDict()
    // variant "NAVI" should reflect the override
    if teams["NAVI"] != "天生赢家" {
        t.Errorf("expected NAVI→天生赢家 after override, got %q", teams["NAVI"])
    }
}
```

- [ ] **Step 3: 运行测试确认通过**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/localization/ -v -count=1
```
Expected: 全部 PASS（包括旧测试）

- [ ] **Step 4: Commit**

```bash
git add internal/localization/catalog.go internal/localization/catalog_test.go
git commit -m "feat: add PlayerCatalog, TeamNickname/PlayerNickname lookup, BuildFullDict with variant expansion"
```

---

### Task 3: Backend — API handlers + routes + main.go init

**Files:**
- Create: `internal/http/handlers/nicknames.go`
- Modify: `internal/http/router.go`
- Modify: `main.go`

- [ ] **Step 1: 创建 `internal/http/handlers/nicknames.go`**

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/arcdent/hltv-mcp/internal/localization"
)

type nicknameReq struct {
    Name     string `json:"name"`
    Nickname string `json:"nickname"`
}

// GetNicknames returns the full nickname dictionaries.
func (h *Handlers) GetNicknames(w http.ResponseWriter, r *http.Request) {
    teams, players := localization.BuildFullDict()
    writeJSON(w, map[string]any{
        "teams":   teams,
        "players": players,
    })
}

// PutTeamNickname saves a team nickname override.
func (h *Handlers) PutTeamNickname(w http.ResponseWriter, r *http.Request) {
    var req nicknameReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    if req.Name == "" {
        writeError(w, http.StatusBadRequest, "name is required")
        return
    }

    // Resolve to canonical via alias lookup
    e := localization.LookupTeam(req.Name)
    if e == nil {
        writeError(w, http.StatusBadRequest, "team not found in catalog")
        return
    }

    if err := localization.SetTeamOverride(e.Canonical, req.Nickname); err != nil {
        writeError(w, http.StatusInternalServerError, "failed to save")
        return
    }
    writeJSON(w, map[string]string{"status": "saved"})
}

// PutPlayerNickname saves a player nickname override.
func (h *Handlers) PutPlayerNickname(w http.ResponseWriter, r *http.Request) {
    var req nicknameReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    if req.Name == "" {
        writeError(w, http.StatusBadRequest, "name is required")
        return
    }

    // Open mode: any player name is accepted
    if err := localization.SetPlayerOverride(req.Name, req.Nickname); err != nil {
        writeError(w, http.StatusInternalServerError, "failed to save")
        return
    }
    writeJSON(w, map[string]string{"status": "saved"})
}
```

- [ ] **Step 2: 在 `router.go` 中添加路由**

在 `r.Post("/api/translate", h.PostTranslate)` 之后添加：

```go
        r.Get("/api/nicknames", h.GetNicknames)
        r.Put("/api/nicknames/team", h.PutTeamNickname)
        r.Put("/api/nicknames/player", h.PutPlayerNickname)
```

- [ ] **Step 3: 在 `main.go` 中添加 `InitOverrides` 调用**

在 `handlers.MigrateConfig()` 调用之后添加：

```go
        // Initialize nickname overrides
        if err := localization.InitOverrides(); err != nil {
            log.Printf("nickname overrides init note: %v", err)
        }
```

需要在 import 中添加 `"github.com/arcdent/hltv-mcp/internal/localization"`。检查 main.go 当前 import — 如果没有 localization，需要添加。

- [ ] **Step 4: 编译验证 + 运行现有测试**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build github.com/arcdent/hltv-mcp && go test ./internal/... -v -count=1 -timeout 30s
```
Expected: 编译成功，所有测试通过

- [ ] **Step 5: 启动服务并 curl 验证 API**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build -o hltv-mcp github.com/arcdent/hltv-mcp && ./hltv-mcp &
sleep 2

# Test GET
curl -s http://localhost:8082/api/nicknames | python3 -c "import sys,json; d=json.load(sys.stdin); print('teams:', len(d['teams']), 'players:', len(d['players']))"

# Test PUT team
curl -s -X PUT http://localhost:8082/api/nicknames/team -H 'Content-Type: application/json' -d '{"name":"NAVI","nickname":"天生赢家"}'

# Test GET after PUT (NAVI should now show 天生赢家)
curl -s http://localhost:8082/api/nicknames | python3 -c "import sys,json; d=json.load(sys.stdin); print('NAVI:', d['teams'].get('NAVI'))"

# Test delete (empty nickname)
curl -s -X PUT http://localhost:8082/api/nicknames/team -H 'Content-Type: application/json' -d '{"name":"Natus Vincere","nickname":""}'

# Cleanup
kill %1
```
Expected: teams count ~60+, players count 58, PUT returns `{"status":"saved"}`, NAVI shows "天生赢家" after PUT, delete reverts to "NaVi"

- [ ] **Step 6: Commit**

```bash
git add internal/http/handlers/nicknames.go internal/http/router.go main.go
git commit -m "feat: add nickname API endpoints and InitOverrides startup"
```

---

### Task 4: 前端 — hook + 组件更新 + 删除 nicknames.ts

**Files:**
- Create: `frontend/src/hooks/useNicknames.ts`
- Delete: `frontend/src/data/nicknames.ts`
- Modify: `frontend/src/components/TeamDetail.tsx`
- Modify: `frontend/src/components/PlayerDetail.tsx`
- Modify: `frontend/src/components/SearchableList.tsx`
- Modify: `frontend/src/pages/Matches.tsx`

- [ ] **Step 1: 创建 `frontend/src/hooks/useNicknames.ts`**

```typescript
import { useState, useEffect, useCallback } from 'react'

type NicknameDict = Record<string, string>

let cachedTeams: NicknameDict | null = null
let cachedPlayers: NicknameDict | null = null
let fetchPromise: Promise<void> | null = null

async function ensureLoaded(): Promise<void> {
  if (cachedTeams && cachedPlayers) return
  if (fetchPromise) {
    await fetchPromise
    return
  }
  fetchPromise = (async () => {
    const resp = await fetch('/api/nicknames')
    const data = await resp.json()
    cachedTeams = data.teams ?? {}
    cachedPlayers = data.players ?? {}
  })()
  await fetchPromise
  fetchPromise = null
}

async function saveNickname(type: 'team' | 'player', name: string, nickname: string): Promise<void> {
  const resp = await fetch(`/api/nicknames/${type}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, nickname }),
  })
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}))
    throw new Error((err as any).error ?? 'save failed')
  }
  // Update local cache
  if (type === 'team') {
    cachedTeams = { ...cachedTeams, [name]: nickname }
  } else {
    cachedPlayers = { ...cachedPlayers, [name]: nickname }
  }
}

export default function useNicknames() {
  const [teamNicknames, setTeamNicknames] = useState<NicknameDict>(cachedTeams ?? {})
  const [playerNicknames, setPlayerNicknames] = useState<NicknameDict>(cachedPlayers ?? {})
  const [loading, setLoading] = useState(!cachedTeams)

  useEffect(() => {
    ensureLoaded().then(() => {
      setTeamNicknames(cachedTeams!)
      setPlayerNicknames(cachedPlayers!)
      setLoading(false)
    })
  }, [])

  const saveTeamNickname = useCallback(async (name: string, nickname: string) => {
    await saveNickname('team', name, nickname)
    setTeamNicknames({ ...cachedTeams! })
  }, [])

  const savePlayerNickname = useCallback(async (name: string, nickname: string) => {
    await saveNickname('player', name, nickname)
    setPlayerNicknames({ ...cachedPlayers! })
  }, [])

  return { teamNicknames, playerNicknames, saveTeamNickname, savePlayerNickname, loading }
}
```

- [ ] **Step 2: 删除 `frontend/src/data/nicknames.ts`**

```bash
rm /home/arcdent/github/hltv-mcp-fully-rebuild/frontend/src/data/nicknames.ts
```

- [ ] **Step 3: 修改 `TeamDetail.tsx`**

修改第 4 行的 import：删除 `{ teamNicknames, playerNicknames } from '../data/nicknames'`，添加 `import useNicknames from '../hooks/useNicknames'`。

在组件内第 17 行后添加：
```typescript
const { teamNicknames, playerNicknames, saveTeamNickname, savePlayerNickname } = useNicknames()
```

把第 36 行的 `const cnName = teamNicknames[p?.name ?? '']` 改为：
```typescript
const [editingTeamNick, setEditingTeamNick] = useState(false)
const [editingPlayerId, setEditingPlayerId] = useState<number | null>(null)
const cnName = teamNicknames[p?.name ?? '']
```

团队金标（第 56 行）改为可编辑：
```tsx
{editingTeamNick ? (
  <input
    autoFocus
    defaultValue={cnName}
    style={{padding:'2px 8px',borderRadius:4,fontSize:11,background:'var(--input-bg)',border:'1px solid var(--gold)',color:'var(--text)',width:80,outline:'none'}}
    onKeyDown={e => {
      if (e.key === 'Enter') { saveTeamNickname(p?.name ?? '', (e.target as HTMLInputElement).value); setEditingTeamNick(false) }
      if (e.key === 'Escape') setEditingTeamNick(false)
    }}
    onBlur={e => { saveTeamNickname(p?.name ?? '', e.target.value); setEditingTeamNick(false) }}
  />
) : (
  <span style={{padding:'2px 10px',borderRadius:4,fontSize:11,background:'var(--gold-dim)',color:'var(--gold)',fontWeight:600,display:'inline-flex',alignItems:'center',gap:4}}>
    {cnName || '无简称'}
    <span onClick={() => setEditingTeamNick(true)} style={{cursor:'pointer',opacity:0.6,fontSize:10}} title="编辑简称">✏️</span>
  </span>
)}
```

roster 列表中（第 155 行）选手昵称改为可编辑：
```tsx
{(playerNicknames[pl.name] || editingPlayerId === pl.id) ? (
  editingPlayerId === pl.id ? (
    <input
      autoFocus
      defaultValue={playerNicknames[pl.name] ?? ''}
      style={{fontSize:11,background:'var(--input-bg)',border:'1px solid var(--gold)',borderRadius:3,padding:'1px 4px',color:'var(--text)',width:60,outline:'none',marginLeft:4}}
      onKeyDown={e => {
        if (e.key === 'Enter') { savePlayerNickname(pl.name, (e.target as HTMLInputElement).value); setEditingPlayerId(null) }
        if (e.key === 'Escape') setEditingPlayerId(null)
      }}
      onBlur={e => { savePlayerNickname(pl.name, e.target.value); setEditingPlayerId(null) }}
      onClick={e => e.stopPropagation()}
    />
  ) : (
    <span style={{fontSize:11,color:'var(--text-muted)',marginLeft:4,fontWeight:400}}>
      {playerNicknames[pl.name]}
      <span onClick={e => { e.stopPropagation(); setEditingPlayerId(pl.id) }} style={{cursor:'pointer',opacity:0.4,fontSize:9,marginLeft:2}} title="编辑简称">✏️</span>
    </span>
  )
) : null}
```

- [ ] **Step 4: 修改 `PlayerDetail.tsx`**

添加 import：
```typescript
import useNicknames from '../hooks/useNicknames'
```

组件内初始化后添加：
```typescript
const { playerNicknames, savePlayerNickname } = useNicknames()
const [editingNick, setEditingNick] = useState(false)
```

header 区域的名字旁（第 42-44 行之间）添加昵称显示，在 `{p.real_name ? ...}` 行之后、`<div style={{display:'flex',...}}>` 之前：

```tsx
<div style={{display:'flex',alignItems:'center',gap:4,marginTop:2}}>
  {editingNick ? (
    <input
      autoFocus
      defaultValue={playerNicknames[p.name] ?? ''}
      style={{fontSize:12,background:'var(--input-bg)',border:'1px solid var(--gold)',borderRadius:3,padding:'2px 6px',color:'var(--text)',width:100,outline:'none'}}
      onKeyDown={e => {
        if (e.key === 'Enter') { savePlayerNickname(p.name, (e.target as HTMLInputElement).value); setEditingNick(false) }
        if (e.key === 'Escape') setEditingNick(false)
      }}
      onBlur={e => { savePlayerNickname(p.name, e.target.value); setEditingNick(false) }}
    />
  ) : (
    <>
      {playerNicknames[p.name] && (
        <span style={{padding:'2px 8px',borderRadius:4,fontSize:11,background:'var(--gold-dim)',color:'var(--gold)',fontWeight:600,display:'inline-flex',alignItems:'center',gap:3}}>
          {playerNicknames[p.name]}
          <span onClick={() => setEditingNick(true)} style={{cursor:'pointer',opacity:0.6,fontSize:9}} title="编辑简称">✏️</span>
        </span>
      )}
      {!playerNicknames[p.name] && (
        <span onClick={() => setEditingNick(true)} style={{cursor:'pointer',fontSize:11,color:'var(--text-muted)',opacity:0.5}} title="添加简称">+ 添加简称</span>
      )}
    </>
  )}
</div>
```

- [ ] **Step 5: 修改 `SearchableList.tsx`**

第 4 行 import 改为：
```typescript
import useNicknames from '../hooks/useNicknames'
```

组件内（第 38 行后）添加：
```typescript
const { teamNicknames, playerNicknames } = useNicknames()
```

第 88-91 行不需要改动（仍使用 `playerNicknames[item.name] || teamNicknames[item.name]`），只是数据来源变了。

- [ ] **Step 6: 修改 `Matches.tsx`**

第 3 行 import 改为：
```typescript
import useNicknames from '../hooks/useNicknames'
```

组件内（第 14 行后）添加：
```typescript
const { teamNicknames: nicknames } = useNicknames()
```

第 129-130 行不需要改动（仍使用 `nicknames[m.team1 ?? '']`），只是数据来源变了。

- [ ] **Step 7: 构建前端 + 编译后端**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npm run build && cd .. && go build github.com/arcdent/hltv-mcp
```
Expected: TypeScript 编译无错误，Go 编译成功

- [ ] **Step 8: 启动服务并验证前端**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && ./hltv-mcp &
sleep 2

# Verify GET /api/nicknames
curl -s http://localhost:8082/api/nicknames | python3 -c "import sys,json; d=json.load(sys.stdin); print('teams:', len(d['teams']), 'players:', len(d['players']))"

# Verify SPA loads
curl -s http://localhost:8082/ | head -5

kill %1
```
Expected: API 返回完整字典，SPA 正常加载

- [ ] **Step 9: Commit**

```bash
git add frontend/src/hooks/useNicknames.ts frontend/src/data/nicknames.ts frontend/src/components/TeamDetail.tsx frontend/src/components/PlayerDetail.tsx frontend/src/components/SearchableList.tsx frontend/src/pages/Matches.tsx
git commit -m "feat: move nicknames to backend, add inline edit on team/player detail pages"
```

---

### 最终验证

- [ ] **全量测试**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/... -v -count=1 -timeout 30s
```
Expected: 全部测试通过

- [ ] **手工冒烟测试**

启动服务后浏览器访问 `http://localhost:8082`：
1. Teams 页面搜索 Vitality → 点击打开详情 → 金标显示"小蜜蜂"，旁边有 ✏️ 按钮
2. 点击 ✏️ → 输入框出现 → 修改为"蜜蜂队" → Enter → 金标更新
3. 刷新页面 → 确认修改持久化
4. 队员列表 → 选手昵称旁有 ✏️ → 点击编辑 → Enter 保存
5. Players 页面搜索 ZywOo → 点击打开详情 → header 显示"载物"+ ✏️
6. Matches 页面 → 点击赛事查看比赛 → 队伍简称正常显示
