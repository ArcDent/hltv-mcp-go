package localization

import "strings"

type TeamEntry struct {
	Canonical, Display, Official, Colloquial string
	Aliases                                   []string
}

// PlayerEntry represents a player nickname mapping.
type PlayerEntry struct {
	Canonical  string
	Colloquial string
}

var TeamCatalog = []TeamEntry{
	{Canonical: "Team Spirit", Display: "Spirit", Official: "", Colloquial: "绿龙", Aliases: []string{"Spirit", "Team Spirit", "绿龙"}},
	{Canonical: "Vitality", Display: "Vitality", Official: "", Colloquial: "小蜜蜂", Aliases: []string{"Vitality", "Team Vitality", "小蜜蜂", "蜜蜂"}},
	{Canonical: "Natus Vincere", Display: "Natus Vincere", Official: "", Colloquial: "NaVi", Aliases: []string{"Natus Vincere", "NaVi", "NAVI", "天生赢家"}},
	{Canonical: "G2", Display: "G2", Official: "", Colloquial: "G2", Aliases: []string{"G2", "G2 Esports", "武士"}},
	{Canonical: "MOUZ", Display: "MOUZ", Official: "", Colloquial: "老鼠", Aliases: []string{"MOUZ", "mouz", "老鼠"}},
	{Canonical: "FaZe", Display: "FaZe", Official: "", Colloquial: "FaZe", Aliases: []string{"FaZe", "FaZe Clan"}},
	{Canonical: "Falcons", Display: "Falcons", Official: "", Colloquial: "猎鹰", Aliases: []string{"Falcons", "Team Falcons", "猎鹰"}},
	{Canonical: "Astralis", Display: "Astralis", Official: "", Colloquial: "A队", Aliases: []string{"Astralis", "A队"}},
	{Canonical: "Virtus.pro", Display: "Virtus.pro", Official: "", Colloquial: "VP", Aliases: []string{"Virtus.pro", "Virtus Pro", "VP"}},
	{Canonical: "Team Liquid", Display: "Liquid", Official: "", Colloquial: "液体", Aliases: []string{"Team Liquid", "Liquid", "液体"}},
	{Canonical: "FURIA", Display: "FURIA", Official: "", Colloquial: "黑豹", Aliases: []string{"FURIA", "黑豹"}},
	{Canonical: "Aurora", Display: "Aurora", Official: "", Colloquial: "欧若拉", Aliases: []string{"Aurora", "欧若拉"}},
	{Canonical: "HEROIC", Display: "HEROIC", Official: "", Colloquial: "X队", Aliases: []string{"HEROIC", "X队"}},
	{Canonical: "PARIVISION", Display: "PARIVISION", Official: "", Colloquial: "PV", Aliases: []string{"PARIVISION", "PARI", "PV"}},
	{Canonical: "paiN", Display: "paiN", Official: "", Colloquial: "paiN", Aliases: []string{"paiN", "paiN Gaming"}},
	{Canonical: "Complexity", Display: "Complexity", Official: "", Colloquial: "COL", Aliases: []string{"Complexity", "Complexity Gaming", "coL", "COL"}},
	{Canonical: "Ninjas in Pyjamas", Display: "Ninjas in Pyjamas", Official: "", Colloquial: "NIP", Aliases: []string{"Ninjas in Pyjamas", "NiP", "NIP"}},
	{Canonical: "GamerLegion", Display: "GamerLegion", Official: "", Colloquial: "GL", Aliases: []string{"GamerLegion", "GL"}},
	{Canonical: "The MongolZ", Display: "The MongolZ", Official: "", Colloquial: "蒙古", Aliases: []string{"The MongolZ", "MongolZ", "蒙古队", "蒙古"}},
	{Canonical: "TYLOO", Display: "TYLOO", Official: "", Colloquial: "天禄", Aliases: []string{"TYLOO", "天禄"}},
	{Canonical: "Rare Atom", Display: "Rare Atom", Official: "", Colloquial: "RA", Aliases: []string{"Rare Atom", "RA"}},
	{Canonical: "Lynn Vision", Display: "Lynn Vision", Official: "", Colloquial: "LVG", Aliases: []string{"Lynn Vision", "LVG"}},
	{Canonical: "fnatic", Display: "fnatic", Official: "", Colloquial: "FNC", Aliases: []string{"fnatic", "Fnatic", "橙黑", "FNC"}},
	{Canonical: "Eternal Fire", Display: "Eternal Fire", Official: "", Colloquial: "EF", Aliases: []string{"Eternal Fire", "永火", "EF"}},
	{Canonical: "RED Canids", Display: "RED Canids", Official: "", Colloquial: "RED", Aliases: []string{"RED Canids", "红犬", "RED"}},
	{Canonical: "3DMAX", Display: "3DMAX", Official: "", Colloquial: "3DMAX", Aliases: []string{"3DMAX"}},
}

