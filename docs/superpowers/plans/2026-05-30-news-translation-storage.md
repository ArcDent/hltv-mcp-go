# News Translation Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add persistent storage for news translation results (title_zh, body_text_zh) with auto-translate titles on ingest and on-demand body translation, integrated into the existing three-tier cache architecture.

**Architecture:** New `internal/translator/` package holds `TranslateConfig` and `Translator`. Facade gets a config-factory function for hot-reload-safe background title translation. Handlers get a `*storage.Store` for body-translation writeback. Migration v2 adds three TEXT columns to existing tables.

**Tech Stack:** Go 1.26, SQLite (modernc.org/sqlite), existing chi router

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `internal/translator/translator.go` | **Create** | `TranslateConfig` struct, `Translator` struct, `New()`, `TranslateTitle()`, `TranslateBody()` |
| `internal/storage/migration.go` | Modify | Add `applyV2` with 3 ALTER TABLE + bump schema version |
| `internal/types/types.go` | Modify | Add `TitleZh`/`BodyTextZh` fields to news types |
| `internal/storage/news.go` | Modify | 3 update methods + 2 Has* check methods + update read SQL/Scan for new cols |
| `internal/http/handlers/translate.go` | Modify | Import translator pkg, refactor PostTranslate, fileConfig for persistence |
| `internal/http/handlers/handlers.go` | Modify | `New()` accepts `*storage.Store`, store it on Handlers |
| `internal/http/router.go` | Modify | `NewRouter()` accepts `*storage.Store`, passes to `handlers.New()` |
| `internal/facade/facade.go` | Modify | Add `translateCfgFn` field, translate methods, update `New()` |
| `internal/facade/news.go` | Modify | Wire `go f.translateNew*(allItems)` after BatchUpsert |
| `main.go` | Modify | Wire translateCfgFn + store into facade, router, and handlers |

---

### Task 1: Create `internal/translator/` package

**Files:**
- Create: `internal/translator/translator.go`

- [ ] **Step 1: Write the translator package**

```go
package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TranslateConfig holds LLM translation provider configuration.
type TranslateConfig struct {
	ProviderURL string
	APIKey      string
	Model       string
}

// Translator proxies translation requests to a configured LLM API.
type Translator struct {
	providerURL string
	apiKey      string
	model       string
	client      *http.Client
}

// New creates a Translator from the given config.
func New(cfg TranslateConfig) *Translator {
	return &Translator{
		providerURL: cfg.ProviderURL,
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// TranslateTitle translates a CS esports news title to Simplified Chinese.
func (t *Translator) TranslateTitle(ctx context.Context, text string) (string, error) {
	systemPrompt := "将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释"
	return t.translate(ctx, systemPrompt, text)
}

// TranslateBody translates CS esports news body text to Simplified Chinese.
func (t *Translator) TranslateBody(ctx context.Context, text string) (string, error) {
	systemPrompt := "将以下CS电竞新闻正文翻译为简体中文"
	return t.translate(ctx, systemPrompt, text)
}

func (t *Translator) translate(ctx context.Context, systemPrompt, text string) (string, error) {
	reqBody := map[string]any{
		"model": t.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": text},
		},
		"temperature": 0.1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(t.providerURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
```

- [ ] **Step 2: Verify the package compiles**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/translator/
```
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/translator/translator.go
git commit -m "feat: add translator package with TranslateConfig and Translator"
```

---

### Task 2: Add migration v2

**Files:**
- Modify: `internal/storage/migration.go`

- [ ] **Step 1: Add applyV2 and wire into migrate()**

In `migrate()`, after the `if v < 1 { ... }` block (line 26), add:

```go
if v < 2 {
    if err := applyV2(db); err != nil {
        return err
    }
}
```

After the `applyV1` function, append:

```go
func applyV2(db *sql.DB) error {
    stmts := []string{
        "ALTER TABLE news ADD COLUMN title_zh TEXT",
        "ALTER TABLE news ADD COLUMN body_text_zh TEXT",
        "ALTER TABLE realtime_news ADD COLUMN title_zh TEXT",
    }
    for _, stmt := range stmts {
        if _, err := db.Exec(stmt); err != nil {
            return err
        }
    }
    _, err := db.Exec("INSERT INTO schema_version(version) VALUES(2)")
    return err
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/storage/migration.go
git commit -m "feat: add migration v2 — news title_zh, body_text_zh; realtime_news title_zh"
```

