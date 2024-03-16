/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"fmt"
	"net/http"

	"github.com/fingon/lixie/data"
)

type LogRuleListModel struct {
	DB *data.Database

	// This is the list actually rendered to the client in this
	// request; subset of DB rules
	LogRules []*data.LogRule

	// Paging support
	HasMore bool
	Index   int
	Limit   int
}

func (self *LogRuleListModel) Filter() {
	last := self.Index + self.Limit
	self.HasMore = len(self.LogRules) >= last
	if last >= len(self.LogRules) {
		last = len(self.LogRules) - 1
	}
	rules := make([]*data.LogRule, 0, self.Limit)
	if last > self.Index {
		rules = append(rules, self.LogRules[self.Index:last]...)
	}
	self.LogRules = rules
}

func (self *LogRuleListModel) NextLinkString() string {
	return topLevelLogRule.Path + fmt.Sprintf("/?%s=%d", indexKey, self.Index+self.Limit)
}

const indexKey = "i"

func (self *LogRuleListModel) Init(r *http.Request) (err error) {
	_, err = intFromForm(r, indexKey, &self.Index)
	return
}

func logRuleListHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rules := db.LogRulesReversed()
		m := LogRuleListModel{DB: db, LogRules: rules, Limit: 10}
		err := m.Init(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		m.Filter()
		err = LogRuleList(m).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
}
