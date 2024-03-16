/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

import (
	"fmt"
	"regexp"
)

type LogFieldMatcher struct {
	Field string
	Op    string
	Value string

	match func(string) bool
}

func (self *LogFieldMatcher) Match(s string) bool {
	if self.match != nil {
		return self.match(s)
	}
	switch self.Op {
	case "=":
		self.match = func(s string) bool {
			return s == self.Value
		}
	case "=~":
		re, err := regexp.Compile(fmt.Sprintf("^%s$", self.Value))
		if err == nil {
			self.match = func(s string) bool {
				return re.Match([]byte(s))
			}
		}
	}
	// unknown operations or broken regexps never match
	if self.match == nil {
		self.match = func(_ string) bool {
			return false
		}
	}
	return self.match(s)
}

type LogRule struct {
	// Id zero is reserved 'not saved'
	ID int

	// Rule may or may not be disabled
	Disabled bool

	// Is the result interesting, or not?
	Ham bool

	// List of matchers the rule matches against
	Matchers []LogFieldMatcher

	// Comment (if any)
	Comment string

	// Version of the rule; any time the rule is changed, the
	// version is incremented
	Version int
}
