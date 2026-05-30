package types

// ResolvedTeam represents a team search result from HLTV
type ResolvedTeam struct {
	Type    string   `json:"type"`
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Country string   `json:"country,omitempty"`
	Rank    int      `json:"rank,omitempty"`
}

// ResolvedPlayer represents a player search result from HLTV
type ResolvedPlayer struct {
	Type    string `json:"type"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Team    string `json:"team,omitempty"`
	Country string `json:"country,omitempty"`
}

// TeamProfile represents a team's detail profile
type TeamProfile struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Country    string `json:"country,omitempty"`
	Rank       int    `json:"rank,omitempty"`
}

// PlayerProfile represents a player's detail profile
type PlayerProfile struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Team       string `json:"team,omitempty"`
	Country string `json:"country,omitempty"`
}

// MatchOutcome is the result of a match
type MatchOutcome string

const (
	OutcomeWin       MatchOutcome = "win"
	OutcomeLoss      MatchOutcome = "loss"
	OutcomeDraw      MatchOutcome = "draw"
	OutcomeScheduled MatchOutcome = "scheduled"
	OutcomeUnknown   MatchOutcome = "unknown"
)

// NormalizedMatch is a standardized match record
type NormalizedMatch struct {
	MatchID     int          `json:"match_id,omitempty"`
	Team1ID     int          `json:"team1_id,omitempty"`
	Team2ID     int          `json:"team2_id,omitempty"`
	OpponentID  int          `json:"opponent_id,omitempty"`
	Team1       string       `json:"team1,omitempty"`
	Team2       string       `json:"team2,omitempty"`
	Opponent    string       `json:"opponent,omitempty"`
	Event       string       `json:"event,omitempty"`
	Result      MatchOutcome `json:"result,omitempty"`
	Score       string       `json:"score,omitempty"`
	Winner      string       `json:"winner,omitempty"`
	BestOf      string       `json:"best_of,omitempty"`
	PlayedAt    string       `json:"played_at,omitempty"`
	ScheduledAt string       `json:"scheduled_at,omitempty"`
	MapText     string       `json:"map_text,omitempty"`
}

// NewsItem is an archive news entry
type NewsItem struct {
	Title       string `json:"title"`
	TitleZh     string `json:"title_zh,omitempty"`
	Link        string `json:"link,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	Tag         string `json:"tag,omitempty"`
}

// RealtimeNewsItem is a realtime news entry
type RealtimeNewsItem struct {
	Section      string `json:"section"`
	Category     string `json:"category,omitempty"`
	Title        string `json:"title"`
	TitleZh      string `json:"title_zh,omitempty"`
	RelativeTime string `json:"relative_time,omitempty"`
	Comments     string `json:"comments,omitempty"`
	Link         string `json:"link,omitempty"`
}

// TeamRecentData contains a team's recent results, upcoming matches, and stats
type TeamRecentData struct {
	Profile         TeamProfile       `json:"profile"`
	RecentResults   []NormalizedMatch `json:"recent_results"`
	UpcomingMatches []NormalizedMatch `json:"upcoming_matches"`
	SummaryStats    TeamSummaryStats  `json:"summary_stats"`
}

// TeamSummaryStats holds win/loss/draw counts
type TeamSummaryStats struct {
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Draws        int    `json:"draws"`
	RecentRecord string `json:"recent_record"`
}

// PlayerRecentData contains a player's profile, stats, and highlights
type PlayerRecentData struct {
	Profile          PlayerProfile     `json:"profile"`
	Overview         map[string]any    `json:"overview"`
	RecentMatches    []NormalizedMatch `json:"recent_matches"`
	RecentHighlights []string          `json:"recent_highlights"`
}