var PlayerCatalog = []PlayerEntry{
	{Canonical: "ZywOo", Colloquial: "载物"},
	{Canonical: "s1mple", Colloquial: "森破"},
	{Canonical: "m0NESY", Colloquial: "小孩"},
	{Canonical: "donk", Colloquial: "小驴"},
	{Canonical: "NiKo", Colloquial: "🦐"},
	{Canonical: "dev1ce", Colloquial: "老爸"},
	{Canonical: "ropz", Colloquial: "车主"},
	{Canonical: "karrigan", Colloquial: "大表哥"},
	{Canonical: "apEX", Colloquial: "豆豆"},
	{Canonical: "flameZ", Colloquial: "火仔/🗿"},
	{Canonical: "Spinx", Colloquial: "米人"},
	{Canonical: "mezii", Colloquial: "妹子"},
	{Canonical: "jL", Colloquial: "jL"},
	{Canonical: "Aleksib", Colloquial: "小李子"},
	{Canonical: "b1t", Colloquial: "b1t"},
	{Canonical: "iM", Colloquial: "iM"},
	{Canonical: "w0nderful", Colloquial: "wdf"},
	{Canonical: "broky", Colloquial: "箱子"},
	{Canonical: "frozen", Colloquial: "寒王"},
	{Canonical: "Twistzz", Colloquial: "总监"},
	{Canonical: "huNter-", Colloquial: "表哥"},
	{Canonical: "jks", Colloquial: "jks"},
	{Canonical: "NAF", Colloquial: "NAF"},
	{Canonical: "YEKINDAR", Colloquial: "狂哥"},
	{Canonical: "cadiaN", Colloquial: "点子哥"},
	{Canonical: "stavn", Colloquial: "蛇"},
	{Canonical: "jabbi", Colloquial: "jabbi"},
	{Canonical: "TeSeS", Colloquial: "龙哥"},
	{Canonical: "EliGE", Colloquial: "鸡哥"},
	{Canonical: "Magisk", Colloquial: "魔男"},
	{Canonical: "dupreeh", Colloquial: "阿杜"},
	{Canonical: "Xyp9x", Colloquial: "九爷"},
	{Canonical: "gla1ve", Colloquial: "队长"},
	{Canonical: "electroNic", Colloquial: "电子哥"},
	{Canonical: "Perfecto", Colloquial: "P皇"},
	{Canonical: "Boombl4", Colloquial: "胖球"},
	{Canonical: "sh1ro", Colloquial: "息若"},
	{Canonical: "Ax1Le", Colloquial: "Ax1Le"},
	{Canonical: "Hobbit", Colloquial: "霍比特"},
	{Canonical: "KSCERATO", Colloquial: "KSCERATO/卡神"},
	{Canonical: "yuurih", Colloquial: "尤里"},
	{Canonical: "arT", Colloquial: "艺术哥"},
	{Canonical: "FalleN", Colloquial: "教父"},
	{Canonical: "rain", Colloquial: "雨神"},
	{Canonical: "siuhy", Colloquial: "siuhy"},
	{Canonical: "xertioN", Colloquial: "xertioN"},
	{Canonical: "torzsi", Colloquial: "托子"},
	{Canonical: "Brollan", Colloquial: "宝蓝"},
	{Canonical: "magixx", Colloquial: "马西西"},
	{Canonical: "chopper", Colloquial: "大超"},
	{Canonical: "zont1x", Colloquial: "宗主"},
	{Canonical: "degster", Colloquial: "degster"},
	{Canonical: "Kyojin", Colloquial: "巨人"},
	{Canonical: "malbsMd", Colloquial: "malbsMd"},
	{Canonical: "HeavyGod", Colloquial: "重God/重神"},
	{Canonical: "ultimate", Colloquial: "ultimate"},
	{Canonical: "jottAA", Colloquial: "jottAA"},
	{Canonical: "NertZ", Colloquial: "NertZ"},
	{Canonical: "misutaaa", Colloquial: "米苏塔"},
	{Canonical: "blameF", Colloquial: "胖虎"},
	{Canonical: "FL1T", Colloquial: "FL1T"},
	{Canonical: "fame", Colloquial: "fame"},
	{Canonical: "n0rb3r7", Colloquial: "n0rb3r7"},
	{Canonical: "XANTARES", Colloquial: "狠人"},
	{Canonical: "woxic", Colloquial: "woxic"},
	{Canonical: "Staehr", Colloquial: "野榜"},
	{Canonical: "br0", Colloquial: "br0"},
	{Canonical: "chelo", Colloquial: "chelo"},
	{Canonical: "skullz", Colloquial: "skullz"},
	{Canonical: "biguzera", Colloquial: "biguzera"},
	{Canonical: "bLitz", Colloquial: "布里茨"},
	{Canonical: "910", Colloquial: "910"},
	{Canonical: "Techno4K", Colloquial: "Techno4K"},
	{Canonical: "Senzu", Colloquial: "森组"},
	{Canonical: "nqz", Colloquial: "nqz"},
	{Canonical: "sl3nd", Colloquial: "sl3nd"},
	{Canonical: "volt", Colloquial: "volt"},
	{Canonical: "SunPayus", Colloquial: "阳叔"},
	{Canonical: "Nilo", Colloquial: "Nilo"},
	{Canonical: "maden", Colloquial: "maden"},
	{Canonical: "Patsi", Colloquial: "Patsi"},
	{Canonical: "BELCHONOKK", Colloquial: "BELCHONOKK"},
	{Canonical: "hallzerk", Colloquial: "hallzerk"},
	{Canonical: "Grim", Colloquial: "Grim"},
	{Canonical: "r1nkle", Colloquial: "r1nkle"},
	{Canonical: "bodyy", Colloquial: "bodyy"},
	{Canonical: "MATYS", Colloquial: "MATYS"},
	{Canonical: "Maka", Colloquial: "Maka"},
	{Canonical: "Lucky", Colloquial: "好运or坏运"},
	{Canonical: "JamYoung", Colloquial: "小鞠"},
	{Canonical: "Mercury", Colloquial: "汉堡"},
	{Canonical: "Starry", Colloquial: "小舅子"},
	{Canonical: "somebody", Colloquial: "sbd"},
	{Canonical: "Lack1", Colloquial: "Lack1"},
	{Canonical: "tN1R", Colloquial: "少爷/特尼尔"},
}

