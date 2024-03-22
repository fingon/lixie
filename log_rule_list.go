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

	"github.com/fingon/lixie/data"
)

// This struct represents external configuration - what we can get as query/form parameters
type LogRuleListConfig struct {
	Search string

	// Paging support
	Index int
}

func (self *LogRuleListConfig) Init(r *http.Request) (err error) {
	_, err = intFromForm(r, indexKey, &self.Index)
	if err != nil {
		return
	}

	self.Search = r.FormValue(searchKey)
	return
}

func (self *LogRuleListConfig) ToLinkString() string {
	base := topLevelLogRule.Path + "/"

	v := url.Values{}
	if self.Index != 0 {
		v.Set(indexKey, strconv.Itoa(self.Index))
	}
	if self.Search != "" {
		v.Set(searchKey, self.Search)
	}
	if len(v) == 0 {
		return base
	}
	return base + "?" + v.Encode()
}

type LogRuleListModel struct {
	Config LogRuleListConfig

	DB *data.Database

	// This is the list actually rendered to the client in this
	// request; subset of DB rules
	LogRules []*data.LogRule

	HasMore bool
	Limit   int
}

func (self *LogRuleListModel) Filter() {
	// First do fts (if necessary)
	last := self.Config.Index + self.Limit
	rules := filterFTS(self.LogRules, self.Config.Search, last+1)
	self.HasMore = len(rules) >= last
	if last >= len(rules) {
		last = len(rules) - 1
	}

	result := make([]*data.LogRule, 0, self.Limit)
	if last > self.Config.Index {
		result = append(result, rules[self.Config.Index:last]...)
	}
	self.LogRules = result
}

func (self *LogRuleListModel) NextLinkString() string {
	next := self.Config
	next.Index += self.Limit
	return next.ToLinkString()
}

const indexKey = "i"

func logRuleListHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := LogRuleListConfig{}
		err := config.Init(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		rules := db.LogRulesReversed()
		m := LogRuleListModel{Config: config, DB: db, LogRules: rules, Limit: 10}
		m.Filter()
		err = LogRuleList(m).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
}
