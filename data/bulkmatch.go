/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

/*
 Bulk rule matcher.

 It first analyzer the rules, and finds the field with the most exact
matches. Then, using that field, it sequence of match objects which
either do single match (slow), or map-based lookup + list based match
(fast) for the given key.
*/

package data

import "log/slog"

type logMatcher interface {
	ToRule(_ string, log *Log) *LogRule
}

type singleLogMatcher struct {
	rule *LogRule
}

func (self *singleLogMatcher) ToRule(_ string, log *Log) *LogRule {
	if LogMatchesRule(log, self.rule) {
		return self.rule
	}
	return nil
}

type mapLogMatcher struct {
	value2Rules map[string][]*LogRule
}

func (self *mapLogMatcher) ToRule(value string, log *Log) *LogRule {
	rules, ok := self.value2Rules[value]
	if !ok {
		return nil
	}
	for _, rule := range rules {
		if LogMatchesRule(log, rule) {
			return rule
		}
	}
	return nil
}

type BulkRuleMatcher struct {
	field    string
	matchers []logMatcher
}

func (self *BulkRuleMatcher) ToRule(log *Log) *LogRule {
	value, ok := log.Stream[self.field]
	if !ok {
		value, ok = log.Fields[self.field].(string)
		if !ok {
			value = ""
		}
	}
	for _, matcher := range self.matchers {
		rule := matcher.ToRule(value, log)
		if rule != nil {
			return rule
		}
	}
	return nil
}

func NewBulkRuleMatcher(rules []*LogRule) *BulkRuleMatcher {
	fieldToEMcount := map[string]int{}
	for _, rule := range rules {
		for _, matcher := range rule.Matchers {
			if matcher.Op != "=" {
				continue
			}
			cnt := fieldToEMcount[matcher.Field]
			cnt++
			fieldToEMcount[matcher.Field] = cnt
		}
	}
	// TODO: While this is semi ok, what if all values are same?
	// it isn't as useful for exact matching..
	bestField := ""
	bestCount := 0
	for field, count := range fieldToEMcount {
		if count > bestCount {
			bestField = field
			bestCount = count
		}
	}
	slog.Debug("NewBulkRuleMatcher chose", "field", bestField, "count", bestCount)
	brm := BulkRuleMatcher{field: bestField}
	var currentMatcher *mapLogMatcher
	var slowRules, fastRules int
	for _, rule := range rules {
		found := false
		if bestField != "" {
			for _, matcher := range rule.Matchers {
				if matcher.Op == "=" && matcher.Field == bestField {
					if currentMatcher == nil {
						currentMatcher = &mapLogMatcher{make(map[string][]*LogRule)}
						brm.matchers = append(brm.matchers, currentMatcher)
					}
					rules := currentMatcher.value2Rules[matcher.Value]
					rules = append(rules, rule)
					currentMatcher.value2Rules[matcher.Value] = rules
					found = true
					fastRules++
					break
				}
			}
		}
		if !found {
			currentMatcher = nil
			// Use fallback code here; unfortunate, but it is what it is
			brm.matchers = append(brm.matchers, &singleLogMatcher{rule})
			slowRules++
		}
	}
	slog.Debug("Produced matcher", "fast", fastRules, "slow", slowRules)
	return &brm
}
