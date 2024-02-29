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

	// Comment (if any)
	Comment string

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
		rule.Matchers = append(rule.Matchers, LogFieldMatcher{Op: "="})
	}
	// Save is dealt with externally
	result = &rule
	return
}

func logRuleEditHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rule, err := NewLogRuleFromForm(r)
		if err != nil {
			// TODO log error?
			return
		}
		if r.FormValue(actionSave) != "" {
			// Look for existing rule first
			for _, v := range db.LogRules {
				if v.Id == rule.Id {
					// TODO do we want to error if version differs?
					if v.Version == rule.Version {
						*v = *rule
						v.Version++
					} else {
						fmt.Printf("Version mismatch - %d <> %d\n", v.Version, rule.Version)
					}
					http.Redirect(w, r, "/rule/", http.StatusSeeOther)
					return
				}
			}

			// Not found. Add new one.
			fmt.Printf("Adding new rule\n")
			db.Add(*rule)
			http.Redirect(w, r, "/rule/", http.StatusSeeOther)
			return
		}
		LogRuleEdit(*rule).Render(r.Context(), w)
	})
}

func logRuleDeleteSpecificHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid_string := r.PathValue("id")
		rid, err := strconv.Atoi(rid_string)
		if err != nil {
			// TODO handle error
			return
		}
		if db.Delete(rid) {
			http.Redirect(w, r, "/rule/", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
}

func logRuleEditSpecificHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid_string := r.PathValue("id")
		rid, err := strconv.Atoi(rid_string)
		if err != nil {
			// TODO handle error
			return
		}
		for _, v := range db.LogRules {
			if v.Id == rid {
				LogRuleEdit(*v).Render(r.Context(), w)
				return
			}
		}
		http.NotFound(w, r)
	})
}

func logRuleListHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogRuleList(db.LogRulesReversed()).Render(r.Context(), w)

	})
}

const idKey = "rid"
const versionKey = "ver"

const hamKey = "h"
const disabledKey = "d"

const deleteField = "del"
const fieldField = "f"
const opField = "o"
const valueField = "v"

const actionAdd = "a"
const actionSave = "s"

func fieldId(id int, suffix string) string {
	return fmt.Sprintf("row-%d-%s", id, suffix)
}

func ruleTitle(rule LogRule) string {
	if rule.Id > 0 {
		return fmt.Sprintf("Log rule editor - editing #%d", rule.Id)
	}
	return "Log rule creator"
}

func ruleIdString(rule LogRule) string {
	return fmt.Sprintf("%d", rule.Id)
}
