/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

const (
	LogVerdictUnknown int = iota
	LogVerdictHam
	LogVerdictSpam
)

func logMatchesRule(log *Log, rule *LogRule) bool {
	if rule.Disabled {
		return false
	}
	for _, matcher := range rule.Matchers {
		if matcher.Field == "" && matcher.Value == "" {
			continue
		}
		value, ok := log.Stream[matcher.Field]
		if !ok {
			value, ok = log.Fields[matcher.Field].(string)
			if !ok {
				if matcher.Field != "message" {
					return false
				}
				value = log.Message
			}
		}
		if !matcher.Match(value) {
			return false
		}
	}
	return true

}

func LogVerdictRule(log *Log, rules []*LogRule) *LogRule {
	for _, rule := range rules {
		if logMatchesRule(log, rule) {
			return rule
		}
	}
	return nil
}

func LogVerdict(log *Log, rules []*LogRule) int {
	rule := LogVerdictRule(log, rules)
	if rule != nil {
		if rule.Ham {
			return LogVerdictHam
		}
		return LogVerdictSpam
	}
	return LogVerdictUnknown
}
