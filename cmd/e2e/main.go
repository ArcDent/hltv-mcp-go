package main

import (
	"fmt"
	"time"

	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/facade"
	"github.com/arcdent/hltv-mcp/internal/types"
)

func main() {
	cfg := &config.Config{
		MCPServerName: "hltv-test", MCPServerVersion: "1.0.0",
		HTTPTimeoutMs: 15000, RetryCount: 1,
		DataSource: config.DataSourceAuto, ChromePath: "",
		CacheMaxEntries: 500, CacheStaleWindowSec: 3600,
		CacheTTLMatches: 60, CacheTTLResults: 120,
		CacheTTLEntity: 3600, CacheTTLTeam: 300, CacheTTLPlayer: 300,
		CacheTTLNews: 180, CacheTTLRealtimeNews: 60,
		DefaultResultLimit: 5, SummaryMode: config.SummaryTemplate,
		Timezone: "Asia/Shanghai",
	}
	_, chromeOK := client.CheckChromeAvailable(cfg)
	cli := client.NewHltvClient(cfg, chromeOK)
	f := facade.New(cfg, cache.New(cfg.CacheMaxEntries, cfg.CacheStaleWindowSec), cli)
	start := time.Now()

	pass, fail := 0, 0
	check := func(name string, ok bool) {
		if ok { pass++; fmt.Printf("  PASS %s\n", name) } else { fail++; fmt.Printf("  FAIL %s\n", name) }
	}

	// 1. Upcoming matches (chromedp)
	fmt.Println("\n1. Upcoming Matches:")
	resp := f.GetUpcomingMatches(types.UpcomingMatchesQuery{Limit: 5})
	ok := resp.Error == nil
	if ok {
		items, _ := resp.Items.([]types.NormalizedMatch)
		ok = len(items) > 0 && items[0].Team1 != "" && items[0].Team2 != ""
		if ok { fmt.Printf("   %s vs %s | %s\n", items[0].Team1, items[0].Team2, items[0].ScheduledAt) }
	}
	check("chromedp bypass + parse", ok)

	// 2. Team search with ID
	fmt.Println("\n2. Resolve Team:")
	resp2 := f.ResolveTeam(types.ResolveQuery{Name: "Vitality", Limit: 3})
	ok = resp2.Error == nil
	if ok {
		teams, _ := resp2.Items.([]types.ResolvedTeam)
		ok = len(teams) > 0 && teams[0].ID > 0
		if ok { fmt.Printf("   %s (id=%d, slug=%s)\n", teams[0].Name, teams[0].ID, teams[0].Slug) }
	}
	check("team ID parsed from link", ok)

	// 3. Player search with ID
	fmt.Println("\n3. Resolve Player:")
	resp3 := f.ResolvePlayer(types.ResolveQuery{Name: "ZywOo", Limit: 3})
	ok = resp3.Error == nil
	if ok {
		players, _ := resp3.Items.([]types.ResolvedPlayer)
		ok = len(players) > 0 && players[0].ID > 0
		if ok { fmt.Printf("   %s (id=%d, slug=%s)\n", players[0].Name, players[0].ID, players[0].Slug) }
	}
	check("player ID parsed from link", ok)

	// 4. Results
	fmt.Println("\n4. Recent Results:")
	resp4 := f.GetResultsRecent(types.ResultsRecentQuery{Limit: 5})
	ok = resp4.Error == nil
	if ok {
		items, _ := resp4.Items.([]types.NormalizedMatch)
		ok = len(items) > 0 && items[0].Score != ""
		if ok { fmt.Printf("   %s vs %s | %s\n", items[0].Team1, items[0].Team2, items[0].Score) }
	}
	check("score parsing", ok)

	// 5. Archive news
	fmt.Println("\n5. Archive News:")
	resp5 := f.GetNewsDigest(types.NewsDigestQuery{Year: 2026, Month: "May", Limit: 5})
	ok = resp5.Error == nil
	if ok {
		items, _ := resp5.Items.([]types.NewsItem)
		ok = len(items) > 0 && items[0].Title != ""
		if ok { fmt.Printf("   %s (%s)\n", items[0].Title, items[0].PublishedAt) }
	}
	check("news archive parsing", ok)

	// 6. Realtime news
	fmt.Println("\n6. Realtime News:")
	resp6 := f.GetRealtimeNews(types.RealtimeNewsQuery{Limit: 5})
	ok = resp6.Error == nil
	if ok {
		items, _ := resp6.Items.([]types.RealtimeNewsItem)
		ok = len(items) > 0 && !containsNewlines(items[0].Title)
		if ok { fmt.Printf("   %s\n", trunc(items[0].Title, 60)) }
	}
	check("news title clean", ok)

	fmt.Printf("\n=== RESULTS: %d/%d PASS | latency=%v ===\n", pass, pass+fail, time.Since(start))
}

func containsNewlines(s string) bool {
	for _, c := range s { if c == '\n' { return true } }
	return false
}
func trunc(s string, n int) string { if len(s) > n { return s[:n] + "..." }; return s }
