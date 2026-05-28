package localization

import "strings"

type EventEntry struct {
	Canonical, Official, Colloquial string
	Aliases                          []string
}

var EventCatalog = []EventEntry{}

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