// PaginationMeta carries pagination metadata
type PaginationMeta struct {
	Offset      int  `json:"offset"`
	Limit       int  `json:"limit"`
	Returned    int  `json:"returned"`
	Total       int  `json:"total"`
	HasMore     bool `json:"has_more"`
	CurrentPage int  `json:"current_page"`
	NextOffset  *int `json:"next_offset,omitempty"`
	NextPage    *int `json:"next_page,omitempty"`
}

// ToolMeta carries metadata about a tool response
type ToolMeta struct {
	Source        string          `json:"source"`
	FetchedAt     string          `json:"fetched_at"`
	Timezone      string          `json:"timezone"`
	CacheHit      bool            `json:"cache_hit"`
	TTLSec        int             `json:"ttl_sec"`
	SchemaVersion string          `json:"schema_version"`
	Partial       bool            `json:"partial"`
	Notes         []string        `json:"notes,omitempty"`
	Stale         bool            `json:"stale,omitempty"`
	StaleAgeSec   int             `json:"stale_age_sec,omitempty"`
	Pagination    *PaginationMeta `json:"pagination,omitempty"`
}

// ToolError carries error information from a tool call
type ToolError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

func (e *ToolError) Error() string { return e.Message }

// ToolResponse is the unified response type for all MCP tools and REST API
type ToolResponse struct {
	Query          map[string]any `json:"query"`
	ResolvedEntity any            `json:"resolved_entity,omitempty"`
	Data           any            `json:"data,omitempty"`
	Items          any            `json:"items,omitempty"`
	Meta           ToolMeta       `json:"meta"`
	Error          *ToolError     `json:"error"`
}

// --- Query types ---

