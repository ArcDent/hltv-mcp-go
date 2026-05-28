# Translate Persistence & Encryption Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Encrypt API keys at rest with AES-256-GCM, proxy translation requests through backend, auto-manage encryption keys for one-click Docker deployment.

**Architecture:** New `internal/crypto` package handles AES-256-GCM encrypt/decrypt with SHA-256 key derivation from passphrase. Translate handler gains encrypted read/write, old-config migration, and a `POST /api/translate` proxy endpoint. Frontend drops `sessionStorage`/`realKey` and calls the backend instead of the LLM directly.

**Tech Stack:** Go 1.26 stdlib (`crypto/aes`, `crypto/cipher`, `crypto/sha256`, `crypto/rand`), existing chi router, React/TypeScript frontend.

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/crypto/crypto.go` (new) | AES-256-GCM encrypt/decrypt, key init from ENV/file/auto-gen |
| `internal/crypto/crypto_test.go` (new) | Unit tests for encrypt/decrypt roundtrip and edge cases |
| `main.go` (modify) | Call `crypto.InitKey()` + `handlers.MigrateConfig()` on startup |
| `internal/http/handlers/translate.go` (modify) | Config path to `data/`, encrypt on save, decrypt on load, POST /api/translate handler, old-config migration |
| `internal/http/router.go` (modify) | Register `POST /api/translate` route |
| `frontend/src/components/TranslateProvider.tsx` (modify) | Remove `realKey`/`sessionStorage`, simplify hook |
| `frontend/src/components/NewsDetail.tsx` (modify) | Call `/api/translate` instead of direct LLM |
| `frontend/src/pages/News.tsx` (modify) | Call `/api/translate` instead of direct LLM |
| `Dockerfile` (modify) | `mkdir /data`, `VOLUME ["/data"]`, `WORKDIR /` |
| `docker-compose.yml` (modify) | Mount `./data:/data` volume |

---

### Task 1: Create crypto module

**Files:**
- Create: `internal/crypto/crypto.go`

- [ ] **Step 1: Write `internal/crypto/crypto.go`**

```go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var aesKey []byte

const keyFilePath = "data/.encryption_key"

// InitKey loads or generates the encryption passphrase, derives a 32-byte AES key,
// and stores it in the package-level aesKey variable. Must be called once at startup.
func InitKey() error {
	// 1. ENCRYPTION_KEY env var
	if key := os.Getenv("ENCRYPTION_KEY"); key != "" {
		h := sha256.Sum256([]byte(key))
		aesKey = h[:]
		return nil
	}
	// 2. data/.encryption_key file
	if data, err := os.ReadFile(keyFilePath); err == nil {
		h := sha256.Sum256(data)
		aesKey = h[:]
		return nil
	}
	// 3. Auto-generate
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("generate encryption key: %w", err)
	}
	passphrase := hex.EncodeToString(randomBytes)
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(keyFilePath, []byte(passphrase), 0600); err != nil {
		return fmt.Errorf("write .encryption_key: %w", err)
	}
	h := sha256.Sum256([]byte(passphrase))
	aesKey = h[:]
	return nil
}

