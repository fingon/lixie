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

func (self *LogListConfig) Init(r FormValued) error {
	self.AutoRefresh = r.FormValue(autoRefreshKey) != ""
	_, err := uint64FromForm(r, expandKey, &self.Expand)
	if err != nil {
		return err
	}

	self.Filter = data.LogVerdictSpam
	_, err = intFromForm(r, filterKey, &self.Filter)
	if err != nil {
		return err
	}

	// before is omitted intentionally
	return nil
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
	Config LogListConfig

	DB *data.Database

	// This is subset of the database actually shown to the user
	Logs []*data.Log

	// This can be used if specifying custom set of rules, it defaults to DB.LogRules
	LogRules []*data.LogRule

	// Filtering criteria
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

func (self *LogListModel) LogToRule(log *data.Log) *data.LogRule {
	if self.LogRules != nil {
		return data.LogToRule(log, self.LogRules)
	}
	return self.DB.LogToRule(log)
}
func (self *LogListModel) LogVerdict(log *data.Log) int {
	rule := self.LogToRule(log)
	return data.LogRuleToVerdict(rule)
}

func (self *LogListModel) Filter() {
	// Some spare capacity but who really cares
	logs := make([]*data.Log, 0, self.Limit)
	active := self.BeforeHash == 0
	count := 0
	all_logs := self.DB.Logs()
	self.TotalCount = len(all_logs)
	for _, log := range all_logs {
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
		_, err := uint64FromForm(r, beforeKey, &before_hash)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		config := LogListConfig{}
		err = config.Init(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		model := LogListModel{Config: config,
			BeforeHash: before_hash,
			Limit:      20,
			DB:         db}
		model.Filter()
		err = LogList(model).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	})
}
