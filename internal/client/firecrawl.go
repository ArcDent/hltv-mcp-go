package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const firecrawlAPI = "https://api.firecrawl.dev/v1/scrape"

type firecrawlReq struct {
	URL     string   `json:"url"`
	Formats []string `json:"formats"`
	Timeout int      `json:"timeout"`
}

type firecrawlResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Data    struct {
		RawHTML string `json:"rawHtml"`
	} `json:"data"`
}

// FetchViaFirecrawl scrapes a URL using Firecrawl API (bypasses Cloudflare)
func (c *HltvClient) FetchViaFirecrawl(ctx context.Context, path string) ([]byte, error) {
	if c.cfg.FirecrawlKey == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not configured")
	}

	url := hltvBaseURL + path
	reqBody := firecrawlReq{
		URL:     url,
		Formats: []string{"rawHtml"},
		Timeout: 90000, // 90s — HLTV matches page is large
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", firecrawlAPI, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.FirecrawlKey)
	httpReq.Header.Set("Content-Type", "application/json")

	cli := &http.Client{Timeout: 120 * time.Second}
	resp, err := cli.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("firecrawl request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fcResp firecrawlResp
	if err := json.Unmarshal(respBytes, &fcResp); err != nil {
		return nil, fmt.Errorf("firecrawl parse error: %w", err)
	}
	if !fcResp.Success {
		return nil, fmt.Errorf("firecrawl: %s", fcResp.Error)
	}
	if fcResp.Data.RawHTML == "" {
		return nil, fmt.Errorf("firecrawl returned empty response")
	}

	return []byte(fcResp.Data.RawHTML), nil
}
