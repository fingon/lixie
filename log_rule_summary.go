/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Jun  7 16:43:58 2024 mstenber
 * Last modified: Fri Jun  7 17:26:30 2024 mstenber
 * Edit time:     24 min
 *
 */

package main

import (
	"fmt"
	"sort"

	"github.com/a-h/templ"
)

type rsSortHeader struct {
	Title       string
	CompareLess func(s1, s2 *LogSourceSummary) bool
	id          int
}

func (self *rsSortHeader) ID() int {
	if self.id == 0 {
		// Initialize our link dynamically. Not super efficient but it does not really matter.
		for i, e := range rsSortRules {
			if e == self {
				self.id = i + 1
				return self.id
			}
		}
	}
	return self.id
}

func (self *rsSortHeader) ActionLink(base string, st int) templ.SafeURL {
	id := self.ID()
	if st == id {
		id = -id
	}
	return templ.URL(fmt.Sprintf("%s?rss=%d", base, id))
}

var rsSource = &rsSortHeader{Title: "Source",
	CompareLess: func(s1, s2 *LogSourceSummary) bool {
		return s1.Source < s2.Source
	}}

var rsRules = &rsSortHeader{Title: "Rules",
	CompareLess: func(s1, s2 *LogSourceSummary) bool {
		return s1.RuleCount < s2.RuleCount
	}}

var rsHits = &rsSortHeader{Title: "Hits",
	CompareLess: func(s1, s2 *LogSourceSummary) bool {
		return s1.Hits < s2.Hits
	}}

var rsHitsPerRule = &rsSortHeader{Title: "Hits per rule",
	CompareLess: func(s1, s2 *LogSourceSummary) bool {
		hpr1 := s1.Hits / s1.RuleCount
		hpr2 := s2.Hits / s2.RuleCount
		return hpr1 < hpr2
	}}

var rsSortRules = []*rsSortHeader{rsSource, rsRules, rsHits, rsHitsPerRule}

func rsSortedRules(sli []*LogSourceSummary, order int) []*LogSourceSummary {
	sh := rsRules
	asc := false

	for _, rule := range rsSortRules {
		rid := rule.ID()
		if order == rid {
			sh = rule
			asc = false
			break
		}
		if order == -rid {
			sh = rule
			asc = true
			break
		}
	}

	sort.Slice(sli, func(i, j int) bool {
		return sh.CompareLess(sli[i], sli[j]) == asc
	})
	return sli
}