// Encrypt encrypts plaintext with AES-256-GCM and returns base64(iv + ciphertext + tag).
func Encrypt(plaintext string) (string, error) {
	if len(aesKey) == 0 {
		return "", errors.New("crypto not initialized: call InitKey first")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)
	result := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts a base64(iv + ciphertext + tag) string with AES-256-GCM.
func Decrypt(encoded string) (string, error) {
	if len(aesKey) == 0 {
		return "", errors.New("crypto not initialized: call InitKey first")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	iv := data[:nonceSize]
	ct := data[nonceSize:]
	plaintext, err := gcm.Open(nil, iv, ct, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/crypto/crypto.go
git commit -m "feat: add AES-256-GCM crypto module with auto key generation"
```

---

### Task 2: Test crypto module

**Files:**
- Create: `internal/crypto/crypto_test.go`

- [ ] **Step 1: Write `internal/crypto/crypto_test.go`**

```go
package crypto

import (
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase-for-unit-tests")
	if err := InitKey(); err != nil {
		t.Fatal(err)
	}

	plaintext := "sk-test-api-key-1234567890abcdef"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "" || encrypted == plaintext {
		t.Error("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase")
	InitKey()

	if _, err := Decrypt("!!not-valid-base64!!"); err == nil {
		t.Error("expected error for invalid base64")
	}
	if _, err := Decrypt(""); err == nil {
		t.Error("expected error for empty input")
	}
}

func TestDecryptTooShort(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase")
	InitKey()

	if _, err := Decrypt("YWJj"); err == nil {
		t.Error("expected error for too-short ciphertext (abc in base64)")
	}
}

func TestEncryptWithoutInit(t *testing.T) {
	aesKey = nil
	if _, err := Encrypt("test"); err == nil {
		t.Error("expected error when not initialized")
	}
}

func TestDecryptWithoutInit(t *testing.T) {
	aesKey = nil
	if _, err := Decrypt("dGVzdA=="); err == nil {
		t.Error("expected error when not initialized")
	}
}

func TestDeterministicWithSameKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase")
	InitKey()

	plaintext := "my-secret-key"
	enc1, _ := Encrypt(plaintext)
	enc2, _ := Encrypt(plaintext)

	// Different IV each time means different ciphertext
	if enc1 == enc2 {
		t.Error("encryptions should produce different ciphertext due to random IV")
	}

	// But both should decrypt to the same plaintext
	dec1, _ := Decrypt(enc1)
	dec2, _ := Decrypt(enc2)
	if dec1 != plaintext || dec2 != plaintext {
		t.Error("both should decrypt to original plaintext")
	}
}

func TestInitKeyFilePersistence(t *testing.T) {
	aesKey = nil
	t.Setenv("ENCRYPTION_KEY", "")
	// Don't write to real data/.encryption_key; the test should trigger
	// auto-generation but we can't easily test file write in unit tests
	// without filesystem isolation. Test that ENCRYPTION_KEY env works.
	t.Setenv("ENCRYPTION_KEY", "from-env")
	if err := InitKey(); err != nil {
		t.Fatal(err)
	}
	if len(aesKey) != 32 {
		t.Errorf("expected 32-byte AES key, got %d", len(aesKey))
	}
}
```

- [ ] **Step 2: Run tests and verify pass**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/crypto/ -v
```
Expected: all tests PASS

- [ ] **Step 3: Commit**

```bash
git add internal/crypto/crypto_test.go
git commit -m "test: add crypto module unit tests"
```

---

### Task 3: Initialize crypto and migrate config on startup

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add crypto init and config migration to `main.go`**

Current `main.go` (relevant section, lines 1-35):
```go
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/facade"
	httppkg "github.com/arcdent/hltv-mcp/internal/http"
	"github.com/arcdent/hltv-mcp/internal/mcp"
	"github.com/arcdent/hltv-mcp/internal/renderer"
	"github.com/arcdent/hltv-mcp/internal/summary"
)

//go:embed dist/*
var embeddedFrontend embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("HLTV MCP starting...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
```

Replace with (add imports + init calls after config load):
```go
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/crypto"
	"github.com/arcdent/hltv-mcp/internal/facade"
	httppkg "github.com/arcdent/hltv-mcp/internal/http"
	"github.com/arcdent/hltv-mcp/internal/http/handlers"
	"github.com/arcdent/hltv-mcp/internal/mcp"
	"github.com/arcdent/hltv-mcp/internal/renderer"
	"github.com/arcdent/hltv-mcp/internal/summary"
)

//go:embed dist/*
var embeddedFrontend embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("HLTV MCP starting...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Initialize encryption key (env → file → auto-generate)
	if err := crypto.InitKey(); err != nil {
		log.Fatalf("crypto init: %v", err)
	}

	// Migrate old config to encrypted data/ directory
	if err := handlers.MigrateConfig(); err != nil {
		log.Printf("config migration note: %v", err)
	}
```

**Change details:**
Add to imports (two additions):
```go
	"github.com/arcdent/hltv-mcp/internal/crypto"
```
```go
	"github.com/arcdent/hltv-mcp/internal/http/handlers"
```

Add after config load (before Chrome detection):
```go
	// Initialize encryption key (env → file → auto-generate)
	if err := crypto.InitKey(); err != nil {
		log.Fatalf("crypto init: %v", err)
	}

	// Migrate old config to encrypted data/ directory
	if err := handlers.MigrateConfig(); err != nil {
		log.Printf("config migration note: %v", err)
	}
```

- [ ] **Step 2: Verify compilation fails (missing MigrateConfig)**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./... 2>&1 | head -5
```
Expected: `handlers.MigrateConfig undefined` — will be added in Task 4

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: add crypto init and config migration call on startup"
```

---

### Task 4: Rewrite translate handler with encryption, migration, and proxy endpoint

**Files:**
- Modify: `internal/http/handlers/translate.go`

- [ ] **Step 1: Replace `translate.go` with full new implementation**

Delete all existing content and write:

```go
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/arcdent/hltv-mcp/internal/crypto"
)

const (
	translateConfigFile = "translate_config.json"
	dataDir             = "data"
)

type TranslateConfig struct {
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
		return nil // new config already exists
	}
	oldPath := oldConfigPath()
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return nil // no old config to migrate
	}
	var cfg TranslateConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	if cfg.APIKey == "" {
		return nil
	}
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	cfg.Encrypted = true
	encryptedKey, err := crypto.Encrypt(cfg.APIKey)
	if err != nil {
		return fmt.Errorf("encrypt key: %w", err)
	}
	cfg.APIKey = encryptedKey
	data, err = json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

func loadTranslateConfig() (TranslateConfig, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return TranslateConfig{}, err
	}
	var cfg TranslateConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return TranslateConfig{}, err
	}
	if cfg.Encrypted {
		key, err := crypto.Decrypt(cfg.APIKey)
		if err != nil {
			return TranslateConfig{}, fmt.Errorf("decrypt api key: %w", err)
		}
		cfg.APIKey = key
	} else if cfg.APIKey != "" {
		// Auto-upgrade plaintext config to encrypted
		cfg.Encrypted = true
		encryptedKey, err := crypto.Encrypt(cfg.APIKey)
		if err == nil {
			upgraded := cfg
			upgraded.APIKey = encryptedKey
			if data, err := json.MarshalIndent(upgraded, "", "  "); err == nil {
				os.WriteFile(configPath(), data, 0600)
			}
		}
	}
	return cfg, nil
}

func saveTranslateConfig(cfg TranslateConfig) error {
	if cfg.APIKey != "" && !strings.Contains(cfg.APIKey, "***") {
		encryptedKey, err := crypto.Encrypt(cfg.APIKey)
		if err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}
		cfg.APIKey = encryptedKey
		cfg.Encrypted = true
	}
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
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

// Public to allow use in NewsArticle translation (if needed via facade)
func LoadTranslateConfig() (TranslateConfig, error) {
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
	var cfg TranslateConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.Contains(cfg.APIKey, "***") {
		existing, err := loadTranslateConfig()
		if err == nil {
			cfg.APIKey = existing.APIKey
		}
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

	systemPrompt := "将以下CS电竞新闻正文翻译为简体中文"
	if req.Type == "title" {
		systemPrompt = "将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释"
	}

	llmReq := map[string]any{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Text},
		},
		"temperature": 0.1,
	}

	body, _ := json.Marshal(llmReq)
	url := strings.TrimRight(cfg.ProviderURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "构造请求失败")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, "翻译服务请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "读取翻译响应失败")
		return
	}
	if resp.StatusCode != http.StatusOK {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("翻译服务返回错误(%d): %s", resp.StatusCode, string(respBody)))
		return
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		writeError(w, http.StatusBadGateway, "翻译结果解析失败")
		return
	}
	if len(result.Choices) == 0 {
		writeError(w, http.StatusBadGateway, "翻译服务未返回结果")
		return
	}

	writeJSON(w, map[string]string{"translated": strings.TrimSpace(result.Choices[0].Message.Content)})
}
```

- [ ] **Step 2: Verify compilation passes**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./...
```
Expected: no errors (missing route registration is fine for compilation)

- [ ] **Step 3: Commit**

```bash
git add internal/http/handlers/translate.go
git commit -m "feat: add encrypted config storage, migration, and translate proxy endpoint"
```

---

### Task 5: Register POST /api/translate route

**Files:**
- Modify: `internal/http/router.go`

- [ ] **Step 1: Add route in `router.go`**

Current `router.go` line 45:
```go
	r.Put("/api/translate/config", h.PutTranslateConfig)
```

After line 45, add:
```go
	r.Post("/api/translate", h.PostTranslate)
```

- [ ] **Step 2: Verify compilation passes**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./...
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/http/router.go
git commit -m "feat: register POST /api/translate route"
```

---

### Task 6: Simplify TranslateProvider (remove realKey/sessionStorage)

**Files:**
- Modify: `frontend/src/components/TranslateProvider.tsx`

- [ ] **Step 1: Replace TranslateProvider.tsx**

Current file lines 1-43 (the hook and config type):
```typescript
import { useState, useEffect } from 'react'
import Modal from './Modal'

const PRESETS = [
  { label: 'OpenAI',         url: 'https://api.openai.com/v1',        model: 'gpt-4o-mini' },
  { label: 'DeepSeek',       url: 'https://api.deepseek.com/v1',      model: 'deepseek-chat' },
  { label: 'Groq',           url: 'https://api.groq.com/openai/v1',   model: 'llama-3.3-70b-versatile' },
  { label: 'Ollama 本地',    url: 'http://localhost:11434/v1',        model: 'qwen2.5:7b' },
]

type Config = { provider_url: string; api_key: string; model: string; configured: boolean }

export function useTranslateConfig() {
  const [cfg, setCfg] = useState<Config | null>(null)
  const [realKey, setRealKey] = useState(() => sessionStorage.getItem('hltv_real_key') ?? '')
  const [saveCount, setSaveCount] = useState(0)
  const [open, setOpen] = useState(false)

  const fetchConfig = async () => {
    try {
      const r = await fetch('/api/translate/config')
      const c = await r.json()
      setCfg(c)
    } catch { setCfg({ provider_url: '', api_key: '', model: '', configured: false } as Config) }
  }

  useEffect(() => { fetchConfig() }, [])

  const save = async (url: string, key: string, model: string) => {
    sessionStorage.setItem('hltv_real_key', key)
    setRealKey(key)
    await fetch('/api/translate/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ provider_url: url, api_key: key, model }),
    })
    await fetchConfig()
    setSaveCount(c => c + 1)
    setOpen(false)
  }

  return { cfg, realKey, save, open, setOpen, saveCount }
}
```

Replace with:
```typescript
import { useState, useEffect } from 'react'
import Modal from './Modal'

const PRESETS = [
  { label: 'OpenAI',         url: 'https://api.openai.com/v1',        model: 'gpt-4o-mini' },
  { label: 'DeepSeek',       url: 'https://api.deepseek.com/v1',      model: 'deepseek-chat' },
  { label: 'Groq',           url: 'https://api.groq.com/openai/v1',   model: 'llama-3.3-70b-versatile' },
  { label: 'Ollama 本地',    url: 'http://localhost:11434/v1',        model: 'qwen2.5:7b' },
]

type Config = { provider_url: string; api_key: string; model: string; configured: boolean }

export function useTranslateConfig() {
  const [cfg, setCfg] = useState<Config | null>(null)
  const [saveCount, setSaveCount] = useState(0)
  const [open, setOpen] = useState(false)

  const fetchConfig = async () => {
    try {
      const r = await fetch('/api/translate/config')
      const c = await r.json()
      setCfg(c)
    } catch { setCfg({ provider_url: '', api_key: '', model: '', configured: false } as Config) }
  }

  useEffect(() => { fetchConfig() }, [])

  const save = async (url: string, key: string, model: string) => {
    await fetch('/api/translate/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ provider_url: url, api_key: key, model }),
    })
    await fetchConfig()
    setSaveCount(c => c + 1)
    setOpen(false)
  }

  return { cfg, save, open, setOpen, saveCount }
}
```

Keep the `TranslateModal` component unchanged (lines 44-106).

- [ ] **Step 2: Verify frontend compiles**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx tsc --noEmit 2>&1 | head -20
```
Expected: errors in News.tsx and NewsDetail.tsx referencing `realKey` (will fix next)

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/TranslateProvider.tsx
git commit -m "refactor: remove realKey and sessionStorage from TranslateProvider"
```

---

### Task 7: Update NewsDetail to use backend translate proxy

**Files:**
- Modify: `frontend/src/components/NewsDetail.tsx`

- [ ] **Step 1: Replace the `doTranslate` function**

Current `NewsDetail.tsx` lines 14 and 37-65:
```typescript
  const { cfg, realKey } = useTranslateConfig()

  // ... (keep everything else the same until doTranslate)

  const doTranslate = async () => {
    if (!data?.body_text || !cfg?.configured) return
    setTranslating(true)
    try {
      const res = await fetch(`${cfg!.provider_url}/chat/completions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${realKey}` },
        body: JSON.stringify({
          model: cfg!.model,
          messages: [
            { role: 'system', content: '将以下CS电竞新闻正文翻译为简体中文' },
            { role: 'user', content: data.body_text.slice(0, 8000) },
          ],
          temperature: 0.1,
        }),
      })
      const body = await res.text()
      if (!res.ok) throw new Error(body)
      const j = JSON.parse(body)
      const zh = (j?.choices?.[0]?.message?.content as string)?.trim() ?? ''
      if (zh) {
        setTranslated(zh)
        try {
          let hash = 0; for (let i = 0; i < url.length; i++) { hash = (hash * 31 + url.charCodeAt(i)) >>> 0 }
          localStorage.setItem(`news_trans:${hash.toString(16)}`, JSON.stringify({ zh, ts: Date.now() }))
        } catch { /* ignore storage errors */ }
      }
    } catch (e) { console.error('translate article failed:', e) }
    setTranslating(false)
  }
```

Change line 14 to:
```typescript
  const { cfg } = useTranslateConfig()
```

Replace the `doTranslate` function with:
```typescript
  const doTranslate = async () => {
    if (!data?.body_text || !cfg?.configured) return
    setTranslating(true)
    try {
      const res = await fetch('/api/translate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: data.body_text.slice(0, 8000), type: 'article' }),
      })
      if (!res.ok) {
        const errBody = await res.text()
        throw new Error(errBody)
      }
      const j = await res.json()
      const zh = j?.translated ?? ''
      if (zh) {
        setTranslated(zh)
        try {
          let hash = 0; for (let i = 0; i < url.length; i++) { hash = (hash * 31 + url.charCodeAt(i)) >>> 0 }
          localStorage.setItem(`news_trans:${hash.toString(16)}`, JSON.stringify({ zh, ts: Date.now() }))
        } catch { /* ignore storage errors */ }
      }
    } catch (e) { console.error('translate article failed:', e) }
    setTranslating(false)
  }
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/NewsDetail.tsx
git commit -m "refactor: use backend translate proxy in NewsDetail"
```

---

### Task 8: Update News page to use backend translate proxy

**Files:**
- Modify: `frontend/src/pages/News.tsx`

- [ ] **Step 1: Update line 25 to remove `realKey` from destructuring**

Current line 25:
```typescript
  const { cfg, realKey, save, open, setOpen, saveCount } = useTranslateConfig()
```

Replace with:
```typescript
  const { cfg, save, open, setOpen, saveCount } = useTranslateConfig()
```

- [ ] **Step 2: Replace the translate fetch call (lines 63-75)**

Current lines 63-75:
```typescript
        try {
          const res = await fetch(`${cfg!.provider_url}/chat/completions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${realKey}` },
            body: JSON.stringify({
              model: cfg!.model,
              messages: [
                { role: 'system', content: '将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释' },
                { role: 'user', content: title },
              ],
              temperature: 0.1,
            }),
          })
```

Replace with:
```typescript
        try {
          const res = await fetch('/api/translate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ text: title, type: 'title' }),
          })
```

- [ ] **Step 3: Update response parsing (lines 76-83)**

Current lines 76-83:
```typescript
          const body = await res.text()
          if (!res.ok) { console.error('translate API error', res.status, body); throw new Error(body) }
          const j = JSON.parse(body)
          const zh = (j?.choices?.[0]?.message?.content as string)?.trim() ?? ''
          if (zh) {
            cache[hashTitle(title)] = { zh, ts: Date.now() }
            saveCache(cache)
            setTranslations(prev => ({ ...prev, [title]: zh }))
          }
```

Replace with:
```typescript
          const body = await res.text()
          if (!res.ok) { console.error('translate API error', res.status, body); throw new Error(body) }
          const j = JSON.parse(body)
          const zh = j?.translated ?? ''
          if (zh) {
            cache[hashTitle(title)] = { zh, ts: Date.now() }
            saveCache(cache)
            setTranslations(prev => ({ ...prev, [title]: zh }))
          }
```

- [ ] **Step 4: Verify frontend compiles cleanly**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx tsc --noEmit
```
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/News.tsx
git commit -m "refactor: use backend translate proxy in News page"
```

---

### Task 9: Update Docker deployment for persistent data volume

**Files:**
- Modify: `Dockerfile`
- Modify: `docker-compose.yml`

- [ ] **Step 1: Update `Dockerfile`**

Current Dockerfile:
```dockerfile
# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./dist/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o hltv-mcp github.com/arcdent/hltv-mcp

# Stage 3: Runtime
# chromedp/headless-shell provides a headless Chrome instance for chromedp
FROM chromedp/headless-shell:latest
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/headless-shell/headless-shell
ENTRYPOINT ["/hltv-mcp"]
```

Replace Stage 3 (lines 21-27) with:
```dockerfile
# Stage 3: Runtime
# chromedp/headless-shell provides a headless Chrome instance for chromedp
FROM chromedp/headless-shell:latest
WORKDIR /
RUN mkdir -p /data
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/headless-shell/headless-shell
VOLUME ["/data"]
ENTRYPOINT ["/hltv-mcp"]
```

- [ ] **Step 2: Update `docker-compose.yml`**

Current:
```yaml
services:
  hltv-mcp:
    build: .
    ports:
      - "8082:8082"
    environment:
      - HTTP_PORT=8082
      - HTTP_HOST=0.0.0.0
      - HLTV_CHROME_PATH=/headless-shell/headless-shell
    volumes:
      - ./translate_config.json:/translate_config.json
    restart: unless-stopped
```

Replace with:
```yaml
services:
  hltv-mcp:
    build: .
    ports:
      - "8082:8082"
    environment:
      - HTTP_PORT=8082
      - HTTP_HOST=0.0.0.0
      - HLTV_CHROME_PATH=/headless-shell/headless-shell
      # ENCRYPTION_KEY is optional; if not set, auto-generated into ./data/.encryption_key
    volumes:
      - ./data:/data
    restart: unless-stopped
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile docker-compose.yml
git commit -m "feat: add persistent /data volume for encrypted config and keys"
```

---

### Task 10: Build and smoke test

- [ ] **Step 1: Build Go binary**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build -ldflags="-s -w" -o hltv-mcp .
```
Expected: successful build

- [ ] **Step 2: Build frontend**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npm run build
```
Expected: successful build

- [ ] **Step 3: Run Go tests**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/... -v
```
Expected: all tests PASS

- [ ] **Step 4: Start the server and test endpoints**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && ./hltv-mcp &
sleep 2

# Test config not configured yet
curl -s http://localhost:8082/api/translate/config | python3 -m json.tool
# Expected: {"api_key":"","configured":false,"model":"","provider_url":""}

# Test translate without config
curl -s -X POST http://localhost:8082/api/translate -H 'Content-Type: application/json' -d '{"text":"hello","type":"title"}'
# Expected: {"error":"翻译服务未配置"}

# Check key file was auto-generated
ls -la data/.encryption_key && cat data/.encryption_key
# Expected: 64-char hex string in data/.encryption_key

kill %1 2>/dev/null
```
Expected: all commands return expected output

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: rebuld binaries after translate persistence feature"
```
