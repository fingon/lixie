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

func LogVerdict(log *Log, rules []*LogRule) int {
	verdict := LogVerdictUnknown
	for _, rule := range rules {
		if rule.Disabled {
			continue
		}
		match := true
		for _, matcher := range rule.Matchers {
			value, ok := log.Stream[matcher.Field]
			if !ok {
				value, ok = log.Fields[matcher.Field].(string)
				if !ok {
					if matcher.Field == "message" {
						value = log.Message
					} else {
						match = false
						break
					}
				}
			}

			// Only supported op is '='
			if matcher.Op == "=" {
				if value == matcher.Value {
					continue
				}
			}
			match = false
			break
		}
		if match {
			if rule.Ham {
				verdict = LogVerdictHam
			} else {
				verdict = LogVerdictSpam
			}
		}
	}
	return verdict
}
