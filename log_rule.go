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
	"github.com/fingon/lixie/cm"
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

func NewLogRuleFromForm(r cm.FormValued) (result *data.LogRule, err error) {
	rule := data.LogRule{}

	if _, err = cm.IntFromForm(r, idKey, &rule.ID); err != nil {
		return
	}
	if _, err = cm.IntFromForm(r, versionKey, &rule.Version); err != nil {
		return
	}
	cm.BoolFromForm(r, disabledKey, &rule.Disabled)
	cm.BoolFromForm(r, hamKey, &rule.Ham)
	rule.Comment = r.FormValue(commentKey)

	// Read the matcher fields
	i := 0
	del := -1
	for {
		// Keep track of what to delete here too
		if r.FormValue(fieldID(i, deleteField)) != "" {
			del = i
		}
		field := r.FormValue(fieldID(i, fieldField))
		op := r.FormValue(fieldID(i, opField))
		value := r.FormValue(fieldID(i, valueField))
		if field == "" && op == "" && value == "" {
			break
		}
		matcher := data.LogFieldMatcher{Field: field, Op: op, Value: value}
		rule.Matchers = append(rule.Matchers, matcher)
		i++
	}

	// Handle the mutation actions
	if del >= 0 {
		rule.Matchers = slices.Delete(rule.Matchers, del, del+1)
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
		DB:                     db,
		LogRules:               []*data.LogRule{rule},
		DisableActions:         true,
		DisablePagination:      true,
		EnableAccurateCounting: true,
		Limit:                  5}
	m.Filter()
	return &m
}

func findMatchingOtherRules(db *data.Database, logs []*data.Log, skipRule *data.LogRule) *LogRuleListModel {
	// We could also check for strict supersets (but regexp+regexp
	// matching is tricky). So we just show rules that out of the
	// box seem to overlap as they match the same rules.
	rules := []*data.LogRule{}

	for _, rule := range db.LogRules {
		if rule == skipRule {
			continue
		}
		for _, log := range logs {
			if data.LogMatchesRule(log, rule) {
				rules = append(rules, rule)
				break
			}
		}
	}
	if len(rules) > 0 {
		return &LogRuleListModel{LogRules: rules}
	}
	return nil
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
				if v.ID != rule.ID {
					continue
				}
				// TODO do we want to error if version differs?
				if v.Version == rule.Version {
					rule.Version++
					db.LogRules[i] = rule
					err = db.Save()
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
				} else {
					fmt.Printf("Version mismatch - %d <> %d\n", v.Version, rule.Version)
				}
				http.Redirect(w, r, topLevelLogRule.Path, http.StatusSeeOther)
				return
			}

			// Not found. Add new one.
			fmt.Printf("Adding new rule\n")
			err = db.Add(*rule)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			http.Redirect(w, r, topLevelLogRule.Path, http.StatusSeeOther)
			return
		}
		logs := findMatchingLogs(db, rule)
		rules := findMatchingOtherRules(db, logs.Logs, rule)
		err = LogRuleEdit(*rule, rules, logs).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
}

func logRuleDeleteSpecificHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ridString := r.PathValue("id")
		rid, err := strconv.Atoi(ridString)
		if err != nil {
			// TODO handle error
			return
		}
		if err = db.Delete(rid); err != nil {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, topLevelLogRule.Path, http.StatusSeeOther)
	})
}

func logRuleEditSpecificHandler(db *data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ridString := r.PathValue("id")
		rid, err := strconv.Atoi(ridString)
		if err != nil {
			// TODO handle error
			return
		}
		for _, rule := range db.LogRules {
			if rule.ID == rid {
				logs := findMatchingLogs(db, rule)
				rules := findMatchingOtherRules(db, logs.Logs, rule)
				err = LogRuleEdit(*rule, rules, logs).Render(r.Context(), w)
				if err != nil {
					http.Error(w, err.Error(), 500)
				}
				return
			}
		}
		http.NotFound(w, r)
	})
}

func fieldID(id int, suffix string) string {
	return fmt.Sprintf("row-%d-%s", id, suffix)
}

func ruleTitle(rule data.LogRule) string {
	if rule.ID > 0 {
		return fmt.Sprintf("Log rule editor - editing #%d", rule.ID)
	}
	return "Log rule creator"
}

func ruleIDString(rule data.LogRule) string {
	return strconv.Itoa(rule.ID)
}

func ruleLinkString(id int, op string) string {
	return fmt.Sprintf("%s/%d/%s", topLevelLogRule.Path, id, op)
}

func ruleLink(id int, op string) templ.SafeURL {
	return templ.URL(ruleLinkString(id, op))
}
