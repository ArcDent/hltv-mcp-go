package facade

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/errors"
	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/scraper"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// HltvFacade orchestrates data fetching with caching, normalization, and filtering
type HltvFacade struct {
	cfg    *config.Config
	cache  *cache.Cache
	client *client.HltvClient
	ts     *scraper.TeamScraper
	ps     *scraper.PlayerScraper
	rs     *scraper.ResultsScraper
	ms     *scraper.MatchesScraper
	ns     *scraper.NewsScraper
	rns    *scraper.RealtimeNewsScraper
	nas    *scraper.NewsArticleScraper
}

// New creates a new HltvFacade
func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient) *HltvFacade {
	return &HltvFacade{
		cfg:    cfg,
		cache:  c,
		client: cli,
		ts:     scraper.NewTeamScraper(cli),
		ps:     scraper.NewPlayerScraper(cli),
		rs:     scraper.NewResultsScraper(cli),
		ms:     scraper.NewMatchesScraper(cli),
		ns:     scraper.NewNewsScraper(cli),
		rns:    scraper.NewRealtimeNewsScraper(cli),
		nas:    scraper.NewNewsArticleScraper(cli),
	}
}

func (f *HltvFacade) createMeta(ttlSec int) types.ToolMeta {
	return types.ToolMeta{
		Source:        "hltv-mcp",
		FetchedAt:     time.Now().UTC().Format(time.RFC3339),
		Timezone:      f.cfg.Timezone,
		TTLSec:        ttlSec,
		SchemaVersion: "1.0",
	}
}

// withCache checks cache, then computes and caches the result
func (f *HltvFacade) ClientIsChromeAvailable() bool { return f.client.IsChromeAvailable() }

func (f *HltvFacade) ScrapePlayerDetail(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	if slug == "" { slug = fmt.Sprintf("player-%d", id) }
	return f.ps.GetPlayer(ctx, id, slug)
}

func (f *HltvFacade) ScrapeTeamDetail(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	if slug == "" { slug = fmt.Sprintf("team-%d", id) }
	return f.ts.GetTeam(ctx, id, slug)
}

// GetPlayerDetailCached returns cached player detail, or scrapes via chromedp and caches for 7 days
func (f *HltvFacade) GetPlayerDetailCached(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("player-%d", id)
	}
	key := fmt.Sprintf("player_detail:%d", id)
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.PlayerDetail), nil
	}
	doc, err := f.ps.GetPlayer(ctx, id, slug)
	if err != nil {
		return types.PlayerDetail{}, err
	}
	pd := normalizer.NormalizePlayerDetail(doc)
	pd.Profile.ID = id
	f.cache.Set(key, pd, f.cfg.CacheTTLPlayerDetail)
	return pd, nil
}

// ScrapeNewsArticle fetches a news article page via chromedp
func (f *HltvFacade) ScrapeNewsArticle(ctx context.Context, url string) (*goquery.Document, error) {
	return f.nas.GetArticle(ctx, url)
}

// GetNewsArticleCached returns cached article body, or scrapes and caches indefinitely
func (f *HltvFacade) GetNewsArticleCached(ctx context.Context, url string) (types.NewsArticle, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	key := fmt.Sprintf("news_article:%s", hash)
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.NewsArticle), nil
	}
	doc, err := f.nas.GetArticle(ctx, url)
	if err != nil {
		return types.NewsArticle{}, err
	}
	article := normalizer.NormalizeNewsArticle(doc, url)
	f.cache.Set(key, article, f.cfg.CacheTTLNewsArticle)
	return article, nil
}

