/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import "slices"

type Database struct {
	// TODO: This should be probably really a map
	LogRules         []*LogRule
	logRulesReversed []*LogRule
	nextId           int
}

func (self *Database) LogRulesReversed() []*LogRule {
	if self.logRulesReversed == nil {
		count := len(self.LogRules)
		reversed := make([]*LogRule, count)
		for k, v := range self.LogRules {
			reversed[count-k-1] = v
		}
		self.logRulesReversed = reversed
	}
	return self.logRulesReversed
}

func (self *Database) Add(r LogRule) {
	r.Id = self.nextLogRuleId()
	self.LogRules = append(self.LogRules, &r)
	self.logRulesReversed = nil
}

func (self *Database) Delete(rid int) bool {
	for i, v := range self.LogRules {
		if v.Id == rid {
			self.LogRules = slices.Delete(self.LogRules, i, i+1)
			self.logRulesReversed = nil
			return true
		}
	}
	return false
}

func (self *Database) nextLogRuleId() int {
	id := self.nextId
	if id == 0 {
		id = 1 // Start at 1 even with empty database
		for _, v := range self.LogRules {
			if v.Id >= id {
				id = v.Id + 1
			}
		}
	}
	self.nextId = id + 1
	return id
}
