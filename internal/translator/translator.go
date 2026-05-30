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
