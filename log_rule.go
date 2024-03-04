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

	"github.com/a-h/templ"
	"github.com/fingon/lixie/data"
)

// top-level form fields

const idKey = "rid"

const commentKey = "com"
const disabledKey = "d"
const hamKey = "h"
const versionKey = "ver"

// per-matcher fields
const deleteField = "del"
const fieldField = "f"
const opField = "o"
const valueField = "v"

// actions
const actionAdd = "a"
const actionSave = "s"

func NewLogRuleFromForm(r FormValued) (result *data.LogRule, err error) {
	rule := data.LogRule{}

	if _, err = intFromForm(r, idKey, &rule.Id); err != nil {
		return
	}
	if _, err = intFromForm(r, versionKey, &rule.Version); err != nil {
		return
	}
	boolFromForm(r, disabledKey, &rule.Disabled)
	boolFromForm(r, hamKey, &rule.Ham)
	rule.Comment = r.FormValue(commentKey)

	// Read the matcher fields
	i := 0
	delete := -1
	for {
		// Keep track of what to delete here too
		if r.FormValue(fieldId(i, deleteField)) != "" {
			delete = i
		}
		field := r.FormValue(fieldId(i, fieldField))
		op := r.FormValue(fieldId(i, opField))
		value := r.FormValue(fieldId(i, valueField))
		if field == "" && op == "" && value == "" {
			break
		}
		matcher := data.LogFieldMatcher{Field: field, Op: op, Value: value}
		rule.Matchers = append(rule.Matchers, matcher)
		i += 1
	}

	// Handle the mutation actions
	if delete >= 0 {
		rule.Matchers = slices.Delete(rule.Matchers, delete, delete+1)
	}
	if r.FormValue(actionAdd) != "" {
		rule.Matchers = append(rule.Matchers, data.LogFieldMatcher{Op: "="})
	}
	// Save is dealt with externally
	result = &rule
	return
}

func findMatchingLogs(db *data.Database, rule *data.LogRule) *LogListModel {
	m := LogListModel{
		Config:                 LogListConfig{Filter: data.LogVerdictUnknown},
		Logs:                   db.Logs(),
		LogRules:               []*data.LogRule{rule},
		DisableActions:         true,
		DisablePagination:      true,
		EnableAccurateCounting: true,
		Limit:                  5}
	m.Filter()
	return &m
}

func logRuleEditHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rule, err := NewLogRuleFromForm(r)
		if err != nil {
			// TODO log error?
			return
		}
		if r.FormValue(actionSave) != "" {
			// Look for existing rule first
			for i, v := range db.LogRules {
				if v.Id == rule.Id {
					// TODO do we want to error if version differs?
					if v.Version == rule.Version {
						rule.Version++
						db.LogRules[i] = rule
						db.Save()
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
		matches := findMatchingLogs(db, rule)
		LogRuleEdit(*rule, matches).Render(r.Context(), w)
	})
}

func logRuleDeleteSpecificHandler(db *data.Database) http.Handler {
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

func logRuleEditSpecificHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid_string := r.PathValue("id")
		rid, err := strconv.Atoi(rid_string)
		if err != nil {
			// TODO handle error
			return
		}
		for _, rule := range db.LogRules {
			if rule.Id == rid {
				matches := findMatchingLogs(db, rule)
				LogRuleEdit(*rule, matches).Render(r.Context(), w)
				return
			}
		}
		http.NotFound(w, r)
	})
}

func fieldId(id int, suffix string) string {
	return fmt.Sprintf("row-%d-%s", id, suffix)
}

func ruleTitle(rule data.LogRule) string {
	if rule.Id > 0 {
		return fmt.Sprintf("Log rule editor - editing #%d", rule.Id)
	}
	return "Log rule creator"
}

func ruleIdString(rule data.LogRule) string {
	return fmt.Sprintf("%d", rule.Id)
}

func ruleLink(id int, op string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/rule/%d/%s", id, op))
}
