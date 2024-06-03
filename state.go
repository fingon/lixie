/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Jun  2 19:57:04 2024 mstenber
 * Last modified: Mon Jun  3 08:03:36 2024 mstenber
 * Edit time:     26 min
 *
 */

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/fingon/lixie/data"
	"golang.org/x/exp/maps"
)

type State struct {
	// How often we retry page load
	RefreshIntervalMs int

	// Current build version
	BuildTimestamp string

	DB *data.Database
}

type LogSourceSummary struct {
	Source    string
	RuleCount int
	Hits      int
}

func (self *LogSourceSummary) SearchLink() templ.SafeURL {
	search := self.Source
	search, _ = strings.CutPrefix(search, "=~")
	search, _ = strings.CutPrefix(search, "=")
	return templ.URL(fmt.Sprintf("%s?%s=%s",
		topLevelLogRule.Path,
		globalSearchKey,
		search))
}

func (self *State) RuleStats(topk int) []*LogSourceSummary {
	temp := make(map[string]*LogSourceSummary)
	for _, rule := range self.DB.LogRules.Rules {
		ss := rule.SourceString()
		entry, ok := temp[ss]
		if !ok {
			entry = &LogSourceSummary{Source: ss}
			temp[ss] = entry
		}
		entry.RuleCount++
		entry.Hits += self.DB.RuleCount(rule.ID)
	}
	stats := maps.Values(temp)
	sort.Slice(stats, func(i, j int) bool {
		r1 := stats[i]
		r2 := stats[j]
		if r1.RuleCount < r2.RuleCount {
			return false
		}
		if r1.RuleCount == r2.RuleCount && r1.Hits < r2.Hits {
			return false
		}
		return true
	})
	if len(stats) > topk {
		return stats[:topk]
	}
	return stats
}
