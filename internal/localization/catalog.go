package localization

import "strings"

type TeamEntry struct {
	Canonical, Display, Official, Colloquial string
	Aliases                                   []string
}

var TeamCatalog = []TeamEntry{
	{Canonical: "Team Spirit", Display: "Spirit", Official: "Spirit战队", Colloquial: "绿龙", Aliases: []string{"Spirit", "Team Spirit", "绿龙"}},
	{Canonical: "Vitality", Display: "Vitality", Official: "Vitality战队", Colloquial: "小蜜蜂", Aliases: []string{"Vitality", "Team Vitality", "小蜜蜂", "蜜蜂"}},
	{Canonical: "Natus Vincere", Display: "Natus Vincere", Official: "Natus Vincere战队", Colloquial: "NaVi", Aliases: []string{"Natus Vincere", "NaVi", "NAVI", "天生赢家"}},
	{Canonical: "G2", Display: "G2", Official: "G2战队", Colloquial: "武士", Aliases: []string{"G2", "G2 Esports", "武士"}},
	{Canonical: "MOUZ", Display: "MOUZ", Official: "MOUZ战队", Colloquial: "老鼠", Aliases: []string{"MOUZ", "mouz", "老鼠"}},
	{Canonical: "FaZe", Display: "FaZe", Official: "FaZe战队", Colloquial: "FaZe", Aliases: []string{"FaZe", "FaZe Clan"}},
	{Canonical: "Falcons", Display: "Falcons", Official: "Falcons战队", Colloquial: "猎鹰", Aliases: []string{"Falcons", "Team Falcons", "猎鹰"}},
	{Canonical: "Astralis", Display: "Astralis", Official: "Astralis战队", Colloquial: "A队", Aliases: []string{"Astralis", "A队"}},
	{Canonical: "Virtus.pro", Display: "Virtus.pro", Official: "Virtus.pro战队", Colloquial: "VP", Aliases: []string{"Virtus.pro", "Virtus Pro", "VP"}},
	{Canonical: "Team Liquid", Display: "Liquid", Official: "Liquid战队", Colloquial: "液体", Aliases: []string{"Team Liquid", "Liquid", "液体"}},
	{Canonical: "FURIA", Display: "FURIA", Official: "FURIA战队", Colloquial: "黑豹", Aliases: []string{"FURIA", "黑豹"}},
	{Canonical: "Aurora", Display: "Aurora", Official: "Aurora战队", Colloquial: "欧若拉", Aliases: []string{"Aurora", "欧若拉"}},
	{Canonical: "HEROIC", Display: "HEROIC", Official: "HEROIC战队", Colloquial: "HEROIC", Aliases: []string{"HEROIC"}},
	{Canonical: "PARIVISION", Display: "PARIVISION", Official: "PARIVISION战队", Colloquial: "PV", Aliases: []string{"PARIVISION", "PARI", "PV"}},
	{Canonical: "paiN", Display: "paiN", Official: "paiN Gaming战队", Colloquial: "paiN", Aliases: []string{"paiN", "paiN Gaming"}},
	{Canonical: "Complexity", Display: "Complexity", Official: "Complexity战队", Colloquial: "coL", Aliases: []string{"Complexity", "Complexity Gaming", "coL"}},
	{Canonical: "Ninjas in Pyjamas", Display: "Ninjas in Pyjamas", Official: "Ninjas in Pyjamas战队", Colloquial: "NIP", Aliases: []string{"Ninjas in Pyjamas", "NiP", "NIP"}},
	{Canonical: "GamerLegion", Display: "GamerLegion", Official: "GamerLegion战队", Colloquial: "GL", Aliases: []string{"GamerLegion", "GL"}},
	{Canonical: "The MongolZ", Display: "The MongolZ", Official: "The MongolZ战队", Colloquial: "蒙古队", Aliases: []string{"The MongolZ", "MongolZ", "蒙古队"}},
	{Canonical: "TYLOO", Display: "TYLOO", Official: "TYLOO战队", Colloquial: "天禄", Aliases: []string{"TYLOO", "天禄"}},
	{Canonical: "Rare Atom", Display: "Rare Atom", Official: "Rare Atom战队", Colloquial: "RA", Aliases: []string{"Rare Atom", "RA"}},
	{Canonical: "Lynn Vision", Display: "Lynn Vision", Official: "Lynn Vision战队", Colloquial: "LVG", Aliases: []string{"Lynn Vision", "LVG"}},
	{Canonical: "fnatic", Display: "fnatic", Official: "fnatic战队", Colloquial: "橙黑", Aliases: []string{"fnatic", "Fnatic", "橙黑"}},
	{Canonical: "Eternal Fire", Display: "Eternal Fire", Official: "Eternal Fire战队", Colloquial: "永火", Aliases: []string{"Eternal Fire", "永火"}},
	{Canonical: "RED Canids", Display: "RED Canids", Official: "RED Canids战队", Colloquial: "红犬", Aliases: []string{"RED Canids", "红犬"}},
	{Canonical: "3DMAX", Display: "3DMAX", Official: "3DMAX战队", Colloquial: "3DMAX", Aliases: []string{"3DMAX"}},
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

