/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"net/http"
	"strconv"

	"github.com/fingon/lixie/cm"
	"github.com/fingon/lixie/data"
)

// This struct represents external configuration - what we can get as query/form parameters
type LogListConfig struct {
	// Global configuration (cookie)
	Global GlobalConfig `json:"-"`

	// Local state (cookie)
	AutoRefresh bool `json:"ar" cm:"auto-refresh"`
	Filter      int  `json:"f" cm:"filter"`

	// These are only handled via links
	BeforeHash uint64
	Expand     uint64
}

// These must match ^ cm tags
const (
	llAutoRefreshKey = "auto-refresh"
	llFilterKey      = "filter"
)

// These are handled only in templates
const (
	expandKey = "expand"
	beforeKey = "before"
)

func (self *LogListConfig) Init(s cm.CookieSource, wr *cm.URLWrapper, w http.ResponseWriter) error {
	// Global config
	err := cm.RunWrapper(s, wr, w, &self.Global)
	if err != nil {
		return err
	}

	// Local config
	err = cm.RunWrapper(s, wr, w, self)
	if err != nil {
		return err
	}

	// Link-based state
	_, err = cm.Uint64FromForm(wr, expandKey, &self.Expand)
	if err != nil {
		return err
	}

	_, err = cm.Uint64FromForm(wr, beforeKey, &self.BeforeHash)
	if err != nil {
		return err
	}

	// before is omitted intentionally
	return nil
}

func (self LogListConfig) WithBeforeHash(v uint64) LogListConfig {
	self.BeforeHash = v
	return self
}

func (self LogListConfig) WithExpand(v uint64) LogListConfig {
	self.Expand = v
	return self
}

func (self LogListConfig) Query() *QueryWrapper {
	base := topLevelLog.Path + "/"

	q := QueryWrapper{Base: base}

	// Cookie-based stuff is handled in Init

	// Link-based things start here
	if self.BeforeHash != 0 {
		q.Add(beforeKey, strconv.FormatUint(self.BeforeHash, 10))
	}
	if self.Expand != 0 {
		q.Add(expandKey, strconv.FormatUint(self.Expand, 10))
	}

	return &q
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
	DisableActions    bool
	DisablePagination bool

	// Is this post request (.. sigh ..)
	Post bool

	// Convenience results from filter()
	EnableAccurateCounting bool
	FilteredCount          int
	TotalCount             int

	// For filtering
	Limit int
}

func (self *LogListModel) LogToRule(log *data.Log) *data.LogRule {
	if self.LogRules != nil {
		return data.LogToRule(log, self.LogRules)
	}
	return log.ToRule(&self.DB.LogRules)
}

func (self *LogListModel) LogVerdict(log *data.Log) int {
	rule := self.LogToRule(log)
	return data.LogRuleToVerdict(rule)
}

func (self *LogListModel) Filter() error {
	// Some spare capacity but who really cares
	logs := make([]*data.Log, 0, self.Limit)
	active := self.Config.BeforeHash == 0
	count := 0
	allLogs, err := self.DB.Logs()
	if err != nil {
		return err
	}

	allLogs = filterFTS(allLogs, self.Config.Global.Search, len(allLogs))
	self.TotalCount = len(allLogs)
	for _, log := range allLogs {
		if !active {
			if log.Hash() == self.Config.BeforeHash {
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
	return nil
}

func logListHandler(st State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := LogListConfig{Filter: data.LogVerdictSpam}
		wr, err := cm.GetWrapper(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = config.Init(r, wr, w)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		model := LogListModel{Config: config, DB: st.DB, Limit: 20, Post: r.Method == "POST"}
		err = model.Filter()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = LogList(st, model).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	})
}