---

### Task 3: Update types with translation fields

**Files:**
- Modify: `internal/types/types.go`

- [ ] **Step 1: Add TitleZh to NewsItem (line 72-77)**

```go
type NewsItem struct {
	Title       string `json:"title"`
	TitleZh     string `json:"title_zh,omitempty"`
	Link        string `json:"link,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	Tag         string `json:"tag,omitempty"`
}
```

- [ ] **Step 2: Add TitleZh and BodyTextZh to NewsArticle (line 350-356)**

```go
type NewsArticle struct {
	Title       string `json:"title"`
	TitleZh     string `json:"title_zh,omitempty"`
	PublishedAt string `json:"published_at"`
	Link        string `json:"link"`
	BodyText    string `json:"body_text"`
	BodyTextZh  string `json:"body_text_zh,omitempty"`
	Author      string `json:"author,omitempty"`
}
```

- [ ] **Step 3: Add TitleZh to RealtimeNewsItem (line 80-87)**

```go
type RealtimeNewsItem struct {
	Section      string `json:"section"`
	Category     string `json:"category,omitempty"`
	Title        string `json:"title"`
	TitleZh      string `json:"title_zh,omitempty"`
	RelativeTime string `json:"relative_time,omitempty"`
	Comments     string `json:"comments,omitempty"`
	Link         string `json:"link,omitempty"`
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/types/
```
Expected: no output (success)

- [ ] **Step 5: Commit**

```bash
git add internal/types/types.go
git commit -m "feat: add TitleZh/BodyTextZh fields to news types"
```

---

### Task 4: Update storage/news.go — add methods and update read queries

**Files:**
- Modify: `internal/storage/news.go`

- [ ] **Step 1: Add HasNewsTitleZh and HasRealtimeTitleZh check methods**

Append to the file:

```go
// HasNewsTitleZh returns true if the news item already has a translated title.
func (s *Store) HasNewsTitleZh(url string) (bool, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	var titleZh sql.NullString
	err := s.db.QueryRow("SELECT title_zh FROM news WHERE url_hash=?", hash).Scan(&titleZh)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return titleZh.String != "", nil
}

// HasRealtimeTitleZh returns true if the realtime news item already has a translated title.
func (s *Store) HasRealtimeTitleZh(url string) (bool, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	var titleZh sql.NullString
	err := s.db.QueryRow("SELECT title_zh FROM realtime_news WHERE url_hash=?", hash).Scan(&titleZh)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return titleZh.String != "", nil
}
```

- [ ] **Step 2: Add UpdateNewsTitleZh, UpdateNewsBodyZh, UpdateRealtimeTitleZh**

Append after the Has* methods:

```go
// UpdateNewsTitleZh stores a translated title for an archive news item.
func (s *Store) UpdateNewsTitleZh(url string, titleZh string) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	_, err := s.db.Exec("UPDATE news SET title_zh=? WHERE url_hash=?", titleZh, hash)
	return err
}

// UpdateNewsBodyZh stores a translated body for an archive news article.
func (s *Store) UpdateNewsBodyZh(url string, bodyZh string) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	_, err := s.db.Exec("UPDATE news SET body_text_zh=? WHERE url_hash=?", bodyZh, hash)
	return err
}

