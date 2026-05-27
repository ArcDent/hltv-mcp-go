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

func TestTranslatePlaceholder(t *testing.T) {
	tests := []struct{ in, want string }{
		{"winner", "胜者"},
		{"Winner", "胜者"},
		{"Winner of Group A", "胜者"},
		{"WINNER", "胜者"},
		{"loser", "败者"},
		{"Loser of match 3", "败者"},
		{"tbd", "待定"},
		{"TBD", "待定"},
		{"  tbd  ", "待定"},
		{"Vitality", "Vitality"},
		{"FaZe Clan", "FaZe Clan"},
	}
	for _, tt := range tests {
		if got := TranslatePlaceholder(tt.in); got != tt.want {
			t.Errorf("TranslatePlaceholder(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeBO1Score(t *testing.T) {
	tests := []struct{ in, want string }{
		{"2:0", "2:0"},
		{"2:1", "2:1"},
		{"0:2", "0:2"},
		{"13:5", "1:0"},
		{"5:13", "0:1"},
		{"16:14", "1:0"},
		{"14:16", "0:1"},
		{"16:16", "平局"},
		{"13:11", "1:0"},
		{"11:13", "0:1"},
		{"", ""},
		{"invalid", "invalid"},
		{"13 : 5", "1:0"},
		{"5 : 13", "0:1"},
	}
	for _, tt := range tests {
		if got := normalizeBO1Score(tt.in); got != tt.want {
			t.Errorf("normalizeBO1Score(%q) = %q, want %q", tt.in, got, tt.want)
		}
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
