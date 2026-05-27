package facade

import (
	"encoding/json"
	"time"

	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/errors"
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
