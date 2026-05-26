package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/errors"
)

const hltvBaseURL = "https://www.hltv.org"

// FallbackTracker remembers which endpoints recently needed chromedp
type FallbackTracker struct {
	mu       sync.RWMutex
	failures map[string]time.Time
	window   time.Duration
}

func NewFallbackTracker(windowSec int) *FallbackTracker {
	return &FallbackTracker{
		failures: make(map[string]time.Time),
		window:   time.Duration(windowSec) * time.Second,
	}
}

func (t *FallbackTracker) ShouldSkipHTTP(endpoint string) bool {
	t.mu.RLock()
	lastFail, ok := t.failures[endpoint]
	t.mu.RUnlock()
	return ok && time.Since(lastFail) < t.window
}

func (t *FallbackTracker) RecordFailure(endpoint string) {
	t.mu.Lock()
	t.failures[endpoint] = time.Now()
	t.mu.Unlock()
}

// HltvClient handles HTTP requests to HLTV with chromedp fallback
type HltvClient struct {
	cfg      *config.Config
	httpCli  *http.Client
	fallback *FallbackTracker
	chromeOK bool
}

func NewHltvClient(cfg *config.Config, chromeAvailable bool) *HltvClient {
	return &HltvClient{
		cfg:      cfg,
		chromeOK: chromeAvailable,
		httpCli:  &http.Client{Timeout: time.Duration(cfg.HTTPTimeoutMs) * time.Millisecond},
		fallback: NewFallbackTracker(300), // 5 minutes
	}
}

// FetchHTML returns the raw HTML body. Uses HTTP first, falls back to chromedp.
func (c *HltvClient) FetchHTML(ctx context.Context, path, endpointKey string) ([]byte, error) {
	shouldTryChromedp := c.chromeOK && c.cfg.DataSource != config.DataSourceDirect

	if shouldTryChromedp && c.fallback.ShouldSkipHTTP(endpointKey) {
		return c.fetchChromedp(ctx, path)
	}

	body, err := c.fetchHTTP(ctx, path)
	if err == nil && !isCloudflareBlock(body) {
		return body, nil
	}

	if err != nil || isCloudflareBlock(body) {
		c.fallback.RecordFailure(endpointKey)
		if shouldTryChromedp {
			return c.fetchChromedp(ctx, path)
		}
		if err != nil {
			return nil, err
		}
		return body, nil // return block page if no fallback available
	}
	return body, nil
}

func (c *HltvClient) fetchHTTP(ctx context.Context, path string) ([]byte, error) {
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
		if resp.StatusCode == 404 {
			return nil, errors.New(errors.CodeUpstreamNotFound, fmt.Sprintf("404 for %s", path), false, nil)
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}
		return body, nil
	}
	return nil, errors.New(errors.CodeUpstreamUnavailable,
		fmt.Sprintf("failed after %d retries: %v", c.cfg.RetryCount, lastErr), true, nil)
}

func isCloudflareBlock(body []byte) bool {
	s := string(body)
	return strings.Contains(s, "Just a moment") ||
		strings.Contains(s, "cf-browser-verify") ||
		strings.Contains(s, "Attention Required") ||
		strings.Contains(s, "Cloudflare")
}

func (c *HltvClient) IsChromeAvailable() bool { return c.chromeOK }
