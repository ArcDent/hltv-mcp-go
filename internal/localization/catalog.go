package localization

import "strings"

type TeamEntry struct {
	Canonical, Display, Official, Colloquial string
	Aliases                                   []string
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

