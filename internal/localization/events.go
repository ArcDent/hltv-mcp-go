package localization

import "strings"

type EventEntry struct {
	Canonical, Official, Colloquial string
	Aliases                          []string
}

var EventCatalog = []EventEntry{
	{Canonical: "IEM Rio", Official: "IEM 里约站", Colloquial: "里约IEM", Aliases: []string{"IEM Rio", "IEM里约", "里约IEM", "里约"}},
	{Canonical: "PGL Astana", Official: "PGL 阿斯塔纳站", Colloquial: "阿斯塔纳PGL", Aliases: []string{"PGL Astana", "PGL阿斯塔纳", "阿斯塔纳PGL"}},
	{Canonical: "BLAST Open Lisbon", Official: "BLAST Open 里斯本站", Colloquial: "里斯本BLAST Open", Aliases: []string{"BLAST Open Lisbon", "BLAST里斯本", "里斯本BLAST"}},
}

var eventLookup = buildEventLookup(EventCatalog)

func buildEventLookup(catalog []EventEntry) map[string]*EventEntry {
	m := make(map[string]*EventEntry)
	for i := range catalog {
		e := &catalog[i]
		all := dedup(append([]string{e.Canonical, e.Official, e.Colloquial}, e.Aliases...))
		for _, a := range all {
			m[strings.ToLower(a)] = e
		}
	}
	return m
}

func LookupEvent(name string) *EventEntry {
	return eventLookup[strings.ToLower(strings.TrimSpace(name))]
}

func FormatEventDisplay(name string) string {
	if e := LookupEvent(name); e != nil {
		parts := dedup([]string{e.Canonical, e.Official})
		if e.Colloquial != "" {
			parts = append(parts, e.Colloquial)
		}
		return strings.Join(parts, "/")
	}
	return name
}

func ExpandEventAliases(name string) []string {
	if e := LookupEvent(name); e != nil {
		return dedup(append([]string{e.Canonical, e.Official, e.Colloquial}, e.Aliases...))
	}
	if name == "" {
		return nil
	}
	return []string{name}
}

func MatchEventName(source, query string) bool {
	sAliases := ExpandEventAliases(source)
	qAliases := ExpandEventAliases(query)
	for _, sa := range sAliases {
		for _, qa := range qAliases {
			if strings.EqualFold(sa, qa) {
				return true
			}
		}
	}
	return false
}