// UpdateRealtimeTitleZh stores a translated title for a realtime news item.
func (s *Store) UpdateRealtimeTitleZh(url string, titleZh string) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	_, err := s.db.Exec("UPDATE realtime_news SET title_zh=? WHERE url_hash=?", titleZh, hash)
	return err
}
```

- [ ] **Step 3: Update GetNewsArticle SQL, Scan variables, and field assignments**

Change the SQL query (line 66):
```go
row := s.db.QueryRow("SELECT title, link, published_at, body_text, author, title_zh, body_text_zh FROM news WHERE url_hash=?", hash)
```

Change the variable declarations (line 69):
```go
var article types.NewsArticle
var bodyText, author, titleZh, bodyTextZh sql.NullString
```

Change the Scan (line 70):
```go
err := row.Scan(&article.Title, &article.Link, &article.PublishedAt, &bodyText, &author, &titleZh, &bodyTextZh)
```

Add after `article.Author = author.String` (after line 78):
```go
article.TitleZh = titleZh.String
article.BodyTextZh = bodyTextZh.String
```

- [ ] **Step 4: Update QueryNews SQL and Scan**

Change the query (line 83):
```go
query := "SELECT title, link, published_at, tag, title_zh FROM news ORDER BY published_at DESC"
```

Change the Scan block (lines 95-98):
```go
var n types.NewsItem
var titleZh sql.NullString
if err := rows.Scan(&n.Title, &n.Link, &n.PublishedAt, &n.Tag, &titleZh); err != nil {
    return nil, err
}
n.TitleZh = titleZh.String
items = append(items, n)
```

- [ ] **Step 5: Update QueryRealtimeNews SQL and Scan**

Change the query (line 144):
```go
query := "SELECT section, category, title, link, relative_time, comments, title_zh FROM realtime_news"
```

Change the Scan block (lines 155-159):
```go
var n types.RealtimeNewsItem
var titleZh sql.NullString
if err := rows.Scan(&n.Section, &n.Category, &n.Title, &n.Link, &n.RelativeTime, &n.Comments, &titleZh); err != nil {
    return nil, err
}
n.TitleZh = titleZh.String
items = append(items, n)
```

- [ ] **Step 6: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no output (success)

- [ ] **Step 7: Commit**

```bash
git add internal/storage/news.go
git commit -m "feat: add translation read/write methods to news storage"
```

---

### Task 5: Update handlers — refactor PostTranslate, add store field

**Files:**
- Modify: `internal/http/handlers/translate.go`
- Modify: `internal/http/handlers/handlers.go`
- Modify: `internal/http/router.go`

- [ ] **Step 1: Rewrite translate.go**

Replace the file. Key structural changes:
- Delete `TranslateConfig` struct — replaced by `translator.TranslateConfig`
- Add `fileConfig` struct (local, for JSON persistence with `encrypted` field)
- `loadTranslateConfig` returns `(translator.TranslateConfig, error)` using `fileConfig` internally
- `saveTranslateConfig` accepts `(translator.TranslateConfig, error)` using `fileConfig` internally  
- `LoadTranslateConfig` returns `(translator.TranslateConfig, error)` — public, used by main.go for facade factory
- `PostTranslate`: creates `translator.New(cfg)`, calls `TranslateTitle`/`TranslateBody`, stores body result when `type=body` and `url` non-empty

```go
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/arcdent/hltv-mcp/internal/crypto"
	"github.com/arcdent/hltv-mcp/internal/storage"
	"github.com/arcdent/hltv-mcp/internal/translator"
)

const (
	translateConfigFile = "translate_config.json"
	dataDir             = "data"
)

// fileConfig is the on-disk format with encrypted flag for key detection.
type fileConfig struct {
	ProviderURL string `json:"provider_url"`
	APIKey      string `json:"api_key"`
	Model       string `json:"model"`
	Encrypted   bool   `json:"encrypted,omitempty"`
}

func configDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, dataDir)
}

func configPath() string {
	return filepath.Join(configDir(), translateConfigFile)
}

func oldConfigPath() string {
	exec, _ := os.Executable()
	dir := filepath.Dir(exec)
	return filepath.Join(dir, translateConfigFile)
}

// MigrateConfig moves an existing plaintext config from the old location
// (next to the executable) to data/translate_config.json with encryption.
func MigrateConfig() error {
	if _, err := os.Stat(configPath()); err == nil {
		return nil
	}
	oldPath := oldConfigPath()
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return nil
	}
	var fcfg fileConfig
	if err := json.Unmarshal(data, &fcfg); err != nil {
		return nil
	}
	if fcfg.APIKey == "" {
		return nil
	}
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	fcfg.Encrypted = true
	encryptedKey, err := crypto.Encrypt(fcfg.APIKey)
	if err != nil {
		return fmt.Errorf("encrypt key: %w", err)
	}
	fcfg.APIKey = encryptedKey
	data, err = json.MarshalIndent(fcfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath(), data, 0600); err != nil {
		return err
	}
	os.Remove(oldPath)
	return nil
}