// ResolveQuery is used by resolve_team and resolve_player
type ResolveQuery struct {
	Name  string `json:"name"`
	Exact bool   `json:"exact,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// TeamRecentQuery is used by hltv_team_recent
type TeamRecentQuery struct {
	TeamID   int    `json:"team_id,omitempty"`
	TeamName string `json:"team_name,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// PlayerRecentQuery is used by hltv_player_recent
type PlayerRecentQuery struct {
	PlayerID   int    `json:"player_id,omitempty"`
	PlayerName string `json:"player_name,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// ResultsRecentQuery is used by hltv_results_recent
type ResultsRecentQuery struct {
	Team  string `json:"team,omitempty"`
	Event string `json:"event,omitempty"`
	Limit int    `json:"limit,omitempty"`
	Days  int    `json:"days,omitempty"`
}

// UpcomingMatchesQuery is used by hltv_matches_upcoming
type UpcomingMatchesQuery struct {
	TeamID    int    `json:"team_id,omitempty"`
	Team      string `json:"team,omitempty"`
	Event     string `json:"event,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Days      int    `json:"days,omitempty"`
	TodayOnly bool   `json:"today_only,omitempty"`
}

// NewsDigestQuery is used by hltv_news_digest
type NewsDigestQuery struct {
	Limit  int    `json:"limit,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Year   int    `json:"year,omitempty"`
	Month  string `json:"month,omitempty"`
	Page   int    `json:"page,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// RealtimeNewsQuery is used by hltv_realtime_news
type RealtimeNewsQuery struct {
	Limit  int `json:"limit,omitempty"`
	Page   int `json:"page,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// PlayerDetail is the full player profile scraped from HLTV player page
type PlayerDetail struct {
	Profile       PlayerDetailProfile `json:"profile"`
	Rating        PlayerRating        `json:"rating"`
	Abilities     []PlayerAbility     `json:"abilities"`
	Career        PlayerCareer        `json:"career"`
	Summary       *PlayerSummary      `json:"summary,omitempty"`
	Top20Ranks    map[string]int      `json:"top20_ranks"`
	Honors        []PlayerHonor       `json:"honors,omitempty"`
	RecentMatches []PlayerRecentMatch `json:"recent_matches,omitempty"`
}
type PlayerDetailProfile struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	RealName   string `json:"real_name,omitempty"`
	Slug       string `json:"slug"`
	Country    string `json:"country"`
	Age        int    `json:"age,omitempty"`
	Team       string `json:"team"`
}
type PlayerRating struct {
	Value float64 `json:"value"`
	Maps  int     `json:"maps"`
}
type PlayerAbility struct {
	Key     string  `json:"key"`
	LabelEn string  `json:"label_en"`
	LabelZh string  `json:"label_zh"`
	Value   float64 `json:"value"`
	Max     int     `json:"max"`
	Format  string  `json:"format,omitempty"`
}
type PlayerCareer struct {
	Matches     int     `json:"matches"`
	WinRate     string  `json:"win_rate"`
	KD          float64 `json:"kd"`
	HeadshotPct string  `json:"headshot_pct"`
	WinStreak   int     `json:"win_streak"`
}
type PlayerSummary struct {
	Teams           int `json:"teams"`
	DaysInTeam      int `json:"days_in_team"`
	DaysInTeams     int `json:"days_in_teams"`
	MajorWon        int `json:"major_won"`
	MajorsPlayed    int `json:"majors_played"`
	LANsWon         int `json:"lans_won"`
	LANsPlayed      int `json:"lans_played"`
	MajorTrophies   int `json:"major_trophies"`
	NotableTrophies int `json:"notable_trophies"`
	MajorMVPs       int `json:"major_mvps"`
	TotalMVPs       int `json:"total_mvps"`
	MajorEVPs       int `json:"major_evps"`
	TotalEVPs       int `json:"total_evps"`
}
type PlayerHonor struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}
type PlayerRecentMatch struct {
	MatchID   int     `json:"match_id"`
	MatchSlug string  `json:"-"`
	Date      string  `json:"date"`
	Team      string  `json:"team"`
	Opponent  string  `json:"opponent"`
	Score     string  `json:"score"`
	Result    string  `json:"result"`
	Rating    float64 `json:"rating"`
	Kills     int     `json:"kills"`
	Deaths    int     `json:"deaths"`
	Event     string  `json:"event"`
}

// TeamDetail is the full team profile scraped from HLTV team page
type TeamHighlightMatch struct {
	Opponent string `json:"opponent"`
	Result   string `json:"result"` // "won" or "lost"
}

type TeamHighlights struct {
	WinRate       string               `json:"win_rate"`
	WinStreak     int                  `json:"win_streak"`
	RecentMatches []TeamHighlightMatch `json:"recent_matches"`
}

type TeamDetail struct {
	Profile       TeamDetailProfile  `json:"profile"`
	Ranking       TeamRanking        `json:"ranking"`
	Stats         TeamStats          `json:"stats"`
	Achievements  []TeamAchievement  `json:"achievements"`
	Roster        []TeamRosterPlayer `json:"roster"`
	RecentMatches []NormalizedMatch  `json:"recent_matches"`
	Highlights    *TeamHighlights    `json:"highlights,omitempty"`
}

// TeamDetailProfile holds basic team identity
type TeamDetailProfile struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Country string `json:"country"`
}

// TeamRanking holds world ranking and points
type TeamRanking struct {
	WorldRank int `json:"world_rank"`
	Points    int `json:"points"`
}

// TeamStats holds win/loss/draw and form
type TeamStats struct {
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	WinRate    string `json:"win_rate"`
}

// TeamAchievement represents a trophy or record
type TeamAchievement struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Tier  string `json:"tier,omitempty"` // "major", "s", "a", "streak"
}

// TeamRosterPlayer is a player in the team roster
type TeamRosterPlayer struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	Rating  float64 `json:"rating"`
}

// NewsArticle is the full text of a news article scraped from HLTV
type NewsArticle struct {
	Title       string `json:"title"`
	TitleZh     string `json:"title_zh,omitempty"`
	PublishedAt string `json:"published_at"`
	Link        string `json:"link"`
	BodyText    string `json:"body_text"`
	BodyTextZh  string `json:"body_text_zh,omitempty"`
	Author      string `json:"author,omitempty"`
}
