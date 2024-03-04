/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/a-h/templ"
	"github.com/fingon/lixie/data"
)

type LogListConfig struct {
	// parse form + send query
	AutoRefresh bool
	Expand      uint64
	Filter      int

	// send query-only
	Before uint64
}

const autoRefreshKey = "ar"
const expandKey = "exp"
const beforeKey = "b"
const filterKey = "f"

func NewLogListConfig(r FormValued) LogListConfig {
	autorefresh := r.FormValue(autoRefreshKey) != ""
	expand := uint64(0)
	uint64FromForm(r, expandKey, &expand)
	filter := data.LogVerdictSpam
	intFromForm(r, filterKey, &filter)
	// before is omitted intentionally
	return LogListConfig{AutoRefresh: autorefresh, Expand: expand, Filter: filter}
}

func (self LogListConfig) WithAutoRefresh(v bool) LogListConfig {
	self.AutoRefresh = v
	return self
}

func (self LogListConfig) WithBefore(v uint64) LogListConfig {
	self.Before = v
	return self
}

func (self LogListConfig) WithExpand(v uint64) LogListConfig {
	self.Expand = v
	return self
}

func (self LogListConfig) WithFilter(v int) LogListConfig {
	self.Filter = v
	return self
}

func (self LogListConfig) ToLinkString() string {
	base := topLevelLog.Path
	v := url.Values{}
	// TODO: Better diffs-from-default handling
	if self.AutoRefresh {
		v.Set(autoRefreshKey, "1")
	}
	if self.Before != 0 {
		v.Set(beforeKey, strconv.FormatUint(self.Before, 10))
	}
	if self.Expand != 0 {
		v.Set(expandKey, strconv.FormatUint(self.Expand, 10))
	}
	if self.Filter != data.LogVerdictSpam {
		v.Set(filterKey, strconv.Itoa(self.Filter))
	}
	if len(v) == 0 {
		return base
	}
	return base + "?" + v.Encode()
}

func (self LogListConfig) ToLink() templ.SafeURL {
	return templ.URL(self.ToLinkString())
}

type LogListModel struct {
	Config   LogListConfig
	Logs     []*data.Log
	LogRules []*data.LogRule

	ExcludeVerdict int

	// Actions
	DisableActions bool

	// Paging support
	BeforeHash        uint64
	DisablePagination bool
	Limit             int

	// Convenience results from filter()
	EnableAccurateCounting bool
	FilteredCount          int
	TotalCount             int
}

// TODO: These should be probably cached
func (self *LogListModel) LogVerdictRule(log *data.Log) *data.LogRule {
	return data.LogVerdictRule(log, self.LogRules)
}
func (self *LogListModel) LogVerdict(log *data.Log) int {
	return data.LogVerdict(log, self.LogRules)
}

func (self *LogListModel) Filter() {
	// Some spare capacity but who really cares
	logs := make([]*data.Log, 0, self.Limit)
	active := self.BeforeHash == 0
	count := 0
	self.TotalCount = len(self.Logs)
	for _, log := range self.Logs {
		if !active {
			if log.Hash() == self.BeforeHash {
				active = true
			}
			if self.EnableAccurateCounting && self.LogVerdict(log) != self.Config.Filter {
				count++
			}
			continue
		}
		if self.LogVerdict(log) == self.Config.Filter {
			continue
		}
		count++
		if len(logs) < self.Limit {
			logs = append(logs, log)
		} else if !self.EnableAccurateCounting {
			break
		}
	}
	self.Logs = logs
	self.FilteredCount = count
}

func logListHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var before_hash uint64
		uint64FromForm(r, beforeKey, &before_hash)
		logs := db.Logs()
		model := LogListModel{Config: NewLogListConfig(r),
			BeforeHash: before_hash,
			Limit:      20,
			Logs:       logs,
			LogRules:   db.LogRulesReversed()}
		model.Filter()
		LogList(model).Render(r.Context(), w)
	})
}