func loadTranslateConfig() (translator.TranslateConfig, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return translator.TranslateConfig{}, err
	}
	var fcfg fileConfig
	if err := json.Unmarshal(data, &fcfg); err != nil {
		return translator.TranslateConfig{}, err
	}
	if fcfg.Encrypted {
		key, err := crypto.Decrypt(fcfg.APIKey)
		if err != nil {
			return translator.TranslateConfig{}, fmt.Errorf("decrypt api key: %w", err)
		}
		fcfg.APIKey = key
	} else if fcfg.APIKey != "" {
		// Auto-upgrade plaintext config to encrypted
		fcfg.Encrypted = true
		encryptedKey, err := crypto.Encrypt(fcfg.APIKey)
		if err == nil {
			fcfg.APIKey = encryptedKey
			if data, err := json.MarshalIndent(fcfg, "", "  "); err == nil {
				os.WriteFile(configPath(), data, 0600)
			}
		}
	}
	return translator.TranslateConfig{
		ProviderURL: fcfg.ProviderURL,
		APIKey:      fcfg.APIKey,
		Model:       fcfg.Model,
	}, nil
}

func saveTranslateConfig(cfg translator.TranslateConfig) error {
	fcfg := fileConfig{
		ProviderURL: cfg.ProviderURL,
		Model:       cfg.Model,
		Encrypted:   true,
	}
	if cfg.APIKey != "" && !strings.Contains(cfg.APIKey, "***") {
		encryptedKey, err := crypto.Encrypt(cfg.APIKey)
		if err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}
		fcfg.APIKey = encryptedKey
	} else {
		fcfg.APIKey = cfg.APIKey
	}
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(fcfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

func maskKey(key string) string {
	if len(key) <= 6 {
		return strings.Repeat("*", len(key))
	}
	return key[:3] + strings.Repeat("*", len(key)-6) + key[len(key)-3:]
}

// LoadTranslateConfig exposes config loading for use by other packages.
func LoadTranslateConfig() (translator.TranslateConfig, error) {
	return loadTranslateConfig()
}

// GetTranslateConfig returns the current translation config with masked API key.
func (h *Handlers) GetTranslateConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadTranslateConfig()
	if err != nil {
		writeJSON(w, map[string]any{
			"provider_url": "",
			"api_key":      "",
			"model":        "",
			"configured":   false,
		})
		return
	}
	writeJSON(w, map[string]any{
		"provider_url": cfg.ProviderURL,
		"api_key":      maskKey(cfg.APIKey),
		"model":        cfg.Model,
		"configured":   cfg.ProviderURL != "" && cfg.APIKey != "",
	})
}

// PutTranslateConfig saves the translation config.
func (h *Handlers) PutTranslateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg translator.TranslateConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.Contains(cfg.APIKey, "***") {
		existing, err := loadTranslateConfig()
		if err != nil {
			writeError(w, http.StatusBadRequest, "无法加载现有配置，请重新输入完整的 API Key")
			return
		}
		cfg.APIKey = existing.APIKey
	}
	if err := saveTranslateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save config")
		return
	}
	writeJSON(w, map[string]string{"status": "saved"})
}

// PostTranslate proxies translation requests to the configured LLM API.
func (h *Handlers) PostTranslate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
		Type string `json:"type"`
		URL  string `json:"url,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	cfg, err := loadTranslateConfig()
	if err != nil || cfg.ProviderURL == "" || cfg.APIKey == "" {
		writeError(w, http.StatusBadRequest, "翻译服务未配置")
		return
	}

	t := translator.New(cfg)

	var translated string
	if req.Type == "title" {
		translated, err = t.TranslateTitle(r.Context(), req.Text)
	} else {
		translated, err = t.TranslateBody(r.Context(), req.Text)
	}
	if err != nil {
		log.Printf("translate: %v", err)
		writeError(w, http.StatusBadGateway, "翻译失败: "+err.Error())
		return
	}

	// Store body translation when URL is provided
	if req.Type == "body" && req.URL != "" && h.store != nil {
		if err := h.store.UpdateNewsBodyZh(req.URL, translated); err != nil {
			log.Printf("translate: store body_zh: %v", err)
		}
	}

	writeJSON(w, map[string]string{"translated": translated})
}
```

- [ ] **Step 2: Add store field to Handlers struct**

In `handlers.go`, add import:
```go
"github.com/arcdent/hltv-mcp/internal/storage"
```

Change the struct and constructor:
```go
type Handlers struct {
	f     *facade.HltvFacade
	store *storage.Store
}

func New(f *facade.HltvFacade, store *storage.Store) *Handlers {
	return &Handlers{f: f, store: store}
}
```

- [ ] **Step 3: Update router.go — pass store through**

Change `NewRouter` signature:
```go
func NewRouter(f *facade.HltvFacade, frontendFS fs.FS, sseHub *SSEHub, store *storage.Store) http.Handler {
```

