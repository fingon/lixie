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
		for _, rule := range self.LogRules[self.Index:last] {
			rules = append(rules, rule)

		}
	}
	self.LogRules = rules
}

func (self *LogRuleListModel) NextLinkString() string {
	return topLevelLogRule.Path + fmt.Sprintf("/?%s=%d", indexKey, self.Index+self.Limit)
}

const indexKey = "i"

func NewLogRuleListModel(r *http.Request, db *data.Database) *LogRuleListModel {
	rules := db.LogRulesReversed()
	m := LogRuleListModel{DB: db, LogRules: rules, Limit: 10}
	intFromForm(r, indexKey, &m.Index)
	return &m
}

func logRuleListHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewLogRuleListModel(r, db)
		m.Filter()
		LogRuleList(*m).Render(r.Context(), w)
	})
}
