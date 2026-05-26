package client

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/arcdent/hltv-mcp/internal/config"
)

func findChromePath(cfgPath string) (string, bool) {
	if cfgPath != "" {
		return cfgPath, true
	}
	for _, p := range []string{"google-chrome", "chromium", "chromium-browser", "chrome", "chrome-headless-shell"} {
		if _, err := exec.LookPath(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func (c *HltvClient) fetchChromedp(ctx context.Context, path string) ([]byte, error) {
	if !c.chromeOK {
		return nil, fmt.Errorf("chromedp not available")
	}
	url := hltvBaseURL + path
	taskCtx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var html string
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &html),
	); err != nil {
		return nil, err
	}
	return []byte(html), nil
}

// CheckChromeAvailable returns the Chrome path and whether it is usable
func CheckChromeAvailable(cfg *config.Config) (path string, available bool) {
	if cfg.DataSource == "direct" {
		return "", false
	}
	path, ok := findChromePath(cfg.ChromePath)
	if !ok {
		return "", false
	}
	return path, true
}
