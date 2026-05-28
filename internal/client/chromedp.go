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
		// DefaultExecAllocatorOptions includes Headless (--headless), which
		// breaks headless-shell (already headless, flag prevents startup).
		// Override with false; chromedp's Allocate() skips bool-flags set to false.
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

	// Start a browser tab with a 10s deadline. When the allocator's Chrome
	// process is dead, chromedp.NewContext hangs forever trying to connect
	// to DevTools. The goroutine+select pattern avoids parent-context
	// cancellation propagation issues.
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
	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var html string
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait up to 20s for CF challenge to resolve
			for i := 0; i < 20; i++ {
				var title string
				chromedp.Title(&title).Do(ctx)
				if title != "" && !strings.Contains(title, "Just a moment") && !strings.Contains(title, "Attention Required") {
					return nil
				}
				time.Sleep(1 * time.Second)
			}
			return nil // proceed even if still on challenge page
		}),
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
