package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/types"
)

const hltvBaseURL = "https://www.hltv.org"

// HltvClient handles HTTP requests to HLTV
type HltvClient struct {
	cfg     *config.Config
	httpCli *http.Client
}

func NewHltvClient(cfg *config.Config) *HltvClient {
	return &HltvClient{
		cfg:     cfg,
		httpCli: &http.Client{Timeout: time.Duration(cfg.HTTPTimeoutMs) * time.Millisecond},
	}
}

// FetchHTML returns the raw HTML body from HLTV
func (c *HltvClient) FetchHTML(ctx context.Context, path, endpointKey string) ([]byte, error) {
	url := hltvBaseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	var lastErr error
	for attempt := 0; attempt <= c.cfg.RetryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		resp, err := c.httpCli.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == 403 || resp.StatusCode == 404 {
			return nil, &types.ToolError{
				Code:    "UPSTREAM_NOT_FOUND",
				Message: fmt.Sprintf("%d for %s", resp.StatusCode, path),
			}
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}
		if isCloudflareBlock(body) {
			return nil, &types.ToolError{
				Code:      "UPSTREAM_UNAVAILABLE",
				Message:   "HLTV returned Cloudflare challenge page",
				Retryable: true,
			}
		}
		return body, nil
	}
	return nil, &types.ToolError{
		Code:      "UPSTREAM_UNAVAILABLE",
		Message:   fmt.Sprintf("failed after %d retries: %v", c.cfg.RetryCount, lastErr),
		Retryable: true,
	}
}

func isCloudflareBlock(body []byte) bool {
	s := string(body)
	return strings.Contains(s, "Just a moment") ||
		strings.Contains(s, "cf-browser-verify") ||
		strings.Contains(s, "Attention Required") ||
		strings.Contains(s, "Cloudflare")
}
