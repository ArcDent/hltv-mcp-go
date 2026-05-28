package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	if err := os.WriteFile(configPath(), data, 0600); err != nil {
		return err
	}
	// Remove the old plaintext config so the key isn't left in two places
	os.Remove(oldPath)
	return nil
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

// LoadTranslateConfig exposes config loading for use by other packages.
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
		log.Printf("translate proxy: LLM API returned %d: %s", resp.StatusCode, string(respBody))
		writeError(w, http.StatusBadGateway, fmt.Sprintf("翻译服务返回错误(%d)", resp.StatusCode))
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
