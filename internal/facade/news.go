package facade

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/types"
)

var genericNewsTags = map[string]bool{
	"news": true, "latest": true, "today": true,
	"新闻": true, "最新": true, "最新新闻": true, "今日": true, "今日新闻": true, "实时新闻": true,
}

func normalizeArchiveNewsTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" || genericNewsTags[strings.ToLower(tag)] {
		return ""
	}
	return tag
}

// GetRealtimeNews fetches realtime news from HLTV homepage
func (f *HltvFacade) GetRealtimeNews(query types.RealtimeNewsQuery) *types.ToolResponse {
	if query.Limit == 0 {
		query.Limit = 25
	}
	q := map[string]any{"limit": query.Limit, "offset": query.Offset}
	key := fmt.Sprintf("realtime_news:%d:%d", query.Limit, query.Offset)
	ttl := f.cfg.CacheTTLRealtimeNews

	return f.withCacheOrStore(key, ttl, q,
		func() (*types.ToolResponse, bool) {
			items, err := f.store.QueryRealtimeNews(query.Limit)
			if err != nil || len(items) == 0 {
				return nil, false
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, true
		},
		func() (*types.ToolResponse, error) {
			doc, err := f.rns.GetRealtimeNews(context.Background())
			if err != nil {
				return nil, err
			}
			allItems := normalizer.NormalizeRealtimeNews(doc)

			if f.store != nil {
				if err := f.store.BatchUpsertRealtimeNews(allItems); err != nil {
					log.Printf("facade: batch upsert realtime news: %v", err)
				}
			}

			start := query.Offset
			end := start + query.Limit
			if end > len(allItems) {
				end = len(allItems)
			}
			items := allItems[start:end]
			hasMore := end < len(allItems)
			pagination := &types.PaginationMeta{
				Offset: start, Limit: query.Limit, Returned: len(items),
				Total: len(allItems), HasMore: hasMore, CurrentPage: query.Page,
			}
			if hasMore {
				next := end
				pagination.NextOffset = &next
				np := query.Page + 1
				pagination.NextPage = &np
			}
			meta := f.createMeta(ttl)
			meta.Pagination = pagination
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
		})
}

// GetNewsDigest fetches monthly archive news with optional tag filter
func (f *HltvFacade) GetNewsDigest(query types.NewsDigestQuery) *types.ToolResponse {
	if query.Limit == 0 {
		query.Limit = 25
	}
	tag := normalizeArchiveNewsTag(query.Tag)
	q := map[string]any{"tag": tag, "year": query.Year, "month": query.Month}
	key := fmt.Sprintf("news_digest:%s:%d:%s", tag, query.Year, query.Month)
	ttl := f.cfg.CacheTTLNews

	return f.withCacheOrStore(key, ttl, q,
		func() (*types.ToolResponse, bool) {
			items, err := f.store.QueryNews(query.Limit)
			if err != nil || len(items) == 0 {
				return nil, false
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, true
		},
		func() (*types.ToolResponse, error) {
			doc, err := f.ns.GetNews(context.Background(), query.Year, query.Month)
			if err != nil {
				return nil, err
			}
			allItems := normalizer.NormalizeNews(doc)

			if f.store != nil {
				if err := f.store.BatchUpsertNews(allItems); err != nil {
					log.Printf("facade: batch upsert news: %v", err)
				}
			}

			var filtered []types.NewsItem
			for _, item := range allItems {
				if tag == "" ||
					strings.Contains(strings.ToLower(item.Title), strings.ToLower(tag)) ||
					strings.Contains(strings.ToLower(item.Tag), strings.ToLower(tag)) {
					filtered = append(filtered, item)
				}
			}
			start := query.Offset
			end := start + query.Limit
			if end > len(filtered) {
				end = len(filtered)
			}
			items := filtered[start:end]
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
		})
}