Add import:
```go
"github.com/arcdent/hltv-mcp/internal/storage"
```

Change the handler creation:
```go
h := handlers.New(f, store)
```

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/http/...
```
Expected: no output (success)

- [ ] **Step 5: Commit**

```bash
git add internal/http/handlers/translate.go internal/http/handlers/handlers.go internal/http/router.go
git commit -m "refactor: use translator package in handlers, add store for body writeback"
```

---

### Task 6: Update facade — add translateCfgFn and translation methods

**Files:**
- Modify: `internal/facade/facade.go`

- [ ] **Step 1: Add import and field to HltvFacade**

Add import:
```go
"github.com/arcdent/hltv-mcp/internal/translator"
```

Add field inside `HltvFacade` struct:
```go
translateCfgFn func() (translator.TranslateConfig, error)
```

- [ ] **Step 2: Update New() signature**

Change from:
```go
func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient, store *storage.Store, notify func(string, int, string)) *HltvFacade {
	return &HltvFacade{
		cfg:    cfg,
		cache:  c,
		client: cli,
		store:  store,
		notify: notify,
		ts:     scraper.NewTeamScraper(cli),
		ps:     scraper.NewPlayerScraper(cli),
		rs:     scraper.NewResultsScraper(cli),
		ms:     scraper.NewMatchesScraper(cli),
		ns:     scraper.NewNewsScraper(cli),
		rns:    scraper.NewRealtimeNewsScraper(cli),
		nas:    scraper.NewNewsArticleScraper(cli),
	}
}
```

To:
```go
func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient, store *storage.Store, notify func(string, int, string), translateCfgFn func() (translator.TranslateConfig, error)) *HltvFacade {
	return &HltvFacade{
		cfg:            cfg,
		cache:          c,
		client:         cli,
		store:          store,
		notify:         notify,
		translateCfgFn: translateCfgFn,
		ts:             scraper.NewTeamScraper(cli),
		ps:             scraper.NewPlayerScraper(cli),
		rs:             scraper.NewResultsScraper(cli),
		ms:             scraper.NewMatchesScraper(cli),
		ns:             scraper.NewNewsScraper(cli),
		rns:            scraper.NewRealtimeNewsScraper(cli),
		nas:            scraper.NewNewsArticleScraper(cli),
	}
}
```

- [ ] **Step 3: Add translateNewTitles and translateNewRealtimeTitles methods**

Append after `ClearCache()`:

```go
// translateNewTitles translates titles for archive news items that don't yet
// have a translation stored, then pushes an SSE notification.
func (f *HltvFacade) translateNewTitles(items []types.NewsItem) {
	if f.translateCfgFn == nil || f.store == nil {
		return
	}
	cfg, err := f.translateCfgFn()
	if err != nil {
		log.Printf("facade: translate config: %v", err)
		return
	}
	t := translator.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	translated := 0
	for _, item := range items {
		if item.Title == "" || item.Link == "" {
			continue
		}
		if has, _, _ := f.store.HasNewsTitleZh(item.Link); has {
			continue
		}
		zh, err := t.TranslateTitle(ctx, item.Title)
		if err != nil {
			log.Printf("facade: translate title %q: %v", item.Title, err)
			continue
		}
		if err := f.store.UpdateNewsTitleZh(item.Link, zh); err != nil {
			log.Printf("facade: store title_zh: %v", err)
			continue
		}
		translated++
	}
	if translated > 0 {
		f.broadcast("news", 0, "")
	}
}

