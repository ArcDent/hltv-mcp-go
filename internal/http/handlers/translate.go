package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const translateConfigFile = "translate_config.json"

type TranslateConfig struct {
	ProviderURL string `json:"provider_url"`
	APIKey      string `json:"api_key"`
	Model       string `json:"model"`
}

func configPath() string {
	exec, _ := os.Executable()
	dir := filepath.Dir(exec)
	if dir == "" || dir == "." {
		dir, _ = os.Getwd()
	}
	return filepath.Join(dir, translateConfigFile)
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
	return cfg, nil
}

func saveTranslateConfig(cfg TranslateConfig) error {
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
