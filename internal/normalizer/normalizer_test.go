package normalizer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

func TestNormalizeMatches(t *testing.T) {
	html := `<div class="result-con"><a class="a-reset" href="/matches/123/foo-vs-bar"><div class="result"><table><tbody><tr><td class="team-cell"><div class="line-align team1"><div class="team">Spirit</div></div></td><td class="result-score">2:1</td><td class="team-cell"><div class="line-align team2"><div class="team">Vitality</div></div></td></tr></tbody></table></div></a></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	matches := NormalizeMatches(doc, "")
	if len(matches) == 0 {
		t.Fatal("expected at least 1 match")
	}
	if matches[0].Team1 != "Spirit" {
		t.Errorf("team1: %s", matches[0].Team1)
	}
	if matches[0].Score != "2:1" {
		t.Errorf("score: %s", matches[0].Score)
	}
	if matches[0].Team2 != "Vitality" {
		t.Errorf("team2: %s", matches[0].Team2)
	}
}

func TestNormalizeNews(t *testing.T) {
	html := `<div class="newstext">Test Title</div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	items := NormalizeNews(doc)
	if len(items) == 0 {
		t.Fatal("expected news items")
	}
	if items[0].Title != "Test Title" {
		t.Errorf("title: %s", items[0].Title)
	}
}

func TestSplitTeamMatches(t *testing.T) {
	matches := []types.NormalizedMatch{
		{Score: "2:1", PlayedAt: "2025-01-01"},
		{ScheduledAt: "2025-02-01"},
		{Score: "1:2", PlayedAt: "2025-01-02"},
	}
	recent, upcoming := SplitTeamMatches(matches)
	if len(recent) != 2 {
		t.Errorf("expected 2 recent, got %d", len(recent))
	}
	if len(upcoming) != 1 {
		t.Errorf("expected 1 upcoming, got %d", len(upcoming))
	}
}
