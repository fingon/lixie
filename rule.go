/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"net/http"
	"slices"
	"strconv"
)

type LogFieldMatcher struct {
	Field string
	Op    string
	Value string
}

type LogRule struct {
	// Id zero is reserved 'not saved'
	Id int

	// Rule may or may not be disabled
	Disabled bool

	// Is the result interesting, or not?
	Ham bool

	// List of matchers the rule matches against
	Matchers []LogFieldMatcher

	// Version of the rule; any time the rule is changed, the
	// version is incremented
	Version int
}

func intFromForm(r *http.Request, key string, value *int) (found bool, err error) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	*value, err = strconv.Atoi(raw)
	return
}

func boolFromForm(r *http.Request, key string, value *bool) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	*value = true
}

func NewLogRuleFromForm(r *http.Request) (result *LogRule, err error) {
	rule := LogRule{}

	if _, err = intFromForm(r, idKey, &rule.Id); err != nil {
		return
	}
	if _, err = intFromForm(r, versionKey, &rule.Version); err != nil {
		return
	}
	boolFromForm(r, disabledKey, &rule.Disabled)
	boolFromForm(r, hamKey, &rule.Ham)

	// Read the matcher fields
	i := 0
	delete := -1
	for {
		field := fieldId(i, fieldField)
		_, ok := r.Form[field]
		if !ok {
			break
		}
		// Keep track of what to delete here too
		if r.FormValue(fieldId(i, deleteField)) != "" {
			delete = i
		}
		op := fieldId(i, opField)
		value := fieldId(i, valueField)
		matcher := LogFieldMatcher{Field: r.FormValue(field), Op: r.FormValue(op), Value: r.FormValue(value)}
		rule.Matchers = append(rule.Matchers, matcher)
		i += 1
	}

	// Handle the mutation actions
	if delete >= 0 {
		rule.Matchers = slices.Delete(rule.Matchers, delete, delete+1)
	}
	if r.FormValue(actionAdd) != "" {
		rule.Matchers = append(rule.Matchers, LogFieldMatcher{})
	}
	// Save is dealt with externally
	result = &rule
	return
}
