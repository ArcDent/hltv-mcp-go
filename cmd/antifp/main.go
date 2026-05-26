package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/arcdent/hltv-mcp/internal/types"
)

func main() {
	userDir, _ := os.MkdirTemp("", "chrome-profile-*")
	defer os.RemoveAll(userDir)

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true), chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "TranslateUI,BlinkGenPropertyTrees"),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"),
		chromedp.UserDataDir(userDir),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.hltv.org/matches"),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil { fmt.Println("ERROR:", err); return }

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	// Find the first few match rows to understand HTML structure
	// Look for parent containers that contain /matches/ links
	matchLinkRE := regexp.MustCompile(`/matches/(\d+)/`)

	// Find all elements containing matches/ links and show their containers
	seenIDs := make(map[int]bool)
	count := 0
	doc.Find("a[href*='/matches/']").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		m := matchLinkRE.FindStringSubmatch(href)
		if len(m) < 2 { return }
		id, _ := strconv.Atoi(m[1])
		if seenIDs[id] || count >= 3 { return }
		seenIDs[id] = true
		count++

		fmt.Printf("\n=== Match %d (href=%s) ===\n", id, href)
		fmt.Printf("  link text: %q\n", cleanText(s.Text()))

		// Get class of the link itself
		cls, _ := s.Attr("class")
		fmt.Printf("  link class: %q\n", cls)

		// Show parent element info (up to 4 levels)
		for depth := 1; depth <= 5; depth++ {
			parent := s.Parent()
			if parent.Length() == 0 { break }
			tag := goquery.NodeName(parent)
			cls, _ := parent.Attr("class")
			if cls == "" {
				// Check for styled-components class
				if cssClass, ok := parent.Attr("class"); ok {
					cls = cssClass
				}
			}
			// Show all class-like attributes
			var allClasses []string
			for _, attr := range []string{"class"} {
				if v, ok := parent.Attr(attr); ok {
					allClasses = append(allClasses, v)
				}
			}
			fmt.Printf("  parent L%d: <%s class=%q>  (siblings: %d)\n",
				depth, tag, strings.Join(allClasses, " "),
				parent.Siblings().Length())
			// On first level, show all siblings
			if depth == 1 {
				parent.Children().Each(func(i int, child *goquery.Selection) {
					if i >= 6 { return }
					ctag := goquery.NodeName(child)
					ccls, _ := child.Attr("class")
					text := cleanText(child.Text())
					if len(text) > 40 { text = text[:40] + "..." }
					fmt.Printf("    sibling[%d]: <%s class=%q> %q\n", i, ctag, ccls, text)
				})
			}
		}
	})

	// Now try a better parsing approach
	fmt.Println("\n\n=== ATTEMPT 2: Parse from link hrefs + parent text ===")

	type matchData struct {
		id       int
		team1    string
		team2    string
		event    string
		time     string
		bestOf   string
	}
	matchMap := make(map[int]*matchData)

	doc.Find("a[href*='/matches/']").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		m := matchLinkRE.FindStringSubmatch(href)
		if len(m) < 2 { return }
		id, _ := strconv.Atoi(m[1])
		text := cleanText(s.Text())
		cls, _ := s.Attr("class")

		if _, exists := matchMap[id]; !exists {
			// Parse teams from URL
			slug := strings.TrimPrefix(href, "/matches/"+strconv.Itoa(id)+"/")
			if idx := strings.Index(slug, "-vs-"); idx >= 0 {
				t1 := slugToName(slug[:idx])
				t2 := extractTeam2FromSlug(slug[idx+4:])
				matchMap[id] = &matchData{id: id, team1: t1, team2: t2}
			} else {
				matchMap[id] = &matchData{id: id}
			}
		}
		md := matchMap[id]

		// Classify the link by text content
		if strings.Contains(text, ":") && len(text) < 10 && !strings.Contains(cls, "team") {
			md.time = text // "09:00"
		} else if strings.Contains(cls, "team") || isTeamName(text) {
			// team link
			if md.team1 == "" || text == md.team1 {
				md.team1 = text
			} else if md.team2 == "" || text == md.team2 {
				md.team2 = text
			}
		} else if len(text) > 10 && !strings.Contains(text, ":") {
			md.event = text // long text → event name
		}
		if strings.Contains(cls, "bo") || strings.Contains(text, "bo") {
			md.bestOf = text
		}
	})

	var matches []types.NormalizedMatch
	for _, md := range matchMap {
		if md.team1 == "" && md.team2 == "" { continue }
		if md.team1 == "" || md.team2 == "" { continue }
		matches = append(matches, types.NormalizedMatch{
			MatchID: md.id, Team1: md.team1, Team2: md.team2,
			Event: md.event, ScheduledAt: md.time, BestOf: md.bestOf,
			Result: types.OutcomeScheduled,
		})
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].MatchID < matches[j].MatchID })

	fmt.Printf("Valid matches: %d\n", len(matches))
	for i, m := range matches {
		if i >= 10 { break }
		fmt.Printf("  [%d] %s vs %s | event=%s | time=%s\n",
			i, m.Team1, m.Team2, trunc(m.Event, 30), m.ScheduledAt)
	}
}

func cleanText(s string) string { return strings.TrimSpace(s) }
func slugToName(s string) string { return strings.Title(strings.ReplaceAll(s, "-", " ")) }
func isTeamName(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 0 && len(s) < 20 && !strings.Contains(s, ":") &&
		!strings.Contains(s, "vs") && !strings.Contains(s, "-")
}
func extractTeam2FromSlug(slug string) string {
	eventKeywords := []string{"esl", "blast", "iem", "cct", "pgl", "yalla", "dust2", "dfrag",
		"nodwin", "lorgar", "tipsport", "elisa", "betboom", "stake", "winline",
		"european", "north-america", "south-america", "asia-pacific",
		"challenger", "league", "season", "series", "closed", "open",
		"qualifier", "cup", "group", "stage", "division", "lan", "online",
		"showdown", "masters", "major", "conquest", "storm", "finals"}
	for _, kw := range eventKeywords {
		if idx := strings.Index(strings.ToLower(slug), kw); idx > 0 {
			slug = strings.TrimRight(slug[:idx], "-")
			break
		}
	}
	if idx := strings.LastIndex(slug, "-"); idx > 0 {
		if _, err := strconv.Atoi(slug[idx+1:]); err == nil {
			slug = slug[:idx]
		}
	}
	return strings.ReplaceAll(slug, "-", " ")
}
func trunc(s string, n int) string { if len(s) > n { return s[:n] + "..." }; return s }