var playerCatalogMap map[string]string

func init() {
	playerCatalogMap = make(map[string]string, len(PlayerCatalog))
	for _, e := range PlayerCatalog {
		playerCatalogMap[e.Canonical] = e.Colloquial
	}
}

var teamLookup = buildLookup(TeamCatalog)

func buildLookup(catalog []TeamEntry) map[string]*TeamEntry {
	m := make(map[string]*TeamEntry)
	for i := range catalog {
		e := &catalog[i]
		for _, a := range allVariants(e) {
			m[strings.ToLower(a)] = e
		}
	}
	return m
}

func allVariants(e *TeamEntry) []string {
	return dedup(append([]string{e.Canonical, e.Display, e.Official, e.Colloquial}, e.Aliases...))
}

func dedup(items []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range items {
		if s != "" && !seen[strings.ToLower(s)] {
			seen[strings.ToLower(s)] = true
			out = append(out, s)
		}
	}
	return out
}

func LookupTeam(name string) *TeamEntry {
	return teamLookup[strings.ToLower(strings.TrimSpace(name))]
}

func FormatTeamDisplay(name string) string {
	if e := LookupTeam(name); e != nil {
		parts := dedup([]string{e.Display, e.Official})
		if e.Colloquial != "" && e.Colloquial != e.Display {
			parts = append(parts, e.Colloquial)
		}
		return strings.Join(parts, "/")
	}
	return name
}

// PlayerNickname returns the colloquial name for a player.
// Checks overrides first, then falls back to the built-in catalog.
func PlayerNickname(name string) string {
	if n := GetPlayerOverride(name); n != "" {
		return n
	}
	return playerCatalogMap[name]
}

// TeamNickname returns the colloquial name for a team.
// Checks overrides first, then falls back to TeamCatalog.Colloquial.
func TeamNickname(name string) string {
	if n := GetTeamOverride(name); n != "" {
		return n
	}
	e := LookupTeam(name)
	if e == nil || e.Colloquial == "" {
		return ""
	}
	return e.Colloquial
}

// BuildFullDict returns the complete team+player nickname dictionaries
// with all variants expanded and user overrides applied on top.
func BuildFullDict() (teams, players map[string]string) {
	teams = make(map[string]string)
	for _, e := range TeamCatalog {
		if e.Colloquial == "" {
			continue
		}
		for _, v := range allVariants(&e) {
			if v != "" {
				teams[v] = e.Colloquial
			}
		}
		// Also add raw aliases to preserve case variants lost by dedup
		for _, a := range e.Aliases {
			if a != "" {
				teams[a] = e.Colloquial
			}
		}
	}
	// Apply team overrides — resolve canonical to all variants
	ov.mu.RLock()
	for canonicalName, nickname := range ov.teams {
		if nickname == "" {
			continue
		}
		if e := LookupTeam(canonicalName); e != nil {
			for _, v := range allVariants(e) {
				if v != "" {
					teams[v] = nickname
				}
			}
		}
	}
	ov.mu.RUnlock()

	players = make(map[string]string, len(playerCatalogMap))
	for k, v := range playerCatalogMap {
		players[k] = v
	}
	ov.mu.RLock()
	for k, v := range ov.players {
		if v != "" {
			players[k] = v
		}
	}
	ov.mu.RUnlock()

	return
}

