package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		userDir, _ := os.MkdirTemp("", "hltv-chrome-*")
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(chromePath),
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("disable-features", "TranslateUI,BlinkGenPropertyTrees"),
			chromedp.WindowSize(1920, 1080),
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"),
			chromedp.UserDataDir(userDir),
		)
		chromedpAllocCtx, chromedpAllocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}

	url := hltvBaseURL + path

	// Start a browser tab with a 10s deadline
	type newCtxResult struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
	ch := make(chan newCtxResult, 1)
	go func() {
		ctx, cancel := chromedp.NewContext(chromedpAllocCtx)
		ch <- newCtxResult{ctx, cancel}
	}()
	var taskCtx context.Context
	var cancel context.CancelFunc
	select {
	case res := <-ch:
		taskCtx, cancel = res.ctx, res.cancel
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("chromedp: NewContext timed out (Chrome may be dead)")
	}
	defer cancel()
	taskCtx, cancel = context.WithTimeout(taskCtx, 90*time.Second)
	defer cancel()

	// CF challenge JS often causes ERR_ABORTED navigation errors.
	// Use recover to prevent crashes from nil cdp.Executor after tab closes.
	var html string
	func() {
		defer func() { recover() }()
		chromedp.Run(taskCtx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				chromedp.Navigate(url).Do(ctx)
				return nil
			}),
		)
	}()

	// Poll for real HLTV content — small pages are CF challenges
	for i := 0; i < 120; i++ {
		if html != "" {
			break
		}
		select {
		case <-taskCtx.Done():
			break
		default:
		}
		func() {
			defer func() { recover() }()
			var body string
			if err := chromedp.OuterHTML("html", &body).Do(taskCtx); err == nil &&
				len(body) > 2000 &&
				!strings.Contains(body, "Just a moment") &&
				!strings.Contains(body, "cf-browser-verify") {
				html = body
			}
		}()
		time.Sleep(500 * time.Millisecond)
	}

	if html == "" || len(html) < 500 {
		return nil, fmt.Errorf("chromedp: page load failed (CF challenge may have blocked)")
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