// GetTeamDetailCached returns cached team detail, or scrapes via chromedp and caches for 7 days
func (f *HltvFacade) GetTeamDetailCached(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("team-%d", id)
	}
	key := fmt.Sprintf("team_detail:%d", id)
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.TeamDetail), nil
	}
	doc, err := f.ts.GetTeam(ctx, id, slug)
	if err != nil {
		return types.TeamDetail{}, err
	}
	td := normalizer.NormalizeTeamDetail(doc)
	td.Profile.ID = id
	td.Profile.Slug = slug
		// Fetch recent matches via standard results/matches pages, filter by team name
		var allMatches []types.NormalizedMatch
		if td.Profile.Name != "" {
			if resultDoc, err := f.rs.GetResults(ctx); err == nil {
				allResults := normalizer.NormalizeMatches(resultDoc, td.Profile.Name)
				allMatches = append(allMatches, allResults...)
			}
			if upcomingDoc, err := f.ms.GetUpcoming(ctx); err == nil {
				allUpcoming := normalizer.NormalizeUpcomingMatches(upcomingDoc, td.Profile.Name)
				allMatches = append(allMatches, allUpcoming...)
			}
			name := td.Profile.Name
			var matches []types.NormalizedMatch
			for _, m := range allMatches {
				if m.Team1 == name || m.Team2 == name || m.Opponent == name {
					matches = append(matches, m)
				}
			}
			normalizer.SortByPlayedAtDesc(matches)
			if len(matches) > 10 {
				matches = matches[:10]
			}
			td.RecentMatches = matches
			// Compute stats from recent matches
			for _, m := range matches {
				switch m.Result {
				case types.OutcomeWin:
					td.Stats.Wins++
				case types.OutcomeLoss:
					td.Stats.Losses++
				case types.OutcomeDraw:
					td.Stats.Draws++
				}
			}
			total := td.Stats.Wins + td.Stats.Losses + td.Stats.Draws
			if total > 0 {
				td.Stats.WinRate = fmt.Sprintf("%.0f%%", float64(td.Stats.Wins)/float64(total)*100)
			}
			// Recent form string (last 5)
			for i, m := range matches {
				if i >= 5 {
					break
				}
				switch m.Result {
				case types.OutcomeWin:
					td.Stats.RecentForm += "W"
				case types.OutcomeLoss:
					td.Stats.RecentForm += "L"
				case types.OutcomeDraw:
					td.Stats.RecentForm += "D"
				}
			}
		}
	f.cache.Set(key, td, f.cfg.CacheTTLPlayerDetail) // reuse 7d TTL
	return td, nil
}

func (f *HltvFacade) withCache(key string, ttlSec int, query map[string]any, compute func() (*types.ToolResponse, error)) *types.ToolResponse {
	if cached, ok := f.cache.Get(key); ok {
		r := cloneResponse(cached.(*types.ToolResponse))
		r.Meta.CacheHit = true
		return r
	}
	if stale, sm, ok := f.cache.GetStale(key); ok {
		r := cloneResponse(stale.(*types.ToolResponse))
		r.Meta.CacheHit = true
		r.Meta.Stale = true
		r.Meta.StaleAgeSec = sm.StaleAgeSec
		return r
	}
	val, err := f.cache.RunOnce(key, func() (any, error) {
		r, computeErr := compute()
		if computeErr != nil {
			return nil, computeErr
		}
		f.cache.Set(key, r, ttlSec)
		return r, nil
	})
	if err != nil {
		return f.errorResponse(query, err)
	}
	return val.(*types.ToolResponse)
}

func cloneResponse(r *types.ToolResponse) *types.ToolResponse {
	data, _ := json.Marshal(r)
	var c types.ToolResponse
	json.Unmarshal(data, &c)
	return &c
}

func (f *HltvFacade) errorResponse(query map[string]any, err error) *types.ToolResponse {
	meta := f.createMeta(60)
	if appErr, ok := err.(*errors.AppError); ok {
		return &types.ToolResponse{
			Query: query,
			Meta:  meta,
			Error: &types.ToolError{
				Code:      string(appErr.Code),
				Message:   appErr.Message,
				Retryable: appErr.Retryable,
				Details:   appErr.Details,
			},
		}
	}
	return &types.ToolResponse{
		Query: query,
		Meta:  meta,
		Error: &types.ToolError{
			Code:      "INTERNAL_ERROR",
			Message:   err.Error(),
			Retryable: false,
		},
	}
}

// CacheEntries returns the number of entries currently in the cache
func (f *HltvFacade) CacheEntries() int { return f.cache.Entries() }

// CacheHits returns the cumulative cache hit count
func (f *HltvFacade) CacheHits() int64 { return f.cache.Hits() }

// CacheMisses returns the cumulative cache miss count
func (f *HltvFacade) CacheMisses() int64 { return f.cache.Misses() }

// ClearCache clears all cached entries and resets counters
func (f *HltvFacade) ClearCache() { f.cache.Clear() }
