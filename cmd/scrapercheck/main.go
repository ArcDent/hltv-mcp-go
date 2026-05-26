package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/normalizer"
)

func main() {
	cfg := &config.Config{
		HTTPTimeoutMs: 15000, RetryCount: 1, DataSource: config.DataSourceAuto,
		CacheMaxEntries: 100, CacheStaleWindowSec: 3600, Timezone: "Asia/Shanghai",
	}

	// === Debug news archive (.newstext structure) ===
	fmt.Println("=== NEWS ARCHIVE DEBUG ===")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cli := client.NewHltvClient(cfg, false) // HTTP only
	body, err := cli.FetchHTML(ctx, "/news/archive/2026/May", "news_debug")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	// Inspect first few .newstext elements
	doc.Find(".newstext").Each(func(i int, s *goquery.Selection) {
		if i >= 2 { return }
		html, _ := s.Html()
		fmt.Printf("\n.newstext[%d] HTML:\n%s\n", i, strings.TrimSpace(html[:min(len(html), 400)]))
		a := s.Find("a").First()
		fmt.Printf("  a text: %q\n", a.Text())
		href, _ := a.Attr("href")
		fmt.Printf("  a href: %q\n", href)
		// Check parent
		parentHtml, _ := s.Parent().Html()
		if len(parentHtml) > 300 { parentHtml = parentHtml[:300] }
		fmt.Printf("  parent HTML: %s\n", strings.TrimSpace(parentHtml))
	})

	// Test normalized parsing
	items := normalizer.NormalizeNews(doc)
	fmt.Printf("\nNormalizeNews count: %d\n", len(items))
	for i, item := range items {
		if i >= 3 { break }
		fmt.Printf("  [%d] %q\n", i, item.Title)
	}

	// === Chromedp test with proper allocator ===
	fmt.Println("\n=== CHROMEDP WITH ALLOCATOR ===")
	_, chromeOK := client.CheckChromeAvailable(cfg)
	if !chromeOK {
		fmt.Println("Chrome not available, skipping")
		return
	}

	// Use a proper persistent allocator
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx2, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()
	ctx2, cancel2 = context.WithTimeout(ctx2, 30*time.Second)
	defer cancel2()

	var html string
	err = chromedp.Run(ctx2,
		chromedp.Navigate("https://www.hltv.org/matches"),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3*time.Second), // Wait for JS rendering
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		fmt.Printf("Chromedp error: %v\n", err)
		return
	}
	fmt.Printf("Chromedp HTML size: %d bytes\n", len(html))
	if strings.Contains(html, "Just a moment") {
		fmt.Println("** CLOUDFLARE BLOCKED EVEN WITH CHROMEDP **")
	} else {
		fmt.Println("** CHROMEDP SUCCESS **")
		doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		for _, sel := range []string{".upcoming-match", ".match-box", ".matchCard", ".liveMatch", "a[href*='/matches/']"} {
			if c := doc2.Find(sel).Length(); c > 0 {
				fmt.Printf("  %-30s: %d\n", sel, c)
			}
		}
	}
}

func trunc2(s string, n int) string {
	if len(s) > n { return s[:n] + "..." }
	return s
}
