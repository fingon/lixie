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

	// List of matchers the rule matches against
	Matchers []LogFieldMatcher
}

func NewLogRuleFromForm(r *http.Request) (*LogRule, error) {
	matchers := []LogFieldMatcher{}
	rid := 0

	// Read ID of the Rule (if any)
	id_string := r.FormValue(idKey)
	if id_string != "" {
		var err error
		rid, err = strconv.Atoi(id_string)
		if err != nil {
			return nil, err
		}
	}

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
		matchers = append(matchers, matcher)
		i += 1
	}

	// Handle the mutation actions
	if delete >= 0 {
		matchers = slices.Delete(matchers, delete, delete+1)
	}
	if r.FormValue(actionAdd) != "" {
		matchers = append(matchers, LogFieldMatcher{})
	}
	// Save is dealt with externally
	return &LogRule{Id: rid, Matchers: matchers}, nil
}
