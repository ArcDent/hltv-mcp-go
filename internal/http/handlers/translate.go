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
		encryptedKey, err := crypto.Encrypt(fcfg.APIKey)
		if err == nil {
			upgraded := fileConfig{
				ProviderURL: fcfg.ProviderURL,
				APIKey:      encryptedKey,
				Model:       fcfg.Model,
				Encrypted:   true,
			}
			if data, err := json.MarshalIndent(upgraded, "", "  "); err == nil {
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
