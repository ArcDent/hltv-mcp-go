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

// chromedpAllocCtx is a persistent Chrome allocator shared by all chromedp fetches
var chromedpAllocCtx context.Context
var chromedpAllocCancel context.CancelFunc

func (c *HltvClient) fetchChromedp(ctx context.Context, path string) ([]byte, error) {
	if !c.chromeOK {
		return nil, fmt.Errorf("chromedp not available")
	}

	if chromedpAllocCtx == nil {
		chromePath, _ := findChromePath(c.cfg.ChromePath)
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(chromePath),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"),
		)
		chromedpAllocCtx, chromedpAllocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}

	url := hltvBaseURL + path
	taskCtx, cancel := chromedp.NewContext(chromedpAllocCtx)
	defer cancel()
	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var html string
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
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
