/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

const (
	LogVerdictUnknown int = iota
	LogVerdictHam
	LogVerdictSpam
	LogVerdictNothing
	NumLogVerdicts
)

func LogVerdictToString(verdict int) string {
	switch verdict {
	case LogVerdictHam:
		return "Ham"
	case LogVerdictSpam:
		return "Spam"
	case LogVerdictNothing:
		return "Nothing"
	}
	return "Unknown"
}

func LogMatchesRule(log *Log, rule *LogRule) bool {
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

func LogToRule(log *Log, rules []*LogRule) *LogRule {
	for _, rule := range rules {
		if LogMatchesRule(log, rule) {
			return rule
		}
	}
	return nil
}

func LogRuleToVerdict(rule *LogRule) int {
	if rule != nil {
		if rule.Ham {
			return LogVerdictHam
		}
		return LogVerdictSpam
	}
	return LogVerdictUnknown
}

func LogVerdict(log *Log, rules []*LogRule) int {
	rule := LogToRule(log, rules)
	return LogRuleToVerdict(rule)
}