// translateNewRealtimeTitles translates titles for realtime news items.
func (f *HltvFacade) translateNewRealtimeTitles(items []types.RealtimeNewsItem) {
	if f.translateCfgFn == nil || f.store == nil {
		return
	}
	cfg, err := f.translateCfgFn()
	if err != nil {
		log.Printf("facade: translate config: %v", err)
		return
	}
	t := translator.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	translated := 0
	for _, item := range items {
		if item.Title == "" || item.Link == "" {
			continue
		}
		if has, _, _ := f.store.HasRealtimeTitleZh(item.Link); has {
			continue
		}
		zh, err := t.TranslateTitle(ctx, item.Title)
		if err != nil {
			log.Printf("facade: translate realtime title %q: %v", item.Title, err)
			continue
		}
		if err := f.store.UpdateRealtimeTitleZh(item.Link, zh); err != nil {
			log.Printf("facade: store realtime title_zh: %v", err)
			continue
		}
		translated++
	}
	if translated > 0 {
		f.broadcast("news", 0, "")
	}
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/facade/
```
Expected: no output (success)

- [ ] **Step 5: Commit**

```bash
git add internal/facade/facade.go
git commit -m "feat: add translateCfgFn and translation methods to facade"
```

---

### Task 7: Wire auto-translation into facade/news.go

**Files:**
- Modify: `internal/facade/news.go`

- [ ] **Step 1: Wire translateNewRealtimeTitles in GetRealtimeNews**

In `GetRealtimeNews` (line 27-78), after `BatchUpsertRealtimeNews(allItems)` (line 52), add:

```go
// Async translate new titles
if f.translateCfgFn != nil && f.store != nil {
    go f.translateNewRealtimeTitles(allItems)
}
```

The exact insertion point is after the error check on `BatchUpsertRealtimeNews`:
```go
if f.store != nil {
    if err := f.store.BatchUpsertRealtimeNews(allItems); err != nil {
        log.Printf("facade: batch upsert realtime news: %v", err)
    }
}
// ADD HERE:
if f.store != nil {
    go f.translateNewRealtimeTitles(allItems)
}
```

- [ ] **Step 2: Wire translateNewTitles in GetNewsDigest**

In `GetNewsDigest` (line 81-129), after `BatchUpsertNews(allItems)` (line 107), add:

```go
if f.store != nil {
    go f.translateNewTitles(allItems)
}
```

The exact insertion point:
```go
if f.store != nil {
    if err := f.store.BatchUpsertNews(allItems); err != nil {
        log.Printf("facade: batch upsert news: %v", err)
    }
}
// ADD HERE:
if f.store != nil {
    go f.translateNewTitles(allItems)
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/facade/
```
Expected: no output (success)

- [ ] **Step 4: Commit**

```bash
git add internal/facade/news.go
git commit -m "feat: wire auto title translation into news facade"
```

---

### Task 8: Update main.go — wire everything together

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add imports and create translateCfgFn**

Add import:
```go
"github.com/arcdent/hltv-mcp/internal/translator"
```

After `c := cache.New(...)` and before `cli := ...`, add:
```go
// Translation config factory for hot-reload-safe background translation
translateCfgFn := func() (translator.TranslateConfig, error) {
    return handlers.LoadTranslateConfig()
}
```

- [ ] **Step 2: Pass translateCfgFn to facade.New**

Change from:
```go
f := facade.New(cfg, c, cli, store, notify)
```
To:
```go
f := facade.New(cfg, c, cli, store, notify, translateCfgFn)
```

- [ ] **Step 3: Pass store to NewRouter**

Change from:
```go
router := httppkg.NewRouter(f, frontendFS, sseHub)
```
To:
```go
router := httppkg.NewRouter(f, frontendFS, sseHub, store)
```

- [ ] **Step 4: Verify full project compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./...
```
Expected: no output (success)

- [ ] **Step 5: Run existing tests**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./...
```
Expected: all existing tests pass

- [ ] **Step 6: Commit**

```bash
git add main.go
git commit -m "feat: wire translateCfgFn and store into main"
```

---

### Task 9: End-to-end verification

- [ ] **Step 1: Build the full binary**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build -o hltv-mcp .
```
Expected: no output (success)

- [ ] **Step 2: Start the service and verify schema migration**

```bash
./hltv-mcp &
sleep 2
# Check schema_version
sqlite3 data/hltv.db "SELECT * FROM schema_version;"
```
Expected: shows versions 1 and 2

- [ ] **Step 3: Verify new columns exist**

```bash
sqlite3 data/hltv.db ".schema news"
sqlite3 data/hltv.db ".schema realtime_news"
```
Expected: both schemas show `title_zh` column (and `body_text_zh` for news)

- [ ] **Step 4: Test translate config API**

```bash
curl -s http://localhost:8082/api/translate/config | jq .
```
Expected: returns current config or empty

- [ ] **Step 5: Test PostTranslate with body type and URL**

```bash
curl -s -X POST http://localhost:8082/api/translate \
  -H "Content-Type: application/json" \
  -d '{"text": "Astralis wins IEM Katowice", "type": "title"}' | jq .
```
Expected: returns `{"translated": "<中文翻译>"}`

- [ ] **Step 6: Kill the test server**

```bash
kill %1
```
